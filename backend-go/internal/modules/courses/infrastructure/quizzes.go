package infrastructure

import (
	"context"
	"errors"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizRepository is the pgx-backed courses.QuizRepository.
type QuizRepository struct {
	pool *pgxpool.Pool
}

// NewQuizRepository builds a QuizRepository.
func NewQuizRepository(pool *pgxpool.Pool) *QuizRepository {
	return &QuizRepository{pool: pool}
}

var _ application.QuizRepository = (*QuizRepository)(nil)

// FindByCourseID returns a course's quiz with its questions.
func (r *QuizRepository) FindByCourseID(ctx context.Context, courseID int64) (*domain.Quiz, error) {
	if courseID < 1 {
		return nil, domain.ErrNotFound
	}
	var q domain.Quiz
	err := r.pool.QueryRow(ctx, `
		SELECT id, course_id, title, passing_score, time_limit_minutes, version
		FROM course.quizzes
		WHERE course_id = $1`, courseID).
		Scan(&q.ID, &q.CourseID, &q.Title, &q.PassingScore, &q.TimeLimitMinutes, &q.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, question, options, correct_index, position
		FROM course.quiz_questions
		WHERE quiz_id = $1
		ORDER BY position, id`, q.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	q.Questions = []domain.QuizQuestion{}
	for rows.Next() {
		var qq domain.QuizQuestion
		if err := rows.Scan(&qq.ID, &qq.Question, &qq.Options, &qq.CorrectIndex, &qq.Position); err != nil {
			return nil, err
		}
		q.Questions = append(q.Questions, qq)
	}
	return &q, rows.Err()
}

// Upsert replaces a course's quiz and questions in one transaction.
func (r *QuizRepository) Upsert(ctx context.Context, q *domain.Quiz) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO course.quizzes (course_id, title, passing_score, time_limit_minutes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (course_id)
		DO UPDATE SET title = EXCLUDED.title, passing_score = EXCLUDED.passing_score,
		              time_limit_minutes = EXCLUDED.time_limit_minutes,
		              version = quizzes.version + 1
		RETURNING id, version`,
		q.CourseID, q.Title, q.PassingScore, q.TimeLimitMinutes,
	).Scan(&q.ID, &q.Version)
	if err != nil {
		if database.Constraint(err) == "quizzes_course_id_fkey" {
			return domain.ErrInvalidCourse
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM course.quiz_questions WHERE quiz_id = $1`, q.ID); err != nil {
		return err
	}
	for i := range q.Questions {
		qq := &q.Questions[i]
		if err := tx.QueryRow(ctx, `
			INSERT INTO course.quiz_questions (quiz_id, question, options, correct_index, position)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			q.ID, qq.Question, qq.Options, qq.CorrectIndex, i,
		).Scan(&qq.ID); err != nil {
			return err
		}
		qq.Position = i
	}
	return tx.Commit(ctx)
}

// InsertAttempt records a quiz attempt.
func (r *QuizRepository) InsertAttempt(ctx context.Context, a *domain.QuizAttempt) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO course.quiz_attempts (user_id, course_id, score)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, a.UserID, a.CourseID, a.Score).Scan(&a.ID, &a.CreatedAt)
}

// ListAttempts returns a user's recent attempts for a course.
func (r *QuizRepository) ListAttempts(ctx context.Context, userID, courseID int64) ([]domain.QuizAttempt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, user_id, course_id, score
		FROM course.quiz_attempts
		WHERE user_id = $1 AND course_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT 20`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.QuizAttempt{}
	for rows.Next() {
		var a domain.QuizAttempt
		if err := rows.Scan(&a.ID, &a.CreatedAt, &a.UserID, &a.CourseID, &a.Score); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AvgScoreForCourses returns the mean attempt score across courses (0 if none).
func (r *QuizRepository) AvgScoreForCourses(ctx context.Context, courseIDs []int64) (float64, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}
	var avg float64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(score), 0)::float8
		FROM course.quiz_attempts
		WHERE course_id = ANY($1)`, courseIDs).Scan(&avg)
	return avg, err
}
