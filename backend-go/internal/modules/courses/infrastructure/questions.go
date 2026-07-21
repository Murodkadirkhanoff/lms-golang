package infrastructure

import (
	"context"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuestionRepository is the pgx-backed courses.QuestionRepository (lesson Q&A).
type QuestionRepository struct {
	pool *pgxpool.Pool
}

// NewQuestionRepository builds a QuestionRepository.
func NewQuestionRepository(pool *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{pool: pool}
}

var _ application.QuestionRepository = (*QuestionRepository)(nil)

// List returns the most recent questions for a lesson.
func (r *QuestionRepository) List(ctx context.Context, lessonID int64) ([]domain.LessonQuestion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, user_name, question
		FROM course.lesson_questions
		WHERE lesson_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 100`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LessonQuestion{}
	for rows.Next() {
		var q domain.LessonQuestion
		if err := rows.Scan(&q.ID, &q.CreatedAt, &q.User, &q.Question); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// Insert records a question, returning domain.ErrNotFound if the lesson (FK)
// does not exist.
func (r *QuestionRepository) Insert(ctx context.Context, lessonID, userID int64, userName, question string) (domain.LessonQuestion, error) {
	q := domain.LessonQuestion{User: userName, Question: question}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course.lesson_questions (lesson_id, user_id, user_name, question)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`, lessonID, userID, userName, question).Scan(&q.ID, &q.CreatedAt)
	if err != nil {
		return domain.LessonQuestion{}, domain.ErrNotFound
	}
	return q, nil
}
