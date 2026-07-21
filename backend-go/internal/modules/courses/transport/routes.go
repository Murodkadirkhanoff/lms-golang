package transport

import (
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Routes registers the courses context endpoints. Public reads and the
// (historically unauthenticated) category management sit outside the auth
// guard; writes that require an identity sit inside it.
func (h *Handler) Routes(r chi.Router) {
	// Public reads.
	r.Get("/v1/courses", h.listCourses)
	r.Get("/v1/courses/{id}", h.showCourse)
	r.Get("/v1/categories", h.listCategories)
	r.Post("/v1/categories", h.createCategory)
	r.Get("/v1/categories/{id}", h.showCategory)
	r.Patch("/v1/categories/{id}", h.updateCategory)
	r.Delete("/v1/categories/{id}", h.deleteCategory)
	r.Get("/v1/instructors", h.listInstructors)
	r.Get("/v1/instructors/{id}", h.showInstructor)
	r.Get("/v1/quizzes/{id}", h.showQuiz)
	r.Get("/v1/lessons/{id}/questions", h.listQuestions)

	// Authenticated writes.
	r.Group(func(r chi.Router) {
		r.Use(web.RequireAuth)
		r.Post("/v1/courses", h.createCourse)
		r.Patch("/v1/courses/{id}", h.updateCourse)
		r.Delete("/v1/courses/{id}", h.deleteCourse)
		r.Put("/v1/courses/{id}/quiz", h.upsertQuiz)
		r.Post("/v1/courses/{id}/reviews", h.createReview)
		r.Post("/v1/quizzes/{id}/attempts", h.submitAttempt)
		r.Get("/v1/quizzes/{id}/attempts", h.listAttempts)
		r.Post("/v1/lessons/{id}/questions", h.askQuestion)
	})
}
