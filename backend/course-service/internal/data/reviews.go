package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"lms.chashma.uz/pkg/uidefaults"
	"lms.chashma.uz/pkg/validator"
)

var ErrInvalidCourse = errors.New("course does not exist")

// Review JSON shakli frontend Review tipiga mos (types/index.ts:55):
// user — ism (snapshot), avatarColor — user_id'dan deterministik.
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

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
	v.Check(len(review.Comment) <= 2000, "comment", "must not be more than 2000 bytes long")
}

type ReviewModel struct {
	DB *sql.DB
}

// Upsert: bitta user bitta kursga bitta sharh — qayta yuborsa yangilanadi.
func (m ReviewModel) Upsert(review *Review) error {
	query := `
		INSERT INTO reviews (course_id, user_id, user_name, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (course_id, user_id)
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment,
		              user_name = EXCLUDED.user_name, created_at = NOW()
		RETURNING id, created_at
	`

	args := []any{review.CourseID, review.UserID, review.User, review.Rating, review.Comment}

	err := m.DB.QueryRow(query, args...).Scan(&review.ID, &review.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "reviews_course_id_fkey" {
			return ErrInvalidCourse
		}
		return err
	}

	review.AvatarColor = uidefaults.AvatarColor(review.UserID)

	return nil
}

func (m ReviewModel) ListForCourse(courseID int64, limit int) ([]*Review, error) {
	query := `
		SELECT id, created_at, course_id, user_id, user_name, rating, comment
		FROM reviews
		WHERE course_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`

	rows, err := m.DB.Query(query, courseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []*Review{}

	for rows.Next() {
		var r Review
		err := rows.Scan(&r.ID, &r.CreatedAt, &r.CourseID, &r.UserID, &r.User, &r.Rating, &r.Comment)
		if err != nil {
			return nil, err
		}
		r.AvatarColor = uidefaults.AvatarColor(r.UserID)
		reviews = append(reviews, &r)
	}

	return reviews, rows.Err()
}
