// Package transport is the enrollment context's HTTP adapter.
package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Handler serves the enrollment context endpoints.
type Handler struct {
	svc *application.Service
}

// NewHandler builds the enrollment HTTP handler.
func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) writeErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		web.NotFound(w)
	case errors.Is(err, domain.ErrNotPermitted):
		web.ErrorResponse(w, http.StatusForbidden,
			"your user account doesn't have the necessary permissions to access this resource")
	default:
		web.ServerError(w, r, err)
	}
}

// enroll enrols the user in a free, published course.
func (h *Handler) enroll(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())

	course, err := h.svc.CourseByID(r.Context(), id)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	if course == nil {
		web.NotFound(w)
		return
	}

	v := web.NewValidator()
	v.Check(course.IsPublished, "course", "course is not published")
	v.Check(course.Price == 0, "course", "this course is not free, please purchase it via checkout")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	enrollment, err := h.svc.Enroll(r.Context(), identity.UserID, *course)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"enrollment": enrollment}, nil)
}

type updateProgressRequest struct {
	LessonID  int64 `json:"lesson_id"`
	Completed bool  `json:"completed"`
}

// updateProgress marks a lesson complete/incomplete for the owner's enrollment.
func (h *Handler) updateProgress(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())

	enrollment, err := h.svc.FindEnrollment(r.Context(), id)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if enrollment.UserID != identity.UserID {
		web.ErrorResponse(w, http.StatusForbidden,
			"your user account doesn't have the necessary permissions to access this resource")
		return
	}

	var in updateProgressRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	v := web.NewValidator()
	v.Check(in.LessonID > 0, "lesson_id", "must be provided")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	err = h.svc.UpdateProgress(r.Context(), identity.UserID, in.LessonID, in.Completed, enrollment.CourseID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			v.AddError("lesson_id", "you don't have access to this lesson")
			web.FailedValidation(w, v.Errors)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "progress updated"}, nil)
}
