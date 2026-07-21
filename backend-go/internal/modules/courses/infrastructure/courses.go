// Package infrastructure contains the courses context's pgx adapters. Only
// this package touches the course schema.
package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseRepository is the pgx-backed courses.CourseRepository.
type CourseRepository struct {
	pool *pgxpool.Pool
}

// NewCourseRepository builds a CourseRepository.
func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

var _ application.CourseRepository = (*CourseRepository)(nil)

// listSelect is the shared projection for course list/detail with aggregates.
// Numeric columns are cast to float8 so they scan cleanly into float64.
const listSelect = `
SELECT count(*) OVER() AS total, c.id, c.created_at, c.slug, c.title, c.description,
       c.thumbnail_url, c.category_id, COALESCE(cat.slug, '') AS category_slug,
       c.lang, c.price::float8, c.is_published, c.instructor_id, c.student_count,
       COALESCE(agg.total_lessons, 0) AS total_lessons, COALESCE(agg.total_seconds, 0) AS total_seconds,
       COALESCE(rv.avg_rating, 0)::float8 AS avg_rating, COALESCE(rv.rating_count, 0) AS rating_count
FROM course.courses c
LEFT JOIN course.categories cat ON cat.id = c.category_id AND cat.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT count(l.id) AS total_lessons,
           COALESCE(sum(l.duration_seconds), 0) AS total_seconds
    FROM course.modules m
    JOIN course.lessons l ON l.module_id = m.id
    WHERE m.course_id = c.id
) agg ON true
LEFT JOIN LATERAL (
    SELECT round(avg(r.rating)::numeric, 1) AS avg_rating,
           count(r.id) AS rating_count
    FROM course.reviews r
    WHERE r.course_id = c.id
) rv ON true`

func scanCourse(row pgx.Row, total *int) (contract.CourseView, error) {
	var c contract.CourseView
	var seconds int
	err := row.Scan(total, &c.ID, &c.CreatedAt, &c.Slug, &c.Title, &c.Description,
		&c.ThumbnailURL, &c.CategoryID, &c.Category, &c.Lang, &c.Price, &c.IsPublished,
		&c.InstructorID, &c.StudentCount, &c.TotalLessons, &seconds, &c.Rating, &c.RatingCount)
	if err != nil {
		return contract.CourseView{}, err
	}
	c.TotalDurationMinutes = seconds / 60
	return c, nil
}

// List returns a filtered, sorted, paginated page of courses.
func (r *CourseRepository) List(ctx context.Context, f application.CourseFilters) ([]contract.CourseView, int, error) {
	orderBy := "c.created_at DESC, c.id DESC"
	switch f.Sort {
	case "popular":
		orderBy = "c.student_count DESC, c.created_at DESC, c.id DESC"
	case "price-asc":
		orderBy = "c.price ASC, c.id DESC"
	case "price-desc":
		orderBy = "c.price DESC, c.id DESC"
	}

	var sb strings.Builder
	sb.WriteString(listSelect)
	args := []any{f.Search, f.CategorySlug}
	sb.WriteString(`
WHERE c.deleted_at IS NULL
  AND ($1 = '' OR c.title ILIKE '%' || $1 || '%' OR c.description ILIKE '%' || $1 || '%')
  AND ($2 = '' OR c.category_id IN (
        SELECT id FROM course.categories
        WHERE slug = $2 OR parent_id = (SELECT id FROM course.categories WHERE slug = $2)
  ))`)

	if len(f.IDs) > 0 {
		args = append(args, f.IDs)
		sb.WriteString(fmt.Sprintf("\n  AND c.id = ANY($%d)", len(args)))
	}
	if f.InstructorID != 0 {
		args = append(args, f.InstructorID)
		sb.WriteString(fmt.Sprintf("\n  AND c.instructor_id = $%d", len(args)))
	}
	if !f.IncludeUnpublished {
		sb.WriteString("\n  AND c.is_published = true")
	}

	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	sb.WriteString(fmt.Sprintf("\nORDER BY %s LIMIT $%d OFFSET $%d", orderBy, len(args)-1, len(args)))

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []contract.CourseView{}
	total := 0
	for rows.Next() {
		c, err := scanCourse(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, c)
	}
	return list, total, rows.Err()
}

