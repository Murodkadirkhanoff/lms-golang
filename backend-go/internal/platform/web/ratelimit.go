package web

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is a per-IP token-bucket limiter applied to a fixed set of
// sensitive paths (register / login / password reset) as brute-force
// protection. Mirrors the Go greenlight rateLimit middleware.
type RateLimiter struct {
	enabled bool
	rps     rate.Limit
	burst   int

	mu      sync.Mutex
	clients map[string]*client
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// limitedPaths are the only routes subject to rate limiting.
var limitedPaths = map[string]bool{
	"/v1/users":                 true,
	"/v1/tokens/authentication": true,
	"/v1/tokens/password-reset": true,
	"/v1/users/password":        true,
}

// NewRateLimiter builds a limiter and starts its background reaper.
func NewRateLimiter(enabled bool, rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		enabled: enabled,
		rps:     rate.Limit(rps),
		burst:   burst,
		clients: make(map[string]*client),
	}
	if enabled {
		go rl.reap()
	}
	return rl
}

// Middleware enforces the limit on limitedPaths, returning 429 when exceeded.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.enabled || !limitedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}
		if !rl.allow(clientIP(r)) {
			ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	c, ok := rl.clients[ip]
	if !ok {
		c = &client{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.clients[ip] = c
	}
	c.lastSeen = time.Now()
	return c.limiter.Allow()
}

func (rl *RateLimiter) reap() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP prefers the first X-Forwarded-For hop (real client behind a gateway).
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
