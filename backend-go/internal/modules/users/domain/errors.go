package domain

import "errors"

// Domain errors owned by the users context. Transport maps these to HTTP.
var (
	ErrNotFound           = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrEditConflict       = errors.New("edit conflict")
	ErrInvalidCredentials = errors.New("invalid authentication credentials")
)

// AvatarColor returns a deterministic Tailwind class for a user id. Duplicated
// per module on purpose: it is UI-default data with no shared owner, and each
// context must stand alone. The palette matches the frontend.
func AvatarColor(id int64) string {
	return palette[int(id%int64(len(palette)))]
}

var palette = []string{
	"bg-indigo-200",
	"bg-amber-200",
	"bg-rose-200",
	"bg-emerald-200",
	"bg-sky-200",
	"bg-fuchsia-200",
}
