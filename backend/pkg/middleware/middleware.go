// Package middleware har servisda takrorlanadigan HTTP middleware'lar:
// panic-recovery, CORS, JWT autentifikatsiya, RBAC va internal-endpoint himoyasi.
package middleware

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"lms.chashma.uz/pkg/auth"
	"lms.chashma.uz/pkg/httperr"
)

type contextKey string

const userContextKey = contextKey("user")

// ContextSetUser claims'ni request contextiga joylaydi.
func ContextSetUser(r *http.Request, claims *auth.Claims) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userContextKey, claims))
}

// ContextGetUser contextdagi claims'ni qaytaradi; yo'q bo'lsa nil (anonim).
func ContextGetUser(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(userContextKey).(*auth.Claims)
	return claims
}

func RecoverPanic(rs httperr.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				pv := recover()
				if pv != nil {
					w.Header().Set("Connection", "close")
					rs.ServerError(w, r, fmt.Errorf("%v", pv))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// EnableCORS frontend originiga ruxsat beradi va preflight so'rovlarga javob qaytaradi.
func EnableCORS(trustedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(trustedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Authenticate Bearer token bo'lsa tekshiradi va user'ni contextga qo'yadi.
// Token bo'lmasa so'rov anonim davom etadi — majburiylikni RequireAuthenticated hal qiladi.
func Authenticate(secret []byte, rs httperr.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Authorization")

			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			headerParts := strings.Split(authorizationHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				rs.InvalidAuthenticationToken(w, r)
				return
			}

			claims, err := auth.ParseToken(secret, headerParts[1])
			if err != nil {
				rs.InvalidAuthenticationToken(w, r)
				return
			}

			next.ServeHTTP(w, ContextSetUser(r, claims))
		})
	}
}

func RequireAuthenticated(rs httperr.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ContextGetUser(r) == nil {
				rs.AuthenticationRequired(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole autentifikatsiyani va ko'rsatilgan rollardan birini talab qiladi.
func RequireRole(rs httperr.Responder, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ContextGetUser(r)
			if claims == nil {
				rs.AuthenticationRequired(w, r)
				return
			}

			if !slices.Contains(roles, claims.Role) {
				rs.NotPermitted(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InternalOnly /internal/* endpointlarni servislararo umumiy kalit bilan himoyalaydi.
// Gateway bu yo'llarni tashqariga chiqarmaydi, bu esa qo'shimcha qatlam.
func InternalOnly(key string, rs httperr.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-Internal-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				rs.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
