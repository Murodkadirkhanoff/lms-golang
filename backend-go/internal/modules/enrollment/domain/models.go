// Package domain holds the enrollment bounded context's entities and rules.
package domain

import "time"

// Enrollment is a user's enrollment in a course (frontend Enrollment type).
type Enrollment struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"userId"`
	CourseID  int64     `json:"courseId"`
}

// OrderItem is one purchased course or lesson (title/instructor/color are
// snapshots captured at purchase time).
type OrderItem struct {
	CourseID       *int64
	LessonID       *int64
	CourseTitle    string
	Instructor     string
	ThumbnailColor string
	Price          float64
}

// Order is a purchase. Status is the internal lifecycle value; the frontend
// sees the mapped PublicStatus.
type Order struct {
	DBID          int64
	CreatedAt     time.Time
	UserID        int64
	Status        string
	PaymentMethod string
	Items         []OrderItem
	Total         float64
}

// PublicStatus maps the internal status to the frontend vocabulary.
func (o Order) PublicStatus() string {
	switch o.Status {
	case "paid":
		return "completed"
	case "failed", "cancelled":
		return "refunded"
	default:
		return "pending"
	}
}

// Certificate is a completion certificate (course_title is a snapshot).
type Certificate struct {
	ID          int64     `json:"id"`
	IssuedAt    time.Time `json:"issuedAt"`
	UserID      int64     `json:"-"`
	CourseID    int64     `json:"-"`
	CourseTitle string    `json:"courseTitle"`
	Color       string    `json:"color"`
}

// Notification is an in-app notification.
type Notification struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"-"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
}

// MonthRevenue is a single month's revenue bucket.
type MonthRevenue struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
}

var palette = []string{
	"bg-indigo-200",
	"bg-amber-200",
	"bg-rose-200",
	"bg-emerald-200",
	"bg-sky-200",
	"bg-fuchsia-200",
}

// ThumbnailColor is the deterministic thumbnail class for a course id.
func ThumbnailColor(id int64) string {
	return palette[int((id+3)%int64(len(palette)))]
}
