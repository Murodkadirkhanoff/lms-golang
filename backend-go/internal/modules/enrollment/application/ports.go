// Package application holds the enrollment use-case service and its ports.
package application

import (
	"context"

	"github.com/chashma/lms/internal/modules/enrollment/domain"
)

// CompletedStats aggregates completed-lesson records for a set of courses.
type CompletedStats struct {
	Counts         map[int64]int
	ActiveStudents int
}

// RevenueResult is total revenue plus a per-month breakdown.
type RevenueResult struct {
	Total   float64
	Monthly []domain.MonthRevenue
}

// EnrollmentRepository is the persistence port for enrollments and lesson access.
type EnrollmentRepository interface {
	Insert(ctx context.Context, userID, courseID int64) (domain.Enrollment, bool, error) // bool = isNew
	FindByID(ctx context.Context, id int64) (*domain.Enrollment, error)
	ListByUser(ctx context.Context, userID int64) ([]domain.Enrollment, error)
	PageByUser(ctx context.Context, userID int64, page, pageSize int) ([]domain.Enrollment, int, error)
	GrantLessonAccess(ctx context.Context, userID, courseID int64, lessonIDs []int64) error
	SetLessonCompleted(ctx context.Context, userID, lessonID int64, completed bool) error // domain.ErrNotFound
	CompletedCounts(ctx context.Context, userID int64) (map[int64]int, error)
	CompletedLessonIDs(ctx context.Context, userID int64) (map[int64]bool, error)
	OwnedCourses(ctx context.Context, userID int64, courseIDs []int64) (map[int64]bool, error)
	OwnedLessons(ctx context.Context, userID int64, lessonIDs []int64) (map[int64]bool, error)
	AccessibleLessonIDs(ctx context.Context, userID, courseID int64) ([]int64, error)
	CountsByCourses(ctx context.Context, courseIDs []int64) (map[int64]int, error)
	DistinctStudentsForCourses(ctx context.Context, courseIDs []int64) (int, error)
	CompletedStatsByCourses(ctx context.Context, courseIDs []int64) (CompletedStats, error)
	EnrollmentCountsByUser(ctx context.Context, userIDs []int64) (map[int64]int, error)
}

// OrderRepository is the persistence port for orders.
type OrderRepository interface {
	Insert(ctx context.Context, o *domain.Order) error
	FindForUser(ctx context.Context, id, userID int64) (*domain.Order, error)
	ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]domain.Order, int, error)
	Revenue(ctx context.Context) (float64, error)
	RevenueForItems(ctx context.Context, courseIDs, lessonIDs []int64) (RevenueResult, error)
}

// CertificateRepository is the persistence port for certificates.
type CertificateRepository interface {
	Issue(ctx context.Context, userID, courseID int64, courseTitle string) (bool, error) // bool = newly issued
	ListByUser(ctx context.Context, userID int64) ([]domain.Certificate, error)
	FindForUser(ctx context.Context, id, userID int64) (*domain.Certificate, error)
	CountByUser(ctx context.Context, userID int64) (int, error)
}

// CertificateRenderer renders a certificate to a PDF document.
type CertificateRenderer interface {
	Render(studentName string, cert domain.Certificate) ([]byte, error)
}

// NotificationRepository is the persistence port for notifications.
type NotificationRepository interface {
	Insert(ctx context.Context, userID int64, ntype, title, body string) error
	ListByUser(ctx context.Context, userID int64) ([]domain.Notification, error)
	MarkAllRead(ctx context.Context, userID int64) error
	MarkRead(ctx context.Context, id, userID int64) error // domain.ErrNotFound
}
