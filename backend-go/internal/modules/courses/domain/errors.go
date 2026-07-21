// Package domain holds the courses bounded context's rules, errors and
// value helpers. No framework or cross-module dependencies beyond the shared
// validator collector and this module's own leaf contract types.
package domain

import "errors"

// Domain errors owned by the courses context.
var (
	ErrNotFound      = errors.New("course resource not found")
	ErrEditConflict  = errors.New("edit conflict")
	ErrDuplicateSlug = errors.New("duplicate slug")
	ErrInvalidParent = errors.New("invalid parent")
	ErrMaxDepth      = errors.New("category nesting too deep")
	ErrInvalidCourse = errors.New("invalid course reference")
)

// Role constant needed for the "owner or admin" authorisation check. Kept
// local to the module (autonomy) — its value matches the auth role vocabulary.
const RoleAdmin = "admin"

var palette = []string{
	"bg-indigo-200",
	"bg-amber-200",
	"bg-rose-200",
	"bg-emerald-200",
	"bg-sky-200",
	"bg-fuchsia-200",
}

// AvatarColor is the deterministic avatar class for a user id.
func AvatarColor(id int64) string {
	return palette[int(id%int64(len(palette)))]
}

// ThumbnailColor is the deterministic course-thumbnail class for a course id.
func ThumbnailColor(id int64) string {
	return palette[int((id+3)%int64(len(palette)))]
}
