package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

// Envelope is the standard JSON object wrapper used for all responses.
type Envelope map[string]any

// WriteJSON serialises data as JSON with the given status code and headers.
func WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for k, v := range headers {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

// ReadJSON decodes a single JSON object from the request body into dst,
// rejecting unknown fields and malformed input with friendly messages that
// match the Java GlobalExceptionHandler.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return &BadRequest{Message: "body contains badly-formed JSON"}
	}
	if dec.More() {
		return &BadRequest{Message: "body must only contain a single JSON value"}
	}
	return nil
}

// The generic HTTP error responders below are the Go equivalent of the Java
// GlobalExceptionHandler / httperr.Responder. They only shape the envelope;
// mapping domain errors to them is each module's transport concern.

// ErrorResponse writes {"error": message} with the given status.
func ErrorResponse(w http.ResponseWriter, status int, message any) {
	if err := WriteJSON(w, status, Envelope{"error": message}, nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ServerError logs err and writes a generic 500.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("unhandled error", "err", err, "method", r.Method, "uri", r.URL.RequestURI())
	ErrorResponse(w, http.StatusInternalServerError,
		"the server encountered a problem and could not process your request")
}

// NotFound writes the standard 404 body.
func NotFound(w http.ResponseWriter) {
	ErrorResponse(w, http.StatusNotFound, "the requested resource could not be found")
}

// FailedValidation writes a 422 with the collected field errors.
func FailedValidation(w http.ResponseWriter, errs map[string]string) {
	ErrorResponse(w, http.StatusUnprocessableEntity, errs)
}

// EditConflict writes the standard 409 body.
func EditConflict(w http.ResponseWriter) {
	ErrorResponse(w, http.StatusConflict,
		"unable to update the record due to an edit conflict, please try again")
}

// BadRequest is a transport-level error carrying a client-facing message.
type BadRequest struct{ Message string }

func (e *BadRequest) Error() string { return e.Message }

// WriteBadRequest writes err as a 400 (used for JSON/body decode failures).
func WriteBadRequest(w http.ResponseWriter, err error) {
	var br *BadRequest
	if errors.As(err, &br) {
		ErrorResponse(w, http.StatusBadRequest, br.Message)
		return
	}
	ErrorResponse(w, http.StatusBadRequest, err.Error())
}

// ParamInt parses a positive path/query integer, returning def on failure —
// matching the lenient Go jsonutil.ReadInt behaviour.
func ParamInt(s string, def int) int {
	if strings.TrimSpace(s) == "" {
		return def
	}
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

// ParamInt64 is the int64 form of ParamInt.
func ParamInt64(s string, def int64) int64 {
	if strings.TrimSpace(s) == "" {
		return def
	}
	if n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
		return n
	}
	return def
}

// ParseIDList parses "1,2,3" into a slice of positive int64s, silently
// dropping invalid parts (Go jsonutil.ReadIDList behaviour).
func ParseIDList(csv string) []int64 {
	ids := []int64{}
	if strings.TrimSpace(csv) == "" {
		return ids
	}
	for _, part := range strings.Split(csv, ",") {
		if n, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil && n > 0 {
			ids = append(ids, n)
		}
	}
	return ids
}

// MethodNotAllowed writes the standard 405 body.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(w, http.StatusMethodNotAllowed,
		fmt.Sprintf("the %s method is not supported for this resource", r.Method))
}

var _ = io.EOF
