package infrastructure

import (
	"context"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReviewRepository is the pgx-backed courses.ReviewRepository.
type ReviewRepository struct {
	pool *pgxpool.Pool
}

// NewReviewRepository builds a ReviewRepository.
func NewReviewRepository(pool *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

var _ application.ReviewRepository = (*ReviewRepository)(nil)

// Upsert creates or updates a user's review for a course.
func (r *ReviewRepository) Upsert(ctx context.Context, rv *contract.Review) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course.reviews (course_id, user_id, user_name, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (course_id, user_id)
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment,
		              user_name = EXCLUDED.user_name, created_at = NOW()
		RETURNING id, created_at`,
		rv.CourseID, rv.UserID, rv.User, rv.Rating, rv.Comment,
	).Scan(&rv.ID, &rv.CreatedAt)
	if err != nil {
		if database.Constraint(err) == "reviews_course_id_fkey" {
			return domain.ErrInvalidCourse
		}
		return err
	}
	rv.AvatarColor = domain.AvatarColor(rv.UserID)
	return nil
}

// ListForCourse returns the latest reviews for a course.
func (r *ReviewRepository) ListForCourse(ctx context.Context, courseID int64, limit int) ([]contract.Review, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, course_id, user_id, user_name, rating, comment
		FROM course.reviews
		WHERE course_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2`, courseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []contract.Review{}
	for rows.Next() {
		var rv contract.Review
		if err := rows.Scan(&rv.ID, &rv.CreatedAt, &rv.CourseID, &rv.UserID, &rv.User, &rv.Rating, &rv.Comment); err != nil {
			return nil, err
		}
		rv.AvatarColor = domain.AvatarColor(rv.UserID)
		out = append(out, rv)
	}
	return out, rows.Err()
}
