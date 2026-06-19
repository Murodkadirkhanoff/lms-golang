package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedReponse)
	// GET/POST /v1/courses
	// GET/PATCH/DELETE /v1/courses/:id
	// GET/POST /v1/courses/:id/lessons
	// GET/PATCH/DELETE /v1/lessons/:id
	// POST /v1/courses/:id/enroll
	// GET /v1/users/:id/enrollments
	// PATCH /v1/enrollments/:id/progress

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/categories", app.createCategoryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/categories/:id", app.showCategoryHandler)

	return app.recoverPanic(router)
}
