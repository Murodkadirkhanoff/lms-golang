package domain

import "time"

// Category is a course category (frontend Category type). createdAt/version
// are internal and excluded from JSON.
type Category struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Slug        string    `json:"slug"`
	NameUz      string    `json:"nameUz"`
	NameRu      string    `json:"nameRu"`
	NameEn      string    `json:"nameEn"`
	ParentID    *int64    `json:"parentId"`
	CourseCount int       `json:"courseCount"`
	Version     int       `json:"-"`
}

// Quiz is a course quiz (grading happens client-side, so correctIndex is
// intentionally part of the response, matching the original services).
type Quiz struct {
	ID               int64          `json:"id"`
	CourseID         int64          `json:"-"`
	Title            string         `json:"title"`
	PassingScore     int            `json:"passingScore"`
	TimeLimitMinutes int            `json:"timeLimitMinutes"`
	Questions        []QuizQuestion `json:"questions"`
	Version          int            `json:"-"`
}

// QuizQuestion is one multiple-choice question.
type QuizQuestion struct {
	ID           int64    `json:"id"`
	Question     string   `json:"question"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
	Position     int      `json:"-"`
}

// QuizAttempt is a recorded quiz score.
type QuizAttempt struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"-"`
	CourseID  int64     `json:"-"`
	Score     int       `json:"score"`
}

// LessonQuestion is a learner question on the Learn page Q&A tab.
type LessonQuestion struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	User      string    `json:"user"`
	Question  string    `json:"question"`
}

// InstructorStat aggregates a single instructor's published-course metrics.
type InstructorStat struct {
	InstructorID int64
	CourseCount  int
	Students     int
	Rating       float64
}
