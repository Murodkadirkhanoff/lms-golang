// Package httperr xato javoblarini markazlashtiradi. Har servis application
// struct'iga Responder'ni embed qiladi (eski cmd/api/errors.go dan ko'chirilgan).
package httperr

import (
	"fmt"
	"log/slog"
	"net/http"

	"lms.chashma.uz/pkg/jsonutil"
)

type Responder struct {
	Logger *slog.Logger
}

func (rs Responder) logError(r *http.Request, err error) {
	rs.Logger.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
}

// ErrorResponse xato envelope'ini yozadi: {"error": string yoki {field: message}}.
func (rs Responder) ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := jsonutil.Envelope{"error": message}

	err := jsonutil.WriteJSON(w, status, env, nil)
	if err != nil {
		rs.logError(r, err)
		w.WriteHeader(500)
	}
}

func (rs Responder) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	rs.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	rs.ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func (rs Responder) NotFound(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	rs.ErrorResponse(w, r, http.StatusNotFound, message)
}

func (rs Responder) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	rs.ErrorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (rs Responder) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	rs.ErrorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (rs Responder) FailedValidation(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	rs.ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (rs Responder) EditConflict(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	rs.ErrorResponse(w, r, http.StatusConflict, message)
}

func (rs Responder) InvalidCredentials(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	rs.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

func (rs Responder) InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	rs.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

func (rs Responder) AuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	rs.ErrorResponse(w, r, http.StatusUnauthorized, message)
}

func (rs Responder) NotPermitted(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	rs.ErrorResponse(w, r, http.StatusForbidden, message)
}
