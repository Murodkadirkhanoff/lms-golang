package transport

import (
	"github.com/chashma/lms/internal/modules/users/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Routes registers the users context endpoints. The global Authenticate
// middleware has already populated the request identity; here we only add the
// per-group authorisation guards.
func (h *Handler) Routes(r chi.Router) {
	// Public (rate-limited at the app edge).
	r.Post("/v1/users", h.register)
	r.Post("/v1/tokens/authentication", h.login)
	r.Post("/v1/tokens/password-reset", h.forgotPassword)
	r.Put("/v1/users/password", h.resetPassword)

	// Authenticated self-service.
	r.Group(func(r chi.Router) {
		r.Use(web.RequireAuth)
		r.Get("/v1/me", h.me)
		r.Put("/v1/me/profile", h.updateProfile)
		r.Put("/v1/me/password", h.changePassword)
	})

	// Admin only.
	r.Group(func(r chi.Router) {
		r.Use(web.RequireRole(domain.RoleAdmin))
		r.Get("/v1/admin/users", h.listUsers)
		r.Patch("/v1/admin/users/{id}/role", h.updateRole)
		r.Get("/v1/admin/stats", h.stats)
	})
}
