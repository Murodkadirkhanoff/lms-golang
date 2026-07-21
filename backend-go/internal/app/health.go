package app

import (
	"net/http"

	"github.com/chashma/lms/internal/platform/config"
	"github.com/chashma/lms/internal/platform/web"
)

const version = "1.0.0"

func healthcheck(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, web.Envelope{
			"status": "available",
			"system_info": web.Envelope{
				"service":     "lms",
				"environment": cfg.Env,
				"version":     version,
			},
		}, nil)
	}
}
