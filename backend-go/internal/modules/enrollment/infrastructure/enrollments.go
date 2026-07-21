// Package infrastructure contains the enrollment context's pgx adapters. Only
// this package touches the enrollment schema.
package infrastructure

import (
	"context"
	"errors"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnrollmentRepository is the pgx-backed enrollment.EnrollmentRepository.
type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

// NewEnrollmentRepository builds an EnrollmentRepository.
func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

var _ application.EnrollmentRepository = (*EnrollmentRepository)(nil)

// Insert is idempotent: returns the existing enrollment (isNew=false) on conflict.
func (r *EnrollmentRepository) Insert(ctx context.Context, userID, courseID int64) (domain.Enrollment, bool, error) {
	enr := domain.Enrollment{UserID: userID, CourseID: courseID}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO enrollment.enrollments (user_id, course_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, course_id) DO NOTHING
		RETURNING id, created_at`, userID, courseID).Scan(&enr.ID, &enr.CreatedAt)
	if err == nil {
		return enr, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Enrollment{}, false, err
	}
	err = r.pool.QueryRow(ctx, `
		SELECT id, created_at FROM enrollment.enrollments
		WHERE user_id = $1 AND course_id = $2`, userID, courseID).Scan(&enr.ID, &enr.CreatedAt)
	if err != nil {
		return domain.Enrollment{}, false, err
	}
	return enr, false, nil
}

// FindByID returns an enrollment.
func (r *EnrollmentRepository) FindByID(ctx context.Context, id int64) (*domain.Enrollment, error) {
	if id < 1 {
		return nil, domain.ErrNotFound
	}
	var e domain.Enrollment
	err := r.pool.QueryRow(ctx, `
		SELECT id, created_at, user_id, course_id FROM enrollment.enrollments WHERE id = $1`, id).
		Scan(&e.ID, &e.CreatedAt, &e.UserID, &e.CourseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// ListByUser returns all of a user's enrollments, newest first.
func (r *EnrollmentRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Enrollment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, user_id, course_id
		FROM enrollment.enrollments
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEnrollments(rows)
}

// PageByUser returns a page of a user's enrollments plus the total count.
func (r *EnrollmentRepository) PageByUser(ctx context.Context, userID int64, page, pageSize int) ([]domain.Enrollment, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT count(*) OVER() AS total, id, created_at, user_id, course_id
		FROM enrollment.enrollments
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []domain.Enrollment{}
	total := 0
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&total, &e.ID, &e.CreatedAt, &e.UserID, &e.CourseID); err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

// GrantLessonAccess grants access to the given lessons (idempotent).
func (r *EnrollmentRepository) GrantLessonAccess(ctx context.Context, userID, courseID int64, lessonIDs []int64) error {
	for _, lid := range lessonIDs {
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO enrollment.lesson_access (user_id, lesson_id, course_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, lesson_id) DO NOTHING`, userID, lid, courseID); err != nil {
			return err
		}
	}
	return nil
}

// SetLessonCompleted toggles a lesson's completion timestamp.
func (r *EnrollmentRepository) SetLessonCompleted(ctx context.Context, userID, lessonID int64, completed bool) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE enrollment.lesson_access
		SET completed_at = CASE WHEN $1 THEN NOW() ELSE NULL END
		WHERE user_id = $2 AND lesson_id = $3`, completed, userID, lessonID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CompletedCounts maps course id -> completed lesson count for a user.
func (r *EnrollmentRepository) CompletedCounts(ctx context.Context, userID int64) (map[int64]int, error) {
	return r.countMap(ctx, `
		SELECT course_id, count(*) AS n
		FROM enrollment.lesson_access
		WHERE user_id = $1 AND completed_at IS NOT NULL
		GROUP BY course_id`, userID)
}

// CompletedLessonIDs returns the set of a user's completed lesson ids.
func (r *EnrollmentRepository) CompletedLessonIDs(ctx context.Context, userID int64) (map[int64]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lesson_id FROM enrollment.lesson_access
		WHERE user_id = $1 AND completed_at IS NOT NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIDSet(rows)
}

// OwnedCourses returns the subset of courseIDs the user is enrolled in.
func (r *EnrollmentRepository) OwnedCourses(ctx context.Context, userID int64, courseIDs []int64) (map[int64]bool, error) {
	if len(courseIDs) == 0 {
		return map[int64]bool{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT course_id FROM enrollment.enrollments
		WHERE user_id = $1 AND course_id = ANY($2)`, userID, courseIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIDSet(rows)
}

