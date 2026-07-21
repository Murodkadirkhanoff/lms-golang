package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Issuer identifies tokens minted by this system. Kept identical to the Java
// and original Go services so tokens remain cross-compatible.
const Issuer = "lms.chashma.uz"

// Identity is the authenticated principal extracted from a JWT.
type Identity struct {
	UserID int64
	Role   string
}

type contextKey string

const identityKey contextKey = "identity"

// TokenMaker issues and verifies HS256 JWTs. Pure crypto plumbing shared the
// way the stdlib is — it carries no domain rules.
type TokenMaker struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenMaker builds a TokenMaker from the signing secret and token TTL.
func NewTokenMaker(secret string, ttl time.Duration) *TokenMaker {
	return &TokenMaker{secret: []byte(secret), ttl: ttl}
}

// New mints a signed token for the given user and role.
func (t *TokenMaker) New(userID int64, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  strconv.FormatInt(userID, 10),
		"iss":  Issuer,
		"iat":  now.Unix(),
		"exp":  now.Add(t.ttl).Unix(),
		"role": role,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
}

// Parse verifies a token and returns its Identity, or an error if invalid.
func (t *TokenMaker) Parse(token string) (Identity, error) {
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return t.secret, nil
	}, jwt.WithIssuer(Issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return Identity{}, err
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return Identity{}, jwt.ErrTokenInvalidClaims
	}
	sub, _ := claims["sub"].(string)
	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil || userID < 1 {
		return Identity{}, jwt.ErrTokenInvalidClaims
	}
	role, _ := claims["role"].(string)
	return Identity{UserID: userID, Role: role}, nil
}

// Authenticate is middleware that, when a Bearer token is present, verifies it
// and stores the Identity in the request context. Missing header → anonymous;
// present-but-invalid → 401 (mirrors Go middleware.Authenticate).
func (t *TokenMaker) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			invalidToken(w)
			return
		}
		identity, err := t.Parse(parts[1])
		if err != nil {
			invalidToken(w)
			return
		}
		r = r.WithContext(WithIdentity(r.Context(), identity))
		next.ServeHTTP(w, r)
	})
}

func invalidToken(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	ErrorResponse(w, http.StatusUnauthorized, "invalid or missing authentication token")
}

// WithIdentity returns a child context carrying id.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFrom extracts the Identity from ctx, reporting whether one is present.
func IdentityFrom(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}

// RequireAuth is middleware that rejects anonymous requests with 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := IdentityFrom(r.Context()); !ok {
			ErrorResponse(w, http.StatusUnauthorized,
				"you must be authenticated to access this resource")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole is middleware that requires an authenticated user with role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := IdentityFrom(r.Context())
			if !ok {
				ErrorResponse(w, http.StatusUnauthorized,
					"you must be authenticated to access this resource")
				return
			}
			if id.Role != role {
				ErrorResponse(w, http.StatusForbidden,
					"your user account doesn't have the necessary permissions to access this resource")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
