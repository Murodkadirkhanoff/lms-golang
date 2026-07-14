package main

import (
	"errors"
	"net/http"
	"strconv"

	"lms.chashma.uz/course-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/validator"
)

// showQuizHandler — GET /v1/quizzes/{id}. Frontend {id} sifatida KURS id'sini
// yuboradi (learn sahifasi ROUTES.quiz(course.id)).
func (app *application) showQuizHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	quiz, err := app.models.Quizzes.GetByCourseID(courseID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"quiz": quiz}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// upsertQuizHandler — PUT /v1/courses/{id}/quiz (kurs egasi yoki admin).
// Kursning quizini savollari bilan butunlay almashtiradi.
func (app *application) upsertQuizHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	course, err := app.models.Courses.GetByIDOrSlug(strconv.FormatInt(courseID, 10))
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
		Title            string `json:"title"`
		PassingScore     int    `json:"passing_score"`
		TimeLimitMinutes int    `json:"time_limit_minutes"`
		Questions        []struct {
			Question     string   `json:"question"`
			Options      []string `json:"options"`
			CorrectIndex int      `json:"correct_index"`
		} `json:"questions"`
	}

	err = jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	quiz := &data.Quiz{
		CourseID:         courseID,
		Title:            input.Title,
		PassingScore:     input.PassingScore,
		TimeLimitMinutes: input.TimeLimitMinutes,
	}

	if quiz.PassingScore == 0 {
		quiz.PassingScore = 70
	}
	if quiz.TimeLimitMinutes == 0 {
		quiz.TimeLimitMinutes = 10
	}

	for _, q := range input.Questions {
		quiz.Questions = append(quiz.Questions, &data.QuizQuestion{
			Question:     q.Question,
			Options:      q.Options,
			CorrectIndex: q.CorrectIndex,
		})
	}

	v := validator.New()

	if data.ValidateQuiz(v, quiz); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Quizzes.Upsert(quiz)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCourse):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"quiz": quiz}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
