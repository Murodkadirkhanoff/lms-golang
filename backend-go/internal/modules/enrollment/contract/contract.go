// Package contract is the enrollment bounded context's public port. Other
// modules depend ONLY on this package.
package contract

import "context"

// EnrollmentGate exposes enrollment/commerce facts to other contexts:
// review gating, paywall lesson access, and admin aggregates.
type EnrollmentGate interface {
	// Revenue is the sum of all paid orders (admin stats).
	Revenue(ctx context.Context) (float64, error)
	// EnrollmentCountsByUser maps user id -> number of enrollments.
	EnrollmentCountsByUser(ctx context.Context, userIDs []int64) (map[int64]int, error)
	// AccessibleLessonIDs lists the lesson ids a user may view in a course.
	AccessibleLessonIDs(ctx context.Context, userID, courseID int64) ([]int64, error)
	// IsEnrolled reports whether the user is enrolled in the course.
	IsEnrolled(ctx context.Context, userID, courseID int64) (bool, error)
}
