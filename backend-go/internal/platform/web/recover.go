package web

import (
	"log/slog"
	"net/http"
)

// Recoverer converts panics into a JSON 500 (and closes the connection so a
// half-written response is not sent).
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Connection", "close")
				slog.Error("recovered panic", "err", rec, "uri", r.URL.RequestURI())
				ErrorResponse(w, http.StatusInternalServerError,
					"the server encountered a problem and could not process your request")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
