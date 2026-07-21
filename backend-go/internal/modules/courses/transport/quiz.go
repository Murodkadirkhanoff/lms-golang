package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

type quizQuestionRequest struct {
	Question     string   `json:"question"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
}

type upsertQuizRequest struct {
	Title            string                `json:"title"`
	PassingScore     int                   `json:"passing_score"`
	TimeLimitMinutes int                   `json:"time_limit_minutes"`
	Questions        []quizQuestionRequest `json:"questions"`
}

type submitAttemptRequest struct {
	Score int `json:"score"`
}

func (h *Handler) showQuiz(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	quiz, err := h.svc.QuizByCourse(r.Context(), id)
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"quiz": quiz}, nil)
}

func (h *Handler) upsertQuiz(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	course, err := h.svc.GetCourse(r.Context(), strconv.FormatInt(id, 10))
	if err != nil {
		h.writeCourseErr(w, r, err)
		return
	}
	if !h.canModify(r.Context(), course) {
		web.ErrorResponse(w, http.StatusForbidden,
			"your user account doesn't have the necessary permissions to access this resource")
		return
	}

	var in upsertQuizRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	quiz := &domain.Quiz{CourseID: id, Title: in.Title, PassingScore: in.PassingScore, TimeLimitMinutes: in.TimeLimitMinutes}
	if quiz.PassingScore == 0 {
		quiz.PassingScore = 70
	}
	if quiz.TimeLimitMinutes == 0 {
		quiz.TimeLimitMinutes = 10
	}
	quiz.Questions = []domain.QuizQuestion{}
	for _, q := range in.Questions {
		opts := q.Options
		if opts == nil {
			opts = []string{}
		}
		quiz.Questions = append(quiz.Questions, domain.QuizQuestion{
			Question: q.Question, Options: opts, CorrectIndex: q.CorrectIndex,
		})
	}

	v := web.NewValidator()
	domain.ValidateQuiz(v, quiz)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	if err := h.svc.UpsertQuiz(r.Context(), quiz); err != nil {
		if errors.Is(err, domain.ErrInvalidCourse) {
			web.NotFound(w)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"quiz": quiz}, nil)
}

func (h *Handler) submitAttempt(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if _, err := h.svc.QuizByCourse(r.Context(), id); err != nil {
		h.writeCourseErr(w, r, err)
		return
	}

	var in submitAttemptRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	v := web.NewValidator()
	v.Check(in.Score >= 0 && in.Score <= 100, "score", "must be between 0 and 100")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	identity, _ := web.IdentityFrom(r.Context())
	attempt := &domain.QuizAttempt{UserID: identity.UserID, CourseID: id, Score: in.Score}
	if err := h.svc.SubmitAttempt(r.Context(), attempt); err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"attempt": attempt}, nil)
}

func (h *Handler) listAttempts(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())
	attempts, err := h.svc.ListAttempts(r.Context(), identity.UserID, id)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"attempts": attempts}, nil)
}
