package main

import (
	"errors"
	"net/http"

	"lms.chashma.uz/course-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
	"lms.chashma.uz/pkg/validator"
)

// createReviewHandler — POST /v1/courses/{id}/reviews (auth).
// Bitta user bitta kursga bitta sharh; qayta yuborsa yangilanadi.
func (app *application) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	var input struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}

	err = jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	claims := middleware.ContextGetUser(r)

	// Ism snapshot uchun auth-service'dan olinadi.
	userName := "Student"
	if users := app.fetchUsers(r.Context(), []int64{claims.UserID}); len(users) > 0 {
		userName = users[claims.UserID].Name
	}

	review := &data.Review{
		CourseID: courseID,
		UserID:   claims.UserID,
		User:     userName,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	v := validator.New()

	if data.ValidateReview(v, review); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Reviews.Upsert(review)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCourse):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"review": review}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
