package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// QuizAttempt — foydalanuvchining bitta urinish natijasi. Baholash clientda
// bo'lgani uchun score tayyor holda keladi (correctIndex baribir ochiq).
type QuizAttempt struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"-"`
	CourseID  int64     `json:"-"`
	Score     int       `json:"score"`
}

func (m QuizModel) InsertAttempt(attempt *QuizAttempt) error {
	query := `
		INSERT INTO quiz_attempts (user_id, course_id, score)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return m.DB.QueryRow(query, attempt.UserID, attempt.CourseID, attempt.Score).
		Scan(&attempt.ID, &attempt.CreatedAt)
}

func (m QuizModel) ListAttempts(userID, courseID int64) ([]*QuizAttempt, error) {
	query := `
		SELECT id, created_at, user_id, course_id, score
		FROM quiz_attempts
		WHERE user_id = $1 AND course_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT 20
	`

	rows, err := m.DB.Query(query, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attempts := []*QuizAttempt{}
	for rows.Next() {
		var a QuizAttempt
		err := rows.Scan(&a.ID, &a.CreatedAt, &a.UserID, &a.CourseID, &a.Score)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, &a)
	}

	return attempts, rows.Err()
}

// AvgScoreForCourses instruktor analitikasi uchun: ko'rsatilgan kurslardagi
// barcha urinishlarning o'rtacha bali (urinish bo'lmasa 0).
func (m QuizModel) AvgScoreForCourses(courseIDs []int64) (float64, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}

	query := `
		SELECT COALESCE(AVG(score), 0)
		FROM quiz_attempts
		WHERE course_id = ANY($1)
	`

	var avg float64
	err := m.DB.QueryRow(query, pq.Array(courseIDs)).Scan(&avg)
	return avg, err
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
