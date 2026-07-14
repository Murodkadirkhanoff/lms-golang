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

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuthenticated(app.Responder))

			// Gateway ~^/v1/courses/[^/]+/enroll$ ni shu servisga yo'naltiradi.
			r.Post("/courses/{id}/enroll", app.enrollHandler)
			r.Patch("/enrollments/{id}/progress", app.updateProgressHandler)

			r.Route("/me", func(r chi.Router) {
				r.Get("/stats", app.meStatsHandler)
				r.Get("/teaching/stats", app.meTeachingStatsHandler)
				r.Get("/courses", app.meCoursesHandler)
				r.Get("/orders", app.meOrdersHandler)
				r.Post("/orders", app.checkoutHandler)
				r.Get("/orders/{id}", app.meOrderHandler)
				r.Get("/certificates", app.meCertificatesHandler)
				r.Get("/notifications", app.meNotificationsHandler)
				r.Post("/notifications/read-all", app.readAllNotificationsHandler)
			})
		})
	})

	// Ichki endpointlar — gateway orqali chiqmaydi, X-Internal-Key talab qilinadi.
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalOnly(app.config.internalKey, app.Responder))
		r.Get("/stats", app.internalStatsHandler)
		r.Get("/enrollment-counts", app.internalEnrollmentCountsHandler)
	})

	return r
}
