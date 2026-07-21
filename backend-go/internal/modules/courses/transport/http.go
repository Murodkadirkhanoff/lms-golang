// Package transport is the courses context's HTTP adapter.
package transport

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	enrollmentcontract "github.com/chashma/lms/internal/modules/enrollment/contract"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Handler serves the courses context endpoints.
type Handler struct {
	svc  *application.Service
	gate enrollmentcontract.EnrollmentGate
}

// NewHandler builds the courses HTTP handler. gate (enrollment) is used only
// for review gating and the paywall.
func NewHandler(svc *application.Service, gate enrollmentcontract.EnrollmentGate) *Handler {
	return &Handler{svc: svc, gate: gate}
}

func (h *Handler) writeCourseErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		web.NotFound(w)
	case errors.Is(err, domain.ErrEditConflict):
		web.EditConflict(w)
	default:
		web.ServerError(w, r, err)
	}
}

// --- courses ---

func (h *Handler) listCourses(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := web.ParamInt(q.Get("page"), 1)
	pageSize := web.ParamInt(q.Get("pageSize"), 8)
	instructorID := web.ParamInt64(q.Get("instructorId"), 0)
	sort := q.Get("sort")
	if sort == "" {
		sort = "popular"
	}

	v := web.NewValidator()
	v.Check(page > 0, "page", "must be greater than zero")
	v.Check(pageSize > 0 && pageSize <= 100, "pageSize", "must be between 1 and 100")
	v.Check(web.Permitted(sort, "popular", "newest", "price-asc", "price-desc"),
		"sort", "must be one of popular, newest, price-asc, price-desc")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	list, total, err := h.svc.List(r.Context(), application.CourseFilters{
		Search:             q.Get("search"),
		CategorySlug:       q.Get("category"),
		Sort:               sort,
		Page:               page,
		PageSize:           pageSize,
		IDs:                web.ParseIDList(q.Get("ids")),
		InstructorID:       instructorID,
		IncludeUnpublished: instructorID != 0,
	})
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"items": list, "page": page, "pageSize": pageSize, "total": total,
	}, nil)
}

func (h *Handler) showCourse(w http.ResponseWriter, r *http.Request) {
	course, err := h.svc.GetCourse(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	reviews, err := h.svc.ReviewsForCourse(r.Context(), course.ID, 20)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	course.Reviews = reviews
	if err := h.decorateOne(r.Context(), course); err != nil {
		web.ServerError(w, r, err)
		return
	}

	h.sanitize(r.Context(), course)
	web.WriteJSON(w, http.StatusOK, web.Envelope{"course": course}, nil)
}

func (h *Handler) createCourse(w http.ResponseWriter, r *http.Request) {
	var in createCourseRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	identity, _ := web.IdentityFrom(r.Context())

	course := &contract.CourseView{
		Title:        in.Title,
		Description:  in.Description,
		ThumbnailURL: in.ThumbnailURL,
		CategoryID:   in.CategoryID,
		Lang:         in.Lang,
		Price:        in.Price,
		IsPublished:  in.IsPublished,
		InstructorID: identity.UserID,
		Modules:      toModules(in.Modules),
	}

	v := web.NewValidator()
	domain.ValidateCourse(v, course)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}
	course.Slug = domain.Slugify(course.Title)
	if course.Slug == "" {
		v.AddError("title", "must contain at least one latin letter or digit for slug generation")
		web.FailedValidation(w, v.Errors)
		return
	}

	if err := h.svc.CreateCourse(r.Context(), course); err != nil {
		if errors.Is(err, domain.ErrInvalidParent) {
			v.AddError("category_id", "category does not exist")
			web.FailedValidation(w, v.Errors)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	if err := h.decorateOne(r.Context(), course); err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"course": course},
		http.Header{"Location": []string{"/v1/courses/" + strconv.FormatInt(course.ID, 10)}})
}

func (h *Handler) updateCourse(w http.ResponseWriter, r *http.Request) {
	course, err := h.svc.GetCourse(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	if !h.canModify(r.Context(), course) {
		web.ErrorResponse(w, http.StatusForbidden,
			"your user account doesn't have the necessary permissions to access this resource")
		return
	}

	var in updateCourseRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	if in.Title != nil {
		course.Title = *in.Title
	}
	if in.Description != nil {
		course.Description = *in.Description
	}
	if in.ThumbnailURL != nil {
		course.ThumbnailURL = *in.ThumbnailURL
	}
	if in.CategoryID != nil {
		course.CategoryID = in.CategoryID
	}
	if in.Lang != nil {
		course.Lang = *in.Lang
	}
	if in.Price != nil {
		course.Price = *in.Price
	}
	if in.IsPublished != nil {
		course.IsPublished = *in.IsPublished
	}

	var newModules []contract.Module
	if in.Modules != nil {
		newModules = toModules(in.Modules)
		course.Modules = newModules
	}

	v := web.NewValidator()
	domain.ValidateCourse(v, course)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	if err := h.svc.UpdateCourse(r.Context(), course); err != nil {
		if errors.Is(err, domain.ErrInvalidParent) {
			v.AddError("category_id", "category does not exist")
			web.FailedValidation(w, v.Errors)
			return
		}
		h.writeCourseErr(w, r, err)
		return
	}
	if newModules != nil {
		if err := h.svc.ReplaceModules(r.Context(), course.ID, newModules); err != nil {
			web.ServerError(w, r, err)
			return
		}
	}

	updated, err := h.svc.GetCourse(r.Context(), strconv.FormatInt(course.ID, 10))
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	if err := h.decorateOne(r.Context(), updated); err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"course": updated}, nil)
}

