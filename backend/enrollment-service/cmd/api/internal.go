package main

import (
	"errors"
	"net/http"

	"lms.chashma.uz/pkg/jsonutil"
)

// internalStatsHandler — GET /internal/stats. Admin panel uchun daromad.
func (app *application) internalStatsHandler(w http.ResponseWriter, r *http.Request) {
	revenue, err := app.models.Orders.Revenue()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	stats := struct {
		Revenue float64 `json:"revenue"`
	}{revenue}

	err = jsonutil.WriteJSON(w, http.StatusOK, stats, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalAccessHandler — GET /internal/access?user_id=N&course_id=N.
// course-service paywall uchun: foydalanuvchining kursda kirish huquqi bor
// dars id'lari.
func (app *application) internalAccessHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	userID := int64(jsonutil.ReadInt(qs, "user_id", 0))
	courseID := int64(jsonutil.ReadInt(qs, "course_id", 0))
	if userID == 0 || courseID == 0 {
		app.BadRequest(w, r, errors.New("user_id and course_id query parameters must be provided"))
		return
	}

	ids, err := app.models.Enrollments.AccessibleLessonIDs(userID, courseID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"lessonIds": ids}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalEnrollmentCountsHandler — GET /internal/enrollment-counts?ids=<userIDs>.
// Admin users ro'yxati uchun: user -> yozilgan kurslari soni.
func (app *application) internalEnrollmentCountsHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	counts, err := app.models.Enrollments.EnrollmentCountsByUser(ids)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"counts": counts}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
