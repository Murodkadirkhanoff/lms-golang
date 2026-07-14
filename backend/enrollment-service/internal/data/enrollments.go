package data

import (
	"database/sql"
	"errors"
	"time"
)

var ErrRecordNotFound = errors.New("record not found")

type Enrollment struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"userId"`
	CourseID  int64     `json:"courseId"`
}

type EnrollmentModel struct {
	DB *sql.DB
}

// Insert idempotent: allaqachon yozilgan bo'lsa mavjud yozuvni qaytaradi.
// isNew — yozuv aynan hozir yaratilganini bildiradi (student_count va
// notification faqat yangi enrollment uchun).
func (m EnrollmentModel) Insert(userID, courseID int64) (enrollment *Enrollment, isNew bool, err error) {
	enrollment = &Enrollment{UserID: userID, CourseID: courseID}

	query := `
		INSERT INTO enrollments (user_id, course_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, course_id) DO NOTHING
		RETURNING id, created_at
	`

	err = m.DB.QueryRow(query, userID, courseID).Scan(&enrollment.ID, &enrollment.CreatedAt)
	if err == nil {
		return enrollment, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	// Konflikt — mavjud yozuvni o'qiymiz.
	existing := `SELECT id, created_at FROM enrollments WHERE user_id = $1 AND course_id = $2`
	err = m.DB.QueryRow(existing, userID, courseID).Scan(&enrollment.ID, &enrollment.CreatedAt)
	if err != nil {
		return nil, false, err
	}

	return enrollment, false, nil
}

func (m EnrollmentModel) Get(id int64) (*Enrollment, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, created_at, user_id, course_id FROM enrollments WHERE id = $1`

	var enrollment Enrollment

	err := m.DB.QueryRow(query, id).Scan(
		&enrollment.ID,
		&enrollment.CreatedAt,
		&enrollment.UserID,
		&enrollment.CourseID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &enrollment, nil
}

func (m EnrollmentModel) ListByUser(userID int64) ([]*Enrollment, error) {
	query := `
		SELECT id, created_at, user_id, course_id
		FROM enrollments
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enrollments := []*Enrollment{}

	for rows.Next() {
		var e Enrollment
		err := rows.Scan(&e.ID, &e.CreatedAt, &e.UserID, &e.CourseID)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, &e)
	}

	return enrollments, rows.Err()
}

// GrantLessonAccess darslarga kirishni beradi (kurs sotib olinganda barcha
// darslar, alohida dars sotib olinganda bittasi). Eski DB-trigger o'rnini bosadi.
func (m EnrollmentModel) GrantLessonAccess(userID, courseID int64, lessonIDs []int64) error {
	query := `
		INSERT INTO lesson_access (user_id, lesson_id, course_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, lesson_id) DO NOTHING
	`

	for _, lessonID := range lessonIDs {
		_, err := m.DB.Exec(query, userID, lessonID, courseID)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetLessonCompleted dars tugatildi/tugatilmadi belgisini qo'yadi.
func (m EnrollmentModel) SetLessonCompleted(userID, lessonID int64, completed bool) error {
	query := `
		UPDATE lesson_access
		SET completed_at = CASE WHEN $3 THEN NOW() ELSE NULL END
		WHERE user_id = $1 AND lesson_id = $2
	`

	result, err := m.DB.Exec(query, userID, lessonID, completed)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// CompletedCounts foydalanuvchining har bir kursi bo'yicha tugatilgan
// darslar sonini qaytaradi.
func (m EnrollmentModel) CompletedCounts(userID int64) (map[int64]int, error) {
	query := `
		SELECT course_id, count(*)
		FROM lesson_access
		WHERE user_id = $1 AND completed_at IS NOT NULL
		GROUP BY course_id
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[int64]int{}

	for rows.Next() {
		var courseID int64
		var count int
		err := rows.Scan(&courseID, &count)
		if err != nil {
			return nil, err
		}
		counts[courseID] = count
	}

	return counts, rows.Err()
}

// CompletedLessonIDs kurs bo'yicha tugatilgan dars id'lari (currentLesson uchun).
func (m EnrollmentModel) CompletedLessonIDs(userID int64) (map[int64]bool, error) {
	query := `
		SELECT lesson_id
		FROM lesson_access
		WHERE user_id = $1 AND completed_at IS NOT NULL
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	done := map[int64]bool{}

	for rows.Next() {
		var lessonID int64
		if err := rows.Scan(&lessonID); err != nil {
			return nil, err
		}
		done[lessonID] = true
	}

	return done, rows.Err()
}

// CountsByCourses teaching stats uchun: kurs -> yozilgan studentlar soni.
func (m EnrollmentModel) CountsByCourses(courseIDs []int64) (map[int64]int, error) {
	counts := map[int64]int{}
	if len(courseIDs) == 0 {
		return counts, nil
	}

	query := `
		SELECT course_id, count(*)
		FROM enrollments
		WHERE course_id = ANY($1)
		GROUP BY course_id
	`

	rows, err := m.DB.Query(query, int64Array(courseIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		counts[id] = count
	}

	return counts, rows.Err()
}

// DistinctStudentsForCourses kurslarga yozilgan noyob foydalanuvchilar soni.
func (m EnrollmentModel) DistinctStudentsForCourses(courseIDs []int64) (int, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}

	query := `SELECT count(DISTINCT user_id) FROM enrollments WHERE course_id = ANY($1)`

	var n int
	err := m.DB.QueryRow(query, int64Array(courseIDs)).Scan(&n)
	return n, err
}

// CompletedStatsByCourses teaching stats uchun: kurs -> tugatilgan darslar
// yozuvlari soni (barcha studentlar bo'yicha) va kamida bitta dars tugatgan
// noyob studentlar soni.
func (m EnrollmentModel) CompletedStatsByCourses(courseIDs []int64) (map[int64]int, int, error) {
	counts := map[int64]int{}
	if len(courseIDs) == 0 {
		return counts, 0, nil
	}

	query := `
		SELECT course_id, count(*)
		FROM lesson_access
		WHERE course_id = ANY($1) AND completed_at IS NOT NULL
		GROUP BY course_id
	`

	rows, err := m.DB.Query(query, int64Array(courseIDs))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, 0, err
		}
		counts[id] = count
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	activeQuery := `
		SELECT count(DISTINCT user_id)
		FROM lesson_access
		WHERE course_id = ANY($1) AND completed_at IS NOT NULL
	`

	var active int
	err = m.DB.QueryRow(activeQuery, int64Array(courseIDs)).Scan(&active)
	if err != nil {
		return nil, 0, err
	}

	return counts, active, nil
}

// EnrollmentCountsByUser admin users ro'yxati uchun: user -> yozilgan kurslari.
func (m EnrollmentModel) EnrollmentCountsByUser(ids []int64) (map[int64]int, error) {
	query := `
		SELECT user_id, count(*)
		FROM enrollments
		WHERE user_id = ANY($1)
		GROUP BY user_id
	`

	rows, err := m.DB.Query(query, int64Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[int64]int{}

	for rows.Next() {
		var id int64
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		counts[id] = count
	}

	return counts, rows.Err()
}