// OwnedLessons returns the subset of lessonIDs the user has access to.
func (r *EnrollmentRepository) OwnedLessons(ctx context.Context, userID int64, lessonIDs []int64) (map[int64]bool, error) {
	if len(lessonIDs) == 0 {
		return map[int64]bool{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT lesson_id FROM enrollment.lesson_access
		WHERE user_id = $1 AND lesson_id = ANY($2)`, userID, lessonIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIDSet(rows)
}

// AccessibleLessonIDs lists the lesson ids a user may access in a course.
func (r *EnrollmentRepository) AccessibleLessonIDs(ctx context.Context, userID, courseID int64) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lesson_id FROM enrollment.lesson_access
		WHERE user_id = $1 AND course_id = $2`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CountsByCourses maps course id -> enrolled student count.
func (r *EnrollmentRepository) CountsByCourses(ctx context.Context, courseIDs []int64) (map[int64]int, error) {
	if len(courseIDs) == 0 {
		return map[int64]int{}, nil
	}
	return r.countMap(ctx, `
		SELECT course_id, count(*) AS n
		FROM enrollment.enrollments
		WHERE course_id = ANY($1)
		GROUP BY course_id`, courseIDs)
}

// DistinctStudentsForCourses counts unique enrolled users across courses.
func (r *EnrollmentRepository) DistinctStudentsForCourses(ctx context.Context, courseIDs []int64) (int, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT count(DISTINCT user_id) FROM enrollment.enrollments WHERE course_id = ANY($1)`, courseIDs).Scan(&n)
	return n, err
}

// CompletedStatsByCourses returns completed-record counts per course and the
// number of distinct students who completed at least one lesson.
func (r *EnrollmentRepository) CompletedStatsByCourses(ctx context.Context, courseIDs []int64) (application.CompletedStats, error) {
	if len(courseIDs) == 0 {
		return application.CompletedStats{Counts: map[int64]int{}}, nil
	}
	counts, err := r.countMap(ctx, `
		SELECT course_id, count(*) AS n
		FROM enrollment.lesson_access
		WHERE course_id = ANY($1) AND completed_at IS NOT NULL
		GROUP BY course_id`, courseIDs)
	if err != nil {
		return application.CompletedStats{}, err
	}
	var active int
	err = r.pool.QueryRow(ctx, `
		SELECT count(DISTINCT user_id)
		FROM enrollment.lesson_access
		WHERE course_id = ANY($1) AND completed_at IS NOT NULL`, courseIDs).Scan(&active)
	if err != nil {
		return application.CompletedStats{}, err
	}
	return application.CompletedStats{Counts: counts, ActiveStudents: active}, nil
}

// EnrollmentCountsByUser maps user id -> enrollment count.
func (r *EnrollmentRepository) EnrollmentCountsByUser(ctx context.Context, userIDs []int64) (map[int64]int, error) {
	if len(userIDs) == 0 {
		return map[int64]int{}, nil
	}
	return r.countMap(ctx, `
		SELECT user_id, count(*) AS n
		FROM enrollment.enrollments
		WHERE user_id = ANY($1)
		GROUP BY user_id`, userIDs)
}

func (r *EnrollmentRepository) countMap(ctx context.Context, q string, args ...any) (map[int64]int, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64]int{}
	for rows.Next() {
		var key int64
		var n int
		if err := rows.Scan(&key, &n); err != nil {
			return nil, err
		}
		m[key] = n
	}
	return m, rows.Err()
}

func scanEnrollments(rows pgx.Rows) ([]domain.Enrollment, error) {
	list := []domain.Enrollment{}
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&e.ID, &e.CreatedAt, &e.UserID, &e.CourseID); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func scanIDSet(rows pgx.Rows) (map[int64]bool, error) {
	set := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		set[id] = true
	}
	return set, rows.Err()
}
