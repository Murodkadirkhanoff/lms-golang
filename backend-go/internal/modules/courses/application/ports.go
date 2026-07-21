// Package application holds the courses use-case service and its ports.
package application

import (
	"context"

	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
)

// CourseFilters describes a course listing query.
type CourseFilters struct {
	Search             string
	CategorySlug       string
	Sort               string
	Page               int
	PageSize           int
	IDs                []int64
	InstructorID       int64
	IncludeUnpublished bool
}

// CourseRepository is the persistence port for courses/modules/lessons.
type CourseRepository interface {
	List(ctx context.Context, f CourseFilters) ([]contract.CourseView, int, error)
	FindByIDOrSlug(ctx context.Context, idOrSlug string) (*contract.CourseView, error)
	Insert(ctx context.Context, c *contract.CourseView) error // ErrDuplicateSlug / ErrInvalidParent
	Update(ctx context.Context, c *contract.CourseView) error // ErrEditConflict / ErrInvalidParent
	ReplaceModules(ctx context.Context, courseID int64, modules []contract.Module) error
	Delete(ctx context.Context, id int64) error
	LessonsForCourses(ctx context.Context, courseIDs []int64) ([]contract.LessonInfo, error)
	LessonsByIDs(ctx context.Context, ids []int64) ([]contract.LessonInfo, error)
	IncrementStudentCount(ctx context.Context, courseID int64) error
	Stats(ctx context.Context) (contract.CourseStats, error)
	CourseCountsByInstructor(ctx context.Context, ids []int64) (map[int64]int, error)
	InstructorStats(ctx context.Context) ([]domain.InstructorStat, error)
}

// CategoryRepository is the persistence port for categories.
type CategoryRepository interface {
	Insert(ctx context.Context, c *domain.Category) error // ErrDuplicateSlug / ErrInvalidParent / ErrMaxDepth
	FindByID(ctx context.Context, id int64) (*domain.Category, error)
	List(ctx context.Context) ([]domain.Category, error)
	Update(ctx context.Context, c *domain.Category) error
	Delete(ctx context.Context, id int64) error
}

// QuizRepository is the persistence port for quizzes and attempts.
type QuizRepository interface {
	FindByCourseID(ctx context.Context, courseID int64) (*domain.Quiz, error)
	Upsert(ctx context.Context, q *domain.Quiz) error // ErrInvalidCourse
	InsertAttempt(ctx context.Context, a *domain.QuizAttempt) error
	ListAttempts(ctx context.Context, userID, courseID int64) ([]domain.QuizAttempt, error)
	AvgScoreForCourses(ctx context.Context, courseIDs []int64) (float64, error)
}

// ReviewRepository is the persistence port for reviews.
type ReviewRepository interface {
	Upsert(ctx context.Context, r *contract.Review) error // ErrInvalidCourse
	ListForCourse(ctx context.Context, courseID int64, limit int) ([]contract.Review, error)
}

// QuestionRepository is the persistence port for lesson Q&A.
type QuestionRepository interface {
	List(ctx context.Context, lessonID int64) ([]domain.LessonQuestion, error)
	Insert(ctx context.Context, lessonID, userID int64, userName, question string) (domain.LessonQuestion, error) // ErrNotFound on bad lesson
}
