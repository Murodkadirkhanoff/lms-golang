// Package auth JWT yaratish/tekshirishni markazlashtiradi. Tokenni faqat
// auth-service yaratadi, lekin har bir servis o'zi tekshiradi (stateless).
package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

const Issuer = "lms.chashma.uz"

// Rollar users.role CHECK constraint bilan bir xil bo'lishi shart.
const (
	RoleStudent    = "student"
	RoleInstructor = "instructor"
	RoleAdmin      = "admin"
)

type Claims struct {
	UserID int64
	Role   string
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// NewToken foydalanuvchi uchun imzolangan JWT qaytaradi.
func NewToken(secret []byte, userID int64, role string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

// ParseToken tokenni tekshirib, claims qaytaradi.
func ParseToken(secret []byte, tokenString string) (*Claims, error) {
	var claims jwtClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	}, jwt.WithIssuer(Issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID < 1 {
		return nil, ErrInvalidToken
	}

	return &Claims{UserID: userID, Role: claims.Role}, nil
}
