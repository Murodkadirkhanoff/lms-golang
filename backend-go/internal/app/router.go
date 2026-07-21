package app

import (
	"net/http"

	"github.com/chashma/lms/internal/platform/config"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// newRouter builds the HTTP handler: global middleware, cross-cutting
// endpoints (health, uploads), then each module's routes via mount.
func newRouter(cfg config.Config, tm *web.TokenMaker, rl *web.RateLimiter, up *uploads, mount func(chi.Router)) http.Handler {
	r := chi.NewRouter()

	// Outermost first: CORS (answers preflight) → rate limit → authenticate.
	r.Use(web.Recoverer)
	r.Use(web.CORS(cfg.CORSTrustedOrigins))
	r.Use(rl.Middleware)
	r.Use(tm.Authenticate)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { web.NotFound(w) })
	r.MethodNotAllowed(web.MethodNotAllowed)

	r.Get("/v1/healthcheck", healthcheck(cfg))

	// Uploaded files are served statically; uploading requires auth.
	r.Handle("/uploads/*", up.serve())
	r.With(web.RequireAuth).Post("/v1/uploads", up.upload)

	mount(r)
	return r
}
