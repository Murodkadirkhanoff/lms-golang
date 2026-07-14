package data

import (
	"database/sql"
	"time"

	"lms.chashma.uz/pkg/uidefaults"
)

// Certificate JSON shakli frontend Certificate tipiga mos (types/index.ts:92).
type Certificate struct {
	ID          int64     `json:"id"`
	IssuedAt    time.Time `json:"issuedAt"`
	UserID      int64     `json:"-"`
	CourseID    int64     `json:"-"`
	CourseTitle string    `json:"courseTitle"`
	Color       string    `json:"color"`
}

type CertificateModel struct {
	DB *sql.DB
}

// Issue idempotent: allaqachon berilgan bo'lsa yangisini yaratmaydi.
// Yangi berilgan bo'lsa true qaytaradi.
func (m CertificateModel) Issue(userID, courseID int64, courseTitle string) (bool, error) {
	query := `
		INSERT INTO certificates (user_id, course_id, course_title)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, course_id) DO NOTHING
		RETURNING id
	`

	var id int64
	err := m.DB.QueryRow(query, userID, courseID, courseTitle).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (m CertificateModel) ListByUser(userID int64) ([]*Certificate, error) {
	query := `
		SELECT id, issued_at, user_id, course_id, course_title
		FROM certificates
		WHERE user_id = $1
		ORDER BY issued_at DESC, id DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	certificates := []*Certificate{}

	for rows.Next() {
		var c Certificate
		err := rows.Scan(&c.ID, &c.IssuedAt, &c.UserID, &c.CourseID, &c.CourseTitle)
		if err != nil {
			return nil, err
		}
		c.Color = uidefaults.ThumbnailColor(c.CourseID)
		certificates = append(certificates, &c)
	}

	return certificates, rows.Err()
}

func (m CertificateModel) CountByUser(userID int64) (int, error) {
	var count int
	err := m.DB.QueryRow(`SELECT count(*) FROM certificates WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}
