package application

import (
	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
)

// EnrolledCourseView is one row of the learner's course library.
type EnrolledCourseView struct {
	EnrollmentID       int64                      `json:"enrollmentId"`
	Course             coursescontract.CourseView `json:"course"`
	Progress           int                        `json:"progress"`
	CurrentLesson      string                     `json:"currentLesson"`
	LessonsCompleted   int                        `json:"lessonsCompleted"`
	CompletedLessonIDs []int64                    `json:"completedLessonIds"`
}

// DashboardStats is the learner dashboard summary.
type DashboardStats struct {
	Enrolled     int `json:"enrolled"`
	InProgress   int `json:"inProgress"`
	Completed    int `json:"completed"`
	Certificates int `json:"certificates"`
}

// CourseEngagement is one course's completion figure for the studio analytics.
type CourseEngagement struct {
	CourseID   int64  `json:"courseId"`
	Title      string `json:"title"`
	Students   int    `json:"students"`
	Completion int    `json:"completion"`
}

// TeachingStats is the instructor studio dashboard summary.
type TeachingStats struct {
	TotalRevenue     float64               `json:"totalRevenue"`
	MonthlyRevenue   []domain.MonthRevenue `json:"monthlyRevenue"`
	TotalStudents    int                   `json:"totalStudents"`
	ActiveStudents   int                   `json:"activeStudents"`
	PublishedCourses int                   `json:"publishedCourses"`
	DraftCourses     int                   `json:"draftCourses"`
	AvgRating        float64               `json:"avgRating"`
	AvgCompletion    int                   `json:"avgCompletion"`
	AvgQuizScore     float64               `json:"avgQuizScore"`
	Engagement       []CourseEngagement    `json:"engagement"`
}
