// Package domain holds the users bounded context's entities and rules.
// It has no framework or cross-module dependencies.
package domain

import (
	"time"

	"github.com/chashma/lms/internal/platform/web"
)

// Roles recognised by the auth context (must match the auth.users CHECK).
const (
	RoleStudent    = "student"
	RoleInstructor = "instructor"
	RoleAdmin      = "admin"
)

// User is the auth.users aggregate.
type User struct {
	ID           int64
	CreatedAt    time.Time
	Name         string
	Email        string
	PasswordHash []byte
	Role         string
	Version      int
}

// ValidateName checks the display name (messages match the frontend).
func ValidateName(v *web.Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.Check(web.ByteLength(name) <= 500, "name", "must not be more than 500 bytes long")
}

// ValidateEmail checks the email address.
func ValidateEmail(v *web.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(web.Matches(email, web.EmailRX), "email", "must be a valid email address")
}

// ValidatePassword checks the raw password length (byte-based, like bcrypt).
func ValidatePassword(v *web.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(web.ByteLength(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(web.ByteLength(password) <= 72, "password", "must not be more than 72 bytes long")
}
