package data

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"lms.chashma.uz/pkg/validator"
)

// Quiz JSON shakli frontend Quiz tipiga mos (types/index.ts:106).
// Frontend quizni kurs id bilan so'raydi (ROUTES.quiz(course.id)) va
// baholashni clientda qiladi, shuning uchun correctIndex javobga kiradi.
type Quiz struct {
	ID               int64           `json:"id"`
	CourseID         int64           `json:"-"`
	Title            string          `json:"title"`
	PassingScore     int             `json:"passingScore"`
	TimeLimitMinutes int             `json:"timeLimitMinutes"`
	Questions        []*QuizQuestion `json:"questions"`
	Version          int             `json:"-"`
}

type QuizQuestion struct {
	ID           int64    `json:"id"`
	Question     string   `json:"question"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
	Position     int      `json:"-"`
}

func ValidateQuiz(v *validator.Validator, quiz *Quiz) {
	v.Check(quiz.Title != "", "title", "must be provided")
	v.Check(quiz.PassingScore >= 0 && quiz.PassingScore <= 100, "passing_score", "must be between 0 and 100")
	v.Check(quiz.TimeLimitMinutes > 0, "time_limit_minutes", "must be greater than zero")
	v.Check(len(quiz.Questions) > 0, "questions", "must contain at least one question")

	for i, q := range quiz.Questions {
		key := fmt.Sprintf("questions[%d]", i)
		v.Check(q.Question != "", key+".question", "must be provided")
		v.Check(len(q.Options) >= 2, key+".options", "must have at least 2 options")
		v.Check(q.CorrectIndex >= 0 && q.CorrectIndex < len(q.Options), key+".correct_index", "must point to an option")
	}
}

type QuizModel struct {
	DB *sql.DB
}

func (m QuizModel) GetByCourseID(courseID int64) (*Quiz, error) {
	if courseID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, course_id, title, passing_score, time_limit_minutes, version
		FROM quizzes
		WHERE course_id = $1
	`

	var quiz Quiz

	err := m.DB.QueryRow(query, courseID).Scan(
		&quiz.ID,
		&quiz.CourseID,
		&quiz.Title,
		&quiz.PassingScore,
		&quiz.TimeLimitMinutes,
		&quiz.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	questionQuery := `
		SELECT id, question, options, correct_index, position
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY position, id
	`

	rows, err := m.DB.Query(questionQuery, quiz.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quiz.Questions = []*QuizQuestion{}

	for rows.Next() {
		var q QuizQuestion
		err := rows.Scan(&q.ID, &q.Question, pq.Array(&q.Options), &q.CorrectIndex, &q.Position)
		if err != nil {
			return nil, err
		}
		quiz.Questions = append(quiz.Questions, &q)
	}

	return &quiz, rows.Err()
}

// Upsert kursning quizini butunlay almashtiradi (savollar bilan birga).
func (m QuizModel) Upsert(quiz *Quiz) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO quizzes (course_id, title, passing_score, time_limit_minutes)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (course_id)
		 DO UPDATE SET title = EXCLUDED.title, passing_score = EXCLUDED.passing_score,
		               time_limit_minutes = EXCLUDED.time_limit_minutes, version = quizzes.version + 1
		 RETURNING id, version`,
		quiz.CourseID, quiz.Title, quiz.PassingScore, quiz.TimeLimitMinutes,
	).Scan(&quiz.ID, &quiz.Version)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "quizzes_course_id_fkey" {
			return ErrInvalidCourse
		}
		return err
	}

	_, err = tx.Exec(`DELETE FROM quiz_questions WHERE quiz_id = $1`, quiz.ID)
	if err != nil {
		return err
	}

	for i, q := range quiz.Questions {
		err = tx.QueryRow(
			`INSERT INTO quiz_questions (quiz_id, question, options, correct_index, position)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			quiz.ID, q.Question, pq.Array(q.Options), q.CorrectIndex, i,
		).Scan(&q.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
