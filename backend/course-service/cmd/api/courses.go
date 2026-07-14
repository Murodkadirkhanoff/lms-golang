package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"lms.chashma.uz/course-service/internal/data"
	"lms.chashma.uz/pkg/auth"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
	"lms.chashma.uz/pkg/uidefaults"
	"lms.chashma.uz/pkg/validator"
)

func (app *application) listCoursesHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	filters := data.CourseFilters{
		Search:       jsonutil.ReadString(qs, "search", ""),
		CategorySlug: jsonutil.ReadString(qs, "category", ""),
		Sort:         jsonutil.ReadString(qs, "sort", "popular"),
		Page:         jsonutil.ReadInt(qs, "page", 1),
		PageSize:     jsonutil.ReadInt(qs, "pageSize", 8),
		IDs:          jsonutil.ReadIDList(qs, "ids"),
		InstructorID: int64(jsonutil.ReadInt(qs, "instructorId", 0)),
	}

	// Studio o'z kurslarini (draft ham) instructorId bilan so'raydi.
	filters.IncludeUnpublished = filters.InstructorID != 0

	v := validator.New()
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.PageSize > 0 && filters.PageSize <= 100, "pageSize", "must be between 1 and 100")
	v.Check(validator.PermittedValue(filters.Sort, "popular", "newest", "price-asc", "price-desc"),
		"sort", "must be one of popular, newest, price-asc, price-desc")
	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	courses, total, err := app.models.Courses.List(filters)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	app.decorateCourses(r.Context(), courses)

	result := jsonutil.Paginated[*data.Course]{
		Items:    courses,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    total,
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *application) showCourseHandler(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "idOrSlug")

	course, err := app.models.Courses.GetByIDOrSlug(idOrSlug)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	course.Reviews, err = app.models.Reviews.ListForCourse(course.ID, 20)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	app.decorateCourses(r.Context(), []*data.Course{course})

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"course": course}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// createCourseHandler — POST /v1/courses. Request kalitlari snake_case
// (frontend courses.service.ts:143), instructor_id tokendagi userdan olinadi.
func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		CategoryID  *int64  `json:"category_id"`
		Lang        string  `json:"lang"`
		Price       float64 `json:"price"`
		IsPublished bool    `json:"is_published"`
		Modules     []struct {
			Title    string `json:"title"`
			Position int    `json:"position"`
			Lessons  []struct {
				Title           string  `json:"title"`
				Type            string  `json:"type"`
				ContentURL      string  `json:"content_url"`
				Content         string  `json:"content"`
				DurationSeconds int     `json:"duration_seconds"`
				Position        int     `json:"position"`
				Price           float64 `json:"price"`
				IsFree          bool    `json:"is_free"`
			} `json:"lessons"`
		} `json:"modules"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	claims := middleware.ContextGetUser(r)

	course := &data.Course{
		Title:        input.Title,
		Description:  input.Description,
		CategoryID:   input.CategoryID,
		Lang:         input.Lang,
		Price:        input.Price,
		IsPublished:  input.IsPublished,
		InstructorID: claims.UserID,
	}

	for _, m := range input.Modules {
		module := &data.Module{Title: m.Title, Position: m.Position}
		for _, l := range m.Lessons {
			lessonType := l.Type
			if lessonType == "" {
				lessonType = "video"
			}
			module.Lessons = append(module.Lessons, &data.Lesson{
				Title:           l.Title,
				Type:            lessonType,
				ContentURL:      l.ContentURL,
				Content:         l.Content,
				DurationSeconds: l.DurationSeconds,
				Position:        l.Position,
				Price:           l.Price,
				IsFree:          l.IsFree,
			})
		}
		course.Modules = append(course.Modules, module)
	}

	v := validator.New()

	if data.ValidateCourse(v, course); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	course.Slug = data.Slugify(course.Title)
	if course.Slug == "" {
		v.AddError("title", "must contain at least one latin letter or digit for slug generation")
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Courses.Insert(course)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidParent):
			v.AddError("category_id", "category does not exist")
			app.FailedValidation(w, r, v.Errors)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	app.decorateCourses(r.Context(), []*data.Course{course})

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/courses/%d", course.ID))

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"course": course}, headers)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *application) updateCourseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	course, err := app.models.Courses.GetByIDOrSlug(strconv.FormatInt(id, 10))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !app.canModifyCourse(r, course) {
		app.NotPermitted(w, r)
		return
	}

	var input struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		CategoryID  *int64   `json:"category_id"`
		Lang        *string  `json:"lang"`
		Price       *float64 `json:"price"`
		IsPublished *bool    `json:"is_published"`
	}

	err = jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if input.Title != nil {
		course.Title = *input.Title
	}
	if input.Description != nil {
		course.Description = *input.Description
	}
	if input.CategoryID != nil {
		course.CategoryID = input.CategoryID
	}
	if input.Lang != nil {
		course.Lang = *input.Lang
	}
	if input.Price != nil {
		course.Price = *input.Price
	}
	if input.IsPublished != nil {
		course.IsPublished = *input.IsPublished
	}

	v := validator.New()

	if data.ValidateCourse(v, course); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Courses.Update(course)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.EditConflict(w, r)
		case errors.Is(err, data.ErrInvalidParent):
			v.AddError("category_id", "category does not exist")
			app.FailedValidation(w, r, v.Errors)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	app.decorateCourses(r.Context(), []*data.Course{course})

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"course": course}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *application) deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	course, err := app.models.Courses.GetByIDOrSlug(strconv.FormatInt(id, 10))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if !app.canModifyCourse(r, course) {
		app.NotPermitted(w, r)
		return
	}

	err = app.models.Courses.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": "course successfully deleted"}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// canModifyCourse — faqat kurs egasi yoki admin o'zgartira oladi.
func (app *application) canModifyCourse(r *http.Request, course *data.Course) bool {
	claims := middleware.ContextGetUser(r)
	if claims == nil {
		return false
	}
	return claims.UserID == course.InstructorID || claims.Role == auth.RoleAdmin
}

// userInfo auth-service /internal/users javobidagi elementlar.
type userInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// fetchUsers auth-service'dan user ma'lumotlarini batch oladi.
func (app *application) fetchUsers(ctx context.Context, ids []int64) map[int64]userInfo {
	users := map[int64]userInfo{}
	if len(ids) == 0 {
		return users
	}

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}

	var response struct {
		Users []userInfo `json:"users"`
	}

	err := app.authClient.Get(ctx, "/internal/users?ids="+url.QueryEscape(strings.Join(parts, ",")), &response)
	if err != nil {
		// Auth ishlamasa ham kurslar ko'rinaveradi — instruktor nomi default bo'ladi.
		app.logger.Warn("failed to fetch users from auth-service", "error", err.Error())
		return users
	}

	for _, u := range response.Users {
		users[u.ID] = u
	}

	return users
}

// decorateCourses instruktor obyekti va UI-default maydonlarni to'ldiradi.
func (app *application) decorateCourses(ctx context.Context, courses []*data.Course) {
	idSet := map[int64]bool{}
	for _, c := range courses {
		idSet[c.InstructorID] = true
	}

	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	users := app.fetchUsers(ctx, ids)

	for _, c := range courses {
		c.ThumbnailColor = uidefaults.ThumbnailColor(c.ID)

		name := "Instructor"
		if u, ok := users[c.InstructorID]; ok {
			name = u.Name
		}

		c.Instructor = &data.Instructor{
			ID:          c.InstructorID,
			Name:        name,
			Headline:    "Instructor",
			AvatarColor: uidefaults.AvatarColor(c.InstructorID),
		}
	}
}