func (h *Handler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	course, err := h.svc.GetCourse(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	if !h.canModify(r.Context(), course) {
		web.ErrorResponse(w, http.StatusForbidden,
			"your user account doesn't have the necessary permissions to access this resource")
		return
	}
	if err := h.svc.DeleteCourse(r.Context(), course.ID); err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "course successfully deleted"}, nil)
}

// sanitize hides paid, non-owned lesson content from the requester (paywall).
func (h *Handler) sanitize(ctx context.Context, course *contract.CourseView) {
	if len(course.Modules) == 0 || h.canModify(ctx, course) {
		return
	}
	accessible := map[int64]bool{}
	if id, ok := web.IdentityFrom(ctx); ok {
		ids, err := h.gate.AccessibleLessonIDs(ctx, id.UserID, course.ID)
		if err != nil {
			slog.Warn("paywall: failed to fetch lesson access", "err", err)
		} else {
			for _, lid := range ids {
				accessible[lid] = true
			}
		}
	}
	for mi := range course.Modules {
		for li := range course.Modules[mi].Lessons {
			l := &course.Modules[mi].Lessons[li]
			if l.IsFree || accessible[l.ID] {
				continue
			}
			l.Content = ""
			l.ContentURL = ""
			locked := true
			l.Locked = &locked
		}
	}
}

func (h *Handler) canModify(ctx context.Context, course *contract.CourseView) bool {
	id, ok := web.IdentityFrom(ctx)
	if !ok {
		return false
	}
	return id.UserID == course.InstructorID || id.Role == domain.RoleAdmin
}

// --- categories ---

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"categories": categories}, nil)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var in categoryRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	c := &domain.Category{
		NameUz: orEmpty(in.NameUz), NameRu: orEmpty(in.NameRu), NameEn: orEmpty(in.NameEn), ParentID: in.ParentID,
	}
	v := web.NewValidator()
	domain.ValidateCategory(v, c.NameUz, c.NameRu, c.NameEn)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}
	c.Slug = domain.Slugify(c.NameEn)
	if c.Slug == "" {
		v.AddError("name_en", "must contain at least one latin letter or digit for slug generation")
		web.FailedValidation(w, v.Errors)
		return
	}
	if err := h.svc.CreateCategory(r.Context(), c); err != nil {
		if h.writeCategoryWriteErr(w, v, err) {
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"category": c},
		http.Header{"Location": []string{"/v1/categories/" + strconv.FormatInt(c.ID, 10)}})
}

func (h *Handler) showCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	c, err := h.svc.GetCategory(r.Context(), id)
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"category": c}, nil)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	c, err := h.svc.GetCategory(r.Context(), id)
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	var in categoryRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	if in.NameUz != nil {
		c.NameUz = *in.NameUz
	}
	if in.NameRu != nil {
		c.NameRu = *in.NameRu
	}
	if in.NameEn != nil {
		c.NameEn = *in.NameEn
	}
	if in.ParentID != nil {
		c.ParentID = in.ParentID
	}
	v := web.NewValidator()
	domain.ValidateCategory(v, c.NameUz, c.NameRu, c.NameEn)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}
	if in.NameEn != nil {
		c.Slug = domain.Slugify(c.NameEn)
		if c.Slug == "" {
			v.AddError("name_en", "must contain at least one latin letter or digit for slug generation")
			web.FailedValidation(w, v.Errors)
			return
		}
	}
	if err := h.svc.UpdateCategory(r.Context(), c); err != nil {
		if h.writeCategoryWriteErr(w, v, err) {
			return
		}
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"category": c}, nil)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "category successfully deleted"}, nil)
}

// writeCategoryWriteErr maps category write errors to 422; returns true if handled.
func (h *Handler) writeCategoryWriteErr(w http.ResponseWriter, v *web.Validator, err error) bool {
	switch {
	case errors.Is(err, domain.ErrDuplicateSlug):
		v.AddError("name_en", "a category with this name already exists")
	case errors.Is(err, domain.ErrInvalidParent):
		v.AddError("parent_id", "parent category does not exist")
	case errors.Is(err, domain.ErrMaxDepth):
		v.AddError("parent_id", "category nesting is too deep (max 2 levels)")
	default:
		return false
	}
	web.FailedValidation(w, v.Errors)
	return true
}
