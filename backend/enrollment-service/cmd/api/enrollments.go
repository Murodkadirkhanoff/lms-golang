package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"lms.chashma.uz/enrollment-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
	"lms.chashma.uz/pkg/svcclient"
	"lms.chashma.uz/pkg/validator"
)

// enrollHandler — POST /v1/courses/{id}/enroll. Faqat bepul kurslar uchun;
// pullik kurslar checkout orqali sotib olinadi.
func (app *application) enrollHandler(w http.ResponseWriter, r *http.Request) {
	courseID, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	claims := middleware.ContextGetUser(r)

	courses, err := app.fetchCourses(r.Context(), []int64{courseID})
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	course, ok := courses[courseID]
	if !ok {
		app.NotFound(w, r)
		return
	}

	v := validator.New()
	v.Check(course.IsPublished, "course", "course is not published")
	v.Check(course.Price == 0, "course", "this course is not free, please purchase it via checkout")
	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	enrollment, isNew, err := app.models.Enrollments.Insert(claims.UserID, courseID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = app.grantCourseAccess(r.Context(), claims.UserID, courseID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	if isNew {
		app.markEnrolled(r.Context(), courseID)
		app.notify(claims.UserID, "course", "Enrolled in a course",
			"You are now enrolled in \""+course.Title+"\". Happy learning!")
	}

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"enrollment": enrollment}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// updateProgressHandler — PATCH /v1/enrollments/{id}/progress.
// Body: {"lesson_id": N, "completed": true|false}.
func (app *application) updateProgressHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	claims := middleware.ContextGetUser(r)

	enrollment, err := app.models.Enrollments.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if enrollment.UserID != claims.UserID {
		app.NotPermitted(w, r)
		return
	}

	var input struct {
		LessonID  int64 `json:"lesson_id"`
		Completed bool  `json:"completed"`
	}

	err = jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.LessonID > 0, "lesson_id", "must be provided")
	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Enrollments.SetLessonCompleted(claims.UserID, input.LessonID, input.Completed)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("lesson_id", "you don't have access to this lesson")
			app.FailedValidation(w, r, v.Errors)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	if input.Completed {
		app.maybeIssueCertificate(r, claims.UserID, enrollment.CourseID)
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": "progress updated"}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// maybeIssueCertificate kursning barcha darslari tugatilgan bo'lsa sertifikat
// beradi. Yordamchi oqim — xatolar progress so'rovini yiqitmaydi.
func (app *application) maybeIssueCertificate(r *http.Request, userID, courseID int64) {
	courses, err := app.fetchCourses(r.Context(), []int64{courseID})
	if err != nil {
		app.logger.Warn("certificate check: failed to fetch course", "error", err.Error())
		return
	}

	course, ok := courses[courseID]
	if !ok || course.TotalLessons == 0 {
		return
	}

	completed, err := app.models.Enrollments.CompletedCounts(userID)
	if err != nil {
		app.logger.Warn("certificate check: failed to count lessons", "error", err.Error())
		return
	}

	if completed[courseID] < course.TotalLessons {
		return
	}

	issued, err := app.models.Certificates.Issue(userID, courseID, course.Title)
	if err != nil {
		app.logger.Warn("failed to issue certificate", "error", err.Error())
		return
	}

	if issued {
		app.notify(userID, "course", "Certificate earned",
			"Congratulations! You completed \""+course.Title+"\" and earned a certificate.")
	}
}

// enrolledCourse frontend EnrolledCourse tipiga mos: course to'liq Course
// JSON'i (course-service'dan o'zgarishsiz uzatiladi). enrollmentId va
// completedLessonIds learn sahifasiga progress yozish uchun kerak.
type enrolledCourse struct {
	EnrollmentID       int64           `json:"enrollmentId"`
	Course             json.RawMessage `json:"course"`
	Progress           int             `json:"progress"`
	CurrentLesson      string          `json:"currentLesson"`
	LessonsCompleted   int             `json:"lessonsCompleted"`
	CompletedLessonIDs []int64         `json:"completedLessonIds"`
}

// meCoursesHandler — GET /v1/me/courses.
func (app *application) meCoursesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	enrollments, err := app.models.Enrollments.ListByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	courseIDs := make([]int64, 0, len(enrollments))
	for _, e := range enrollments {
		courseIDs = append(courseIDs, e.CourseID)
	}

	courses, err := app.fetchCourses(r.Context(), courseIDs)
	if err != nil && !errors.Is(err, svcclient.ErrNotFound) {
		app.ServerError(w, r, err)
		return
	}

	completed, err := app.models.Enrollments.CompletedCounts(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	// currentLesson: tartibdagi birinchi tugatilmagan dars nomi.
	lessonsByCourse, err := app.fetchLessonsByCourses(r.Context(), courseIDs)
	if err != nil {
		app.logger.Warn("failed to fetch lessons for currentLesson", "error", err.Error())
		lessonsByCourse = map[int64][]*lessonInfo{}
	}

	doneLessons, err := app.models.Enrollments.CompletedLessonIDs(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	items := []*enrolledCourse{}
	for _, e := range enrollments {
		course, ok := courses[e.CourseID]
		if !ok {
			continue // kurs o'chirilgan bo'lishi mumkin
		}

		done := completed[e.CourseID]
		progress := 0
		if course.TotalLessons > 0 {
			progress = done * 100 / course.TotalLessons
		}

		currentLesson := ""
		completedIDs := []int64{}
		for _, lesson := range lessonsByCourse[e.CourseID] {
			if doneLessons[lesson.ID] {
				completedIDs = append(completedIDs, lesson.ID)
			} else if currentLesson == "" {
				currentLesson = lesson.Title
			}
		}

		items = append(items, &enrolledCourse{
			EnrollmentID:       e.ID,
			Course:             course.Raw,
			Progress:           progress,
			CurrentLesson:      currentLesson,
			LessonsCompleted:   done,
			CompletedLessonIDs: completedIDs,
		})
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"items": items}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// meStatsHandler — GET /v1/me/stats (dashboard.service DashboardStats shakli).
func (app *application) meStatsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	enrollments, err := app.models.Enrollments.ListByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	courseIDs := make([]int64, 0, len(enrollments))
	for _, e := range enrollments {
		courseIDs = append(courseIDs, e.CourseID)
	}

	courses, err := app.fetchCourses(r.Context(), courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	completed, err := app.models.Enrollments.CompletedCounts(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	certificates, err := app.models.Certificates.CountByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	stats := struct {
		Enrolled     int `json:"enrolled"`
		InProgress   int `json:"inProgress"`
		Completed    int `json:"completed"`
		Certificates int `json:"certificates"`
	}{
		Enrolled:     len(enrollments),
		Certificates: certificates,
	}

	for _, e := range enrollments {
		course, ok := courses[e.CourseID]
		if !ok || course.TotalLessons == 0 {
			continue
		}

		done := completed[e.CourseID]
		switch {
		case done >= course.TotalLessons:
			stats.Completed++
		case done > 0:
			stats.InProgress++
		}
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, stats, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
