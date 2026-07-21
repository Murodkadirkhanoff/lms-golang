package web

import "net/http"

// CORS returns middleware that allows the configured trusted origins with the
// method/header set the frontend needs. Mirrors the Java CorsConfigurationSource.
func CORS(trustedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(trustedOrigins))
	for _, o := range trustedOrigins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")

			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
