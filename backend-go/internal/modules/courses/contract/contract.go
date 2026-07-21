// Package contract is the courses bounded context's public port. Other modules
// depend ONLY on this package. The CourseView / Lesson / Module / Review /
// Instructor structs are the canonical course wire model: whether a course is
// returned by the courses module directly or embedded by enrollment, the JSON
// is identical because it is produced from these types.
package contract

import (
	"context"
	"time"
)

// Instructor is the decorated instructor projection embedded in a course.
type Instructor struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Headline    string  `json:"headline"`
	AvatarColor string  `json:"avatarColor"`
	Students    int     `json:"students"`
	Courses     int     `json:"courses"`
	Rating      float64 `json:"rating"`
}

// Lesson is a single lesson in a course's curriculum.
type Lesson struct {
	ID              int64   `json:"id"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	ContentURL      string  `json:"contentUrl,omitempty"`
	Content         string  `json:"content,omitempty"`
	DurationSeconds int     `json:"durationSeconds"`
	Position        int     `json:"-"`
	Price           float64 `json:"price"`
	IsFree          bool    `json:"isFree"`
	Locked          *bool   `json:"locked,omitempty"`
}

// Module groups lessons within a course.
type Module struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Position int      `json:"-"`
	Lessons  []Lesson `json:"lessons"`
}

// Review is a course review (user is a display-name snapshot).
type Review struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	CourseID    int64     `json:"-"`
	UserID      int64     `json:"-"`
	User        string    `json:"user"`
	AvatarColor string    `json:"avatarColor"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
}

// CourseView is the canonical course wire object (frontend Course type).
type CourseView struct {
	ID                   int64       `json:"id"`
	CreatedAt            time.Time   `json:"createdAt"`
	Slug                 string      `json:"slug"`
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	ThumbnailColor       string      `json:"thumbnailColor"`
	ThumbnailURL         string      `json:"thumbnailUrl"`
	CategoryID           *int64      `json:"categoryId"`
	Category             string      `json:"category"`
	Lang                 string      `json:"lang"`
	Price                float64     `json:"price"`
	Rating               float64     `json:"rating"`
	RatingCount          int         `json:"ratingCount"`
	StudentCount         int         `json:"studentCount"`
	IsPublished          bool        `json:"isPublished"`
	Instructor           *Instructor `json:"instructor"`
	Modules              []Module    `json:"modules,omitempty"`
	Reviews              []Review    `json:"reviews,omitempty"`
	TotalLessons         int         `json:"totalLessons"`
	TotalDurationMinutes int         `json:"totalDurationMinutes"`

	InstructorID int64 `json:"-"`
	Version      int   `json:"-"`
}

// LessonInfo is a lightweight lesson projection used for pricing, access
// grants and progress ("current lesson").
type LessonInfo struct {
	ID          int64
	Title       string
	Price       float64
	IsFree      bool
	CourseID    int64
	CourseTitle string
}

// CourseStats is the admin dashboard course aggregate.
type CourseStats struct {
	TotalCourses      int
	ActiveInstructors int
}

// CourseCatalog exposes course data to other bounded contexts.
type CourseCatalog interface {
	CoursesByIDs(ctx context.Context, ids []int64) ([]CourseView, error)
	CoursesByInstructor(ctx context.Context, instructorID int64) ([]CourseView, error)
	LessonsForCourse(ctx context.Context, courseID int64) ([]LessonInfo, error)
	LessonsForCourses(ctx context.Context, courseIDs []int64) ([]LessonInfo, error)
	LessonsByIDs(ctx context.Context, ids []int64) ([]LessonInfo, error)
	IncrementStudentCount(ctx context.Context, courseID int64) error
	AvgQuizScore(ctx context.Context, courseIDs []int64) (float64, error)
	Stats(ctx context.Context) (CourseStats, error)
	CourseCountsByInstructor(ctx context.Context, userIDs []int64) (map[int64]int, error)
}
