package main

import (
	"errors"
	"net/http"

	"lms.chashma.uz/course-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
)

// Ichki endpointlar enrollment-service uchun: kurs narxi/holati (checkout),
// dars ro'yxati (lesson_access to'ldirish) va to'liq kurs obyektlari (me/courses).

// internalListCoursesHandler — GET /internal/courses?ids=1,2.
// To'liq Course JSON qaytaradi (public list bilan bir xil shakl).
func (app *application) internalListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	courses, _, err := app.models.Courses.List(data.CourseFilters{
		IDs:                ids,
		Page:               1,
		PageSize:           len(ids),
		IncludeUnpublished: true,
	})
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	app.decorateCourses(r.Context(), courses)

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"courses": courses}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalCourseLessonsHandler — GET /internal/courses/{id}/lessons.
func (app *application) internalCourseLessonsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	lessons, err := app.models.Courses.LessonsForCourse(id)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"lessons": lessons}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalCoursesLessonsHandler — GET /internal/courses/lessons?ids=<courseIDs>.
// Bir nechta kursning darslarini tartib bilan qaytaradi (currentLesson uchun).
func (app *application) internalCoursesLessonsHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	lessons, err := app.models.Courses.LessonsForCourses(ids)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"lessons": lessons}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalEnrolledHandler — POST /internal/courses/{id}/enrolled.
// enrollment-service yangi enrollment yaratganda student_count'ni oshiradi.
func (app *application) internalEnrolledHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	err = app.models.Courses.IncrementStudentCount(id)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": "ok"}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalStatsHandler — GET /internal/stats (admin panel aggregatsiyasi).
func (app *application) internalStatsHandler(w http.ResponseWriter, r *http.Request) {
	totalCourses, activeInstructors, err := app.models.Courses.Stats()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	stats := struct {
		TotalCourses      int `json:"totalCourses"`
		ActiveInstructors int `json:"activeInstructors"`
	}{totalCourses, activeInstructors}

	err = jsonutil.WriteJSON(w, http.StatusOK, stats, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalCourseCountsHandler — GET /internal/course-counts?ids=<userIDs>.
// Admin users ro'yxati uchun: user -> yaratgan kurslari soni.
func (app *application) internalCourseCountsHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	counts, err := app.models.Courses.CourseCountsByInstructor(ids)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"counts": counts}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalLessonsHandler — GET /internal/lessons?ids=1,2 (checkout uchun).
func (app *application) internalLessonsHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	lessons, err := app.models.Courses.LessonsByIDs(ids)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"lessons": lessons}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
