package transport

import (
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Routes registers the enrollment context endpoints. Every endpoint requires
// an authenticated user.
func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(web.RequireAuth)

		r.Post("/v1/courses/{id}/enroll", h.enroll)
		r.Patch("/v1/enrollments/{id}/progress", h.updateProgress)

		r.Get("/v1/me/stats", h.stats)
		r.Get("/v1/me/courses", h.myCourses)
		r.Get("/v1/me/certificates", h.myCertificates)
		r.Get("/v1/me/certificates/{id}/download", h.downloadCertificate)
		r.Get("/v1/me/notifications", h.myNotifications)
		r.Post("/v1/me/notifications/{id}/read", h.readNotification)
		r.Post("/v1/me/notifications/read-all", h.readAllNotifications)
		r.Get("/v1/me/orders", h.myOrders)
		r.Get("/v1/me/orders/{id}", h.myOrder)
		r.Post("/v1/me/orders", h.checkout)
		r.Get("/v1/me/teaching/stats", h.teachingStats)
	})
}
