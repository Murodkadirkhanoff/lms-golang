package transport

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// --- reviews ---

type createReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())

	enrolled, err := h.gate.IsEnrolled(r.Context(), identity.UserID, id)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	if !enrolled {
		web.ErrorResponse(w, http.StatusForbidden, "you must be enrolled in this course to leave a review")
		return
	}

	var in createReviewRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	review := &contract.Review{
		CourseID: id,
		UserID:   identity.UserID,
		User:     h.svc.LookupUserName(r.Context(), identity.UserID, "Student"),
		Rating:   in.Rating,
		Comment:  in.Comment,
	}
	v := web.NewValidator()
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
	v.Check(web.ByteLength(review.Comment) <= 2000, "comment", "must not be more than 2000 bytes long")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	if err := h.svc.UpsertReview(r.Context(), review); err != nil {
		if errors.Is(err, domain.ErrInvalidCourse) {
			web.NotFound(w)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"review": review}, nil)
}

// --- instructors ---

func (h *Handler) listInstructors(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Instructors(r.Context())
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"items": items}, nil)
}

func (h *Handler) showInstructor(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	inst, err := h.svc.Instructor(r.Context(), id)
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"instructor": inst}, nil)
}

// --- lesson Q&A ---

type askRequest struct {
	Question string `json:"question"`
}

func (h *Handler) listQuestions(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	items, err := h.svc.ListQuestions(r.Context(), id)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"items": items}, nil)
}

func (h *Handler) askQuestion(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())

	var in askRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	question := strings.TrimSpace(in.Question)

	v := web.NewValidator()
	v.Check(question != "", "question", "must be provided")
	v.Check(web.ByteLength(question) <= 2000, "question", "must not be more than 2000 bytes long")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	userName := h.svc.LookupUserName(r.Context(), identity.UserID, "Student")
	saved, err := h.svc.AskQuestion(r.Context(), id, identity.UserID, userName, question)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			web.NotFound(w)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"question": saved}, nil)
}
