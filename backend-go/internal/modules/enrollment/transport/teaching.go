package transport

import (
	"net/http"

	"github.com/chashma/lms/internal/platform/web"
)

// teachingStats returns the instructor studio dashboard.
func (h *Handler) teachingStats(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	stats, err := h.svc.TeachingStats(r.Context(), identity.UserID)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, stats, nil)
}
