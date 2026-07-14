package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"lms.chashma.uz/pkg/middleware"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RecoverPanic(app.Responder))
	r.Use(middleware.EnableCORS(app.config.trustedOrigins))
	r.Use(middleware.Authenticate(app.config.jwtSecret, app.Responder))

	r.NotFound(app.NotFound)
	r.MethodNotAllowed(app.MethodNotAllowed)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthcheckHandler)

		r.Get("/categories", app.listCategoriesHandler)
		r.Post("/categories", app.createCategoryHandler)
		r.Get("/categories/{id}", app.showCategoryHandler)
		r.Patch("/categories/{id}", app.updateCategoryHandler)
		r.Delete("/categories/{id}", app.deleteCategoryHandler)

		r.Get("/courses", app.listCoursesHandler)
		r.Get("/courses/{idOrSlug}", app.showCourseHandler)
		r.Get("/quizzes/{id}", app.showQuizHandler)

		// Kurs yaratish uchun autentifikatsiya yetarli — frontend qarori
		// bo'yicha role-gating yo'q (har user o'qishi ham, o'qitishi ham mumkin).
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuthenticated(app.Responder))
			r.Post("/courses", app.createCourseHandler)
			r.Patch("/courses/{id}", app.updateCourseHandler)
			r.Delete("/courses/{id}", app.deleteCourseHandler)
			r.Post("/courses/{id}/reviews", app.createReviewHandler)
			r.Put("/courses/{id}/quiz", app.upsertQuizHandler)
		})

		r.Get("/instructors", app.listInstructorsHandler)
		r.Get("/instructors/{id}", app.showInstructorHandler)
	})

	// Ichki endpointlar — gateway orqali chiqmaydi, X-Internal-Key talab qilinadi.
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalOnly(app.config.internalKey, app.Responder))
		r.Get("/courses", app.internalListCoursesHandler)
		r.Get("/courses/lessons", app.internalCoursesLessonsHandler)
		r.Get("/courses/{id}/lessons", app.internalCourseLessonsHandler)
		r.Post("/courses/{id}/enrolled", app.internalEnrolledHandler)
		r.Get("/lessons", app.internalLessonsHandler)
		r.Get("/stats", app.internalStatsHandler)
		r.Get("/course-counts", app.internalCourseCountsHandler)
	})

	return r
}
