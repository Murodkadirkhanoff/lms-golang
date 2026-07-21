package web

import "regexp"

// EmailRX matches the same addresses as the Java Validator.EMAIL_RX so
// validation behaviour is identical across backends.
var EmailRX = regexp.MustCompile(
	"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator collects field errors keyed by field name. It is a generic
// collector only — the actual rules live in each module's domain/transport.
// Mirrors the Go pkg/validator used by the original services.
type Validator struct {
	Errors map[string]string
}

// NewValidator returns an empty validator.
func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid reports whether no errors have been recorded.
func (v *Validator) Valid() bool { return len(v.Errors) == 0 }

// AddError records message for key unless key already has an error.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check records message for key when ok is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// ByteLength returns the UTF-8 byte length of s (Go len semantics), matching
// the byte-based length limits the frontend expects.
func ByteLength(s string) int { return len(s) }

// Matches reports whether s matches rx.
func Matches(s string, rx *regexp.Regexp) bool { return rx.MatchString(s) }

// Permitted reports whether value is one of allowed.
func Permitted(value string, allowed ...string) bool {
	for _, a := range allowed {
		if a == value {
			return true
		}
	}
	return false
}