// FindByIDOrSlug returns a full course (with curriculum) by numeric id or slug.
func (r *CourseRepository) FindByIDOrSlug(ctx context.Context, idOrSlug string) (*contract.CourseView, error) {
	id, _ := strconv.ParseInt(idOrSlug, 10, 64)
	byID := id > 0

	q := listSelect + "\nWHERE c.deleted_at IS NULL AND "
	var arg any
	if byID {
		q += "c.id = $1"
		arg = id
	} else {
		q += "c.slug = $1"
		arg = idOrSlug
	}

	var total int
	c, err := scanCourse(r.pool.QueryRow(ctx, q, arg), &total)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	modules, err := r.modulesForCourse(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Modules = modules
	return &c, nil
}

func (r *CourseRepository) modulesForCourse(ctx context.Context, courseID int64) ([]contract.Module, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, position
		FROM course.modules
		WHERE course_id = $1
		ORDER BY position, id`, courseID)
	if err != nil {
		return nil, err
	}
	modules := []contract.Module{}
	byID := map[int64]int{} // module id -> index
	for rows.Next() {
		var m contract.Module
		if err := rows.Scan(&m.ID, &m.Title, &m.Position); err != nil {
			rows.Close()
			return nil, err
		}
		m.Lessons = []contract.Lesson{}
		byID[m.ID] = len(modules)
		modules = append(modules, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	lrows, err := r.pool.Query(ctx, `
		SELECT l.id, l.module_id, l.title, l.type, l.content_url, l.content,
		       l.duration_seconds, l.position, l.price::float8, l.is_free
		FROM course.lessons l
		JOIN course.modules m ON m.id = l.module_id
		WHERE m.course_id = $1
		ORDER BY l.position, l.id`, courseID)
	if err != nil {
		return nil, err
	}
	defer lrows.Close()
	for lrows.Next() {
		var l contract.Lesson
		var moduleID int64
		if err := lrows.Scan(&l.ID, &moduleID, &l.Title, &l.Type, &l.ContentURL, &l.Content,
			&l.DurationSeconds, &l.Position, &l.Price, &l.IsFree); err != nil {
			return nil, err
		}
		if idx, ok := byID[moduleID]; ok {
			modules[idx].Lessons = append(modules[idx].Lessons, l)
		}
	}
	return modules, lrows.Err()
}

// Insert writes a course plus its curriculum in one transaction.
func (r *CourseRepository) Insert(ctx context.Context, c *contract.CourseView) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	base := c.Slug
	for i := 2; ; i++ {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM course.courses WHERE slug = $1)`, c.Slug).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			break
		}
		c.Slug = base + "-" + strconv.Itoa(i)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO course.courses (title, slug, description, thumbnail_url, instructor_id, category_id, lang, price, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, version`,
		c.Title, c.Slug, c.Description, c.ThumbnailURL, c.InstructorID, c.CategoryID, c.Lang, c.Price, c.IsPublished,
	).Scan(&c.ID, &c.CreatedAt, &c.Version)
	if err != nil {
		return mapCourseErr(err)
	}

	if err := insertModules(ctx, tx, c.ID, c.Modules); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	seconds := 0
	c.TotalLessons = 0
	for _, m := range c.Modules {
		c.TotalLessons += len(m.Lessons)
		for _, l := range m.Lessons {
			seconds += l.DurationSeconds
		}
	}
	c.TotalDurationMinutes = seconds / 60
	return nil
}

// Update writes a course's main fields (curriculum handled separately).
func (r *CourseRepository) Update(ctx context.Context, c *contract.CourseView) error {
	err := r.pool.QueryRow(ctx, `
		UPDATE course.courses
		SET title = $1, description = $2, thumbnail_url = $3, category_id = $4, lang = $5,
		    price = $6, is_published = $7, version = version + 1
		WHERE id = $8 AND version = $9 AND deleted_at IS NULL
		RETURNING version`,
		c.Title, c.Description, c.ThumbnailURL, c.CategoryID, c.Lang, c.Price, c.IsPublished, c.ID, c.Version,
	).Scan(&c.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEditConflict
		}
		return mapCourseErr(err)
	}
	return nil
}

// ReplaceModules deletes and re-inserts a course's curriculum in a transaction.
func (r *CourseRepository) ReplaceModules(ctx context.Context, courseID int64, modules []contract.Module) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM course.modules WHERE course_id = $1`, courseID); err != nil {
		return err
	}
	if err := insertModules(ctx, tx, courseID, modules); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertModules(ctx context.Context, tx pgx.Tx, courseID int64, modules []contract.Module) error {
	for mi := range modules {
		m := &modules[mi]
		if err := tx.QueryRow(ctx, `
			INSERT INTO course.modules (course_id, title, position)
			VALUES ($1, $2, $3) RETURNING id`, courseID, m.Title, m.Position).Scan(&m.ID); err != nil {
			return err
		}
		for li := range m.Lessons {
			l := &m.Lessons[li]
			if err := tx.QueryRow(ctx, `
				INSERT INTO course.lessons (module_id, title, type, content_url, content,
				                            duration_seconds, position, price, is_free)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING id`,
				m.ID, l.Title, l.Type, l.ContentURL, l.Content, l.DurationSeconds, l.Position, l.Price, l.IsFree,
			).Scan(&l.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete soft-deletes a course.
func (r *CourseRepository) Delete(ctx context.Context, id int64) error {
	if id < 1 {
		return domain.ErrNotFound
	}
	tag, err := r.pool.Exec(ctx, `UPDATE course.courses SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CourseRepository) lessonInfoQuery(ctx context.Context, q string, ids []int64) ([]contract.LessonInfo, error) {
	if len(ids) == 0 {
		return []contract.LessonInfo{}, nil
	}
	rows, err := r.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []contract.LessonInfo{}
	for rows.Next() {
		var li contract.LessonInfo
		if err := rows.Scan(&li.ID, &li.Title, &li.Price, &li.IsFree, &li.CourseID, &li.CourseTitle); err != nil {
			return nil, err
		}
		out = append(out, li)
	}
	return out, rows.Err()
}

// LessonsForCourses returns lessons for the given courses, in curriculum order.
func (r *CourseRepository) LessonsForCourses(ctx context.Context, courseIDs []int64) ([]contract.LessonInfo, error) {
	return r.lessonInfoQuery(ctx, `
		SELECT l.id, l.title, l.price::float8, l.is_free, c.id, c.title
		FROM course.lessons l
		JOIN course.modules m ON m.id = l.module_id
		JOIN course.courses c ON c.id = m.course_id
		WHERE c.id = ANY($1) AND c.deleted_at IS NULL
		ORDER BY c.id, m.position, m.id, l.position, l.id`, courseIDs)
}

// LessonsByIDs returns lessons by their ids.
func (r *CourseRepository) LessonsByIDs(ctx context.Context, ids []int64) ([]contract.LessonInfo, error) {
	return r.lessonInfoQuery(ctx, `
		SELECT l.id, l.title, l.price::float8, l.is_free, c.id, c.title
		FROM course.lessons l
		JOIN course.modules m ON m.id = l.module_id
		JOIN course.courses c ON c.id = m.course_id
		WHERE l.id = ANY($1) AND c.deleted_at IS NULL`, ids)
}

// IncrementStudentCount bumps the student counter.
func (r *CourseRepository) IncrementStudentCount(ctx context.Context, courseID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE course.courses SET student_count = student_count + 1 WHERE id = $1`, courseID)
	return err
}

// Stats returns published-course and active-instructor counts.
func (r *CourseRepository) Stats(ctx context.Context) (contract.CourseStats, error) {
	var s contract.CourseStats
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) AS total, count(DISTINCT instructor_id) AS instructors
		FROM course.courses
		WHERE deleted_at IS NULL AND is_published = true`).Scan(&s.TotalCourses, &s.ActiveInstructors)
	return s, err
}

// CourseCountsByInstructor maps instructor id -> course count.
func (r *CourseRepository) CourseCountsByInstructor(ctx context.Context, ids []int64) (map[int64]int, error) {
	counts := map[int64]int{}
	if len(ids) == 0 {
		return counts, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT instructor_id, count(*) AS n
		FROM course.courses
		WHERE instructor_id = ANY($1) AND deleted_at IS NULL
		GROUP BY instructor_id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		counts[id] = n
	}
	return counts, rows.Err()
}

// InstructorStats aggregates published-course metrics per instructor.
func (r *CourseRepository) InstructorStats(ctx context.Context) ([]domain.InstructorStat, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.instructor_id, count(*) AS course_count, COALESCE(sum(c.student_count), 0) AS students,
		       COALESCE(round(avg(rv.avg_rating)::numeric, 1), 0)::float8 AS rating
		FROM course.courses c
		LEFT JOIN LATERAL (
		    SELECT avg(r.rating) AS avg_rating
		    FROM course.reviews r
		    WHERE r.course_id = c.id
		) rv ON true
		WHERE c.deleted_at IS NULL AND c.is_published = true
		GROUP BY c.instructor_id
		ORDER BY sum(c.student_count) DESC, count(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InstructorStat{}
	for rows.Next() {
		var st domain.InstructorStat
		if err := rows.Scan(&st.InstructorID, &st.CourseCount, &st.Students, &st.Rating); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func mapCourseErr(err error) error {
	switch database.Constraint(err) {
	case "courses_slug_key":
		return domain.ErrDuplicateSlug
	case "courses_category_id_fkey":
		return domain.ErrInvalidParent
	}
	return err
}
