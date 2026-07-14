package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"lms.chashma.uz/pkg/auth"
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

		r.Post("/users", app.registerUserHandler)
		r.Put("/users/password", app.resetPasswordHandler)

		r.Post("/tokens/authentication", app.createAuthenticationTokenHandler)
		r.Post("/tokens/password-reset", app.forgotPasswordHandler)

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireRole(app.Responder, auth.RoleAdmin))
			r.Get("/users", app.adminListUsersHandler)
			r.Get("/stats", app.adminStatsHandler)
		})
	})

	// Ichki endpointlar — gateway orqali chiqmaydi, X-Internal-Key talab qilinadi.
	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalOnly(app.config.internalKey, app.Responder))
		r.Get("/users", app.internalUsersHandler)
	})

	return r
}
