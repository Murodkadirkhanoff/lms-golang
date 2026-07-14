package data

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"lms.chashma.uz/pkg/validator"
)

// JSON kalitlari frontend Course/Module/Lesson/Instructor tiplariga mos
// (frontend/src/types/index.ts). Instructor ma'lumoti auth-service'da —
// handler to'ldiradi.

type Instructor struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Headline    string  `json:"headline"`
	AvatarColor string  `json:"avatarColor"`
	Students    int     `json:"students"`
	Courses     int     `json:"courses"`
	Rating      float64 `json:"rating"`
}

type Lesson struct {
	ID              int64   `json:"id"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	ContentURL      string  `json:"contentUrl,omitempty"`
	Content         string  `json:"content,omitempty"`
	DurationSeconds int     `json:"durationSeconds"`
	Position        int     `json:"-"`
	Price           float64 `json:"price"`
	IsFree          bool    `json:"isFree"`
}

type Module struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Position int       `json:"-"`
	Lessons  []*Lesson `json:"lessons"`
}

type Course struct {
	ID                   int64       `json:"id"`
	CreatedAt            time.Time   `json:"createdAt"`
	Slug                 string      `json:"slug"`
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	ThumbnailColor       string      `json:"thumbnailColor"`
	CategoryID           *int64      `json:"-"`
	Category             string      `json:"category"`
	Lang                 string      `json:"lang"`
	Price                float64     `json:"price"`
	Rating               float64     `json:"rating"`
	RatingCount          int         `json:"ratingCount"`
	StudentCount         int         `json:"studentCount"`
	IsPublished          bool        `json:"isPublished"`
	InstructorID         int64       `json:"-"`
	Instructor           *Instructor `json:"instructor"`
	Modules              []*Module   `json:"modules,omitempty"`
	Reviews              []*Review   `json:"reviews,omitempty"`
	TotalLessons         int         `json:"totalLessons"`
	TotalDurationMinutes int         `json:"totalDurationMinutes"`
	Version              int         `json:"-"`
}

func ValidateCourse(v *validator.Validator, course *Course) {
	v.Check(course.Title != "", "title", "must be provided")
	v.Check(len(course.Title) <= 200, "title", "must not be more than 200 bytes long")
	v.Check(validator.PermittedValue(course.Lang, "uz", "ru", "en"), "lang", "must be one of uz, ru, en")
	v.Check(course.Price >= 0, "price", "must not be negative")

	for mi, module := range course.Modules {
		v.Check(module.Title != "", fmt.Sprintf("modules[%d].title", mi), "must be provided")
		for li, lesson := range module.Lessons {
			key := fmt.Sprintf("modules[%d].lessons[%d]", mi, li)
			v.Check(lesson.Title != "", key+".title", "must be provided")
			v.Check(validator.PermittedValue(lesson.Type, "video", "text"), key+".type", "must be one of video, text")
			v.Check(lesson.Price >= 0, key+".price", "must not be negative")
			v.Check(!lesson.IsFree || lesson.Price == 0, key+".price", "free lessons must have price 0")
			v.Check(lesson.DurationSeconds >= 0, key+".durationSeconds", "must not be negative")
		}
	}
}

// CourseFilters - GET /v1/courses query parametrlari.
type CourseFilters struct {
	Search             string
	CategorySlug       string
	Sort               string
	Page               int
	PageSize           int
	IDs                []int64
	InstructorID       int64
	IncludeUnpublished bool
}

type CourseModel struct {
	DB *sql.DB
}

// listSelect kurs ro'yxati/detali uchun umumiy ustunlar va aggregatlar.
const listSelect = `
	SELECT count(*) OVER(), c.id, c.created_at, c.slug, c.title, c.description,
	       c.category_id, COALESCE(cat.slug, ''), c.lang, c.price, c.is_published,
	       c.instructor_id, c.student_count,
	       COALESCE(agg.total_lessons, 0), COALESCE(agg.total_seconds, 0),
	       COALESCE(rv.avg_rating, 0), COALESCE(rv.rating_count, 0)
	FROM courses c
	LEFT JOIN categories cat ON cat.id = c.category_id AND cat.deleted_at IS NULL
	LEFT JOIN LATERAL (
	    SELECT count(l.id) AS total_lessons,
	           COALESCE(sum(l.duration_seconds), 0) AS total_seconds
	    FROM modules m
	    JOIN lessons l ON l.module_id = m.id
	    WHERE m.course_id = c.id
	) agg ON true
	LEFT JOIN LATERAL (
	    SELECT round(avg(r.rating)::numeric, 1) AS avg_rating,
	           count(r.id) AS rating_count
	    FROM reviews r
	    WHERE r.course_id = c.id
	) rv ON true
`

func scanCourse(rows interface{ Scan(...any) error }, total *int) (*Course, error) {
	var course Course
	var totalSeconds int

	err := rows.Scan(
		total,
		&course.ID,
		&course.CreatedAt,
		&course.Slug,
		&course.Title,
		&course.Description,
		&course.CategoryID,
		&course.Category,
		&course.Lang,
		&course.Price,
		&course.IsPublished,
		&course.InstructorID,
		&course.StudentCount,
		&course.TotalLessons,
		&totalSeconds,
		&course.Rating,
		&course.RatingCount,
	)
	if err != nil {
		return nil, err
	}

	course.TotalDurationMinutes = totalSeconds / 60

	return &course, nil
}

func (m CourseModel) List(filters CourseFilters) ([]*Course, int, error) {
	orderBy := "c.created_at DESC, c.id DESC"
	switch filters.Sort {
	case "popular":
		orderBy = "c.student_count DESC, c.created_at DESC, c.id DESC"
	case "price-asc":
		orderBy = "c.price ASC, c.id DESC"
	case "price-desc":
		orderBy = "c.price DESC, c.id DESC"
	}

	query := listSelect + `
		WHERE c.deleted_at IS NULL
		  AND ($1 = '' OR c.title ILIKE '%' || $1 || '%' OR c.description ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR c.category_id IN (
		        SELECT id FROM categories
		        WHERE slug = $2 OR parent_id = (SELECT id FROM categories WHERE slug = $2)
		  ))
		  AND (cardinality($3::bigint[]) = 0 OR c.id = ANY($3))
		  AND ($4::bigint = 0 OR c.instructor_id = $4)
		  AND ($5 OR c.is_published = true)
		ORDER BY ` + orderBy + `
		LIMIT $6 OFFSET $7
	`

	// nil slice pq.Array'da SQL NULL bo'ladi va cardinality(NULL) butun
	// shartni yiqitadi — bo'sh bo'lsa ham haqiqiy bo'sh massiv yuboramiz.
	ids := filters.IDs
	if ids == nil {
		ids = []int64{}
	}

	args := []any{
		filters.Search,
		filters.CategorySlug,
		pq.Array(ids),
		filters.InstructorID,
		filters.IncludeUnpublished,
		filters.PageSize,
		(filters.Page - 1) * filters.PageSize,
	}

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	total := 0
	courses := []*Course{}

	for rows.Next() {
		course, err := scanCourse(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		courses = append(courses, course)
	}

	return courses, total, rows.Err()
}

// GetByIDOrSlug bitta kursni modules[].lessons[] bilan to'liq qaytaradi.
func (m CourseModel) GetByIDOrSlug(idOrSlug string) (*Course, error) {
	id, err := strconv.ParseInt(idOrSlug, 10, 64)
	byID := err == nil && id > 0

	query := listSelect + ` WHERE c.deleted_at IS NULL AND `
	var arg any
	if byID {
		query += `c.id = $1`
		arg = id
	} else {
		query += `c.slug = $1`
		arg = idOrSlug
	}

	var total int
	course, err := scanCourse(m.DB.QueryRow(query, arg), &total)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	course.Modules, err = m.modulesForCourse(course.ID)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (m CourseModel) modulesForCourse(courseID int64) ([]*Module, error) {
	query := `
		SELECT id, title, position
		FROM modules
		WHERE course_id = $1
		ORDER BY position, id
	`

	rows, err := m.DB.Query(query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modules := []*Module{}
	byID := map[int64]*Module{}

	for rows.Next() {
		var module Module
		err := rows.Scan(&module.ID, &module.Title, &module.Position)
		if err != nil {
			return nil, err
		}
		module.Lessons = []*Lesson{}
		modules = append(modules, &module)
		byID[module.ID] = &module
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	lessonQuery := `
		SELECT l.id, l.module_id, l.title, l.type, l.content_url, l.content,
		       l.duration_seconds, l.position, l.price, l.is_free
		FROM lessons l
		JOIN modules m ON m.id = l.module_id
		WHERE m.course_id = $1
		ORDER BY l.position, l.id
	`

	lessonRows, err := m.DB.Query(lessonQuery, courseID)
	if err != nil {
		return nil, err
	}
	defer lessonRows.Close()

	for lessonRows.Next() {
		var lesson Lesson
		var moduleID int64
		err := lessonRows.Scan(
			&lesson.ID,
			&moduleID,
			&lesson.Title,
			&lesson.Type,
			&lesson.ContentURL,
			&lesson.Content,
			&lesson.DurationSeconds,
			&lesson.Position,
			&lesson.Price,
			&lesson.IsFree,
		)
		if err != nil {
			return nil, err
		}
		if module, ok := byID[moduleID]; ok {
			module.Lessons = append(module.Lessons, &lesson)
		}
	}

	return modules, lessonRows.Err()
}

// Insert kurs + modules + lessons'ni bitta tranzaksiyada yozadi.
func (m CourseModel) Insert(course *Course) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Slug band bo'lsa -2, -3... qo'shib ketamiz.
	base := course.Slug
	for i := 2; ; i++ {
		var exists bool
		err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM courses WHERE slug = $1)`, course.Slug).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			break
		}
		course.Slug = fmt.Sprintf("%s-%d", base, i)
	}

	query := `
		INSERT INTO courses (title, slug, description, instructor_id, category_id, lang, price, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, version
	`

	args := []any{
		course.Title,
		course.Slug,
		course.Description,
		course.InstructorID,
		course.CategoryID,
		course.Lang,
		course.Price,
		course.IsPublished,
	}

	err = tx.QueryRow(query, args...).Scan(&course.ID, &course.CreatedAt, &course.Version)
	if err != nil {
		return courseConstraintError(err)
	}

	for _, module := range course.Modules {
		err = tx.QueryRow(
			`INSERT INTO modules (course_id, title, position) VALUES ($1, $2, $3) RETURNING id`,
			course.ID, module.Title, module.Position,
		).Scan(&module.ID)
		if err != nil {
			return err
		}

		for _, lesson := range module.Lessons {
			err = tx.QueryRow(
				`INSERT INTO lessons (module_id, title, type, content_url, content, duration_seconds, position, price, is_free)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				 RETURNING id`,
				module.ID, lesson.Title, lesson.Type, lesson.ContentURL, lesson.Content,
				lesson.DurationSeconds, lesson.Position, lesson.Price, lesson.IsFree,
			).Scan(&lesson.ID)
			if err != nil {
				return err
			}
		}
	}

	// Aggregatlar javob uchun (DB'dan qayta o'qimaymiz).
	for _, module := range course.Modules {
		course.TotalLessons += len(module.Lessons)
		for _, lesson := range module.Lessons {
			course.TotalDurationMinutes += lesson.DurationSeconds
		}
	}
	course.TotalDurationMinutes /= 60

	return tx.Commit()
}

// Update kursning asosiy maydonlarini yangilaydi (modules/lessons alohida).
func (m CourseModel) Update(course *Course) error {
	query := `
		UPDATE courses
		SET title = $1, description = $2, category_id = $3, lang = $4,
		    price = $5, is_published = $6, version = version + 1
		WHERE id = $7 AND version = $8 AND deleted_at IS NULL
		RETURNING version
	`

	args := []any{
		course.Title,
		course.Description,
		course.CategoryID,
		course.Lang,
		course.Price,
		course.IsPublished,
		course.ID,
		course.Version,
	}

	err := m.DB.QueryRow(query, args...).Scan(&course.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return courseConstraintError(err)
		}
	}

	return nil
}

// Delete soft-delete qiladi (enrollments boshqa servisda bo'lgani uchun
// qattiq o'chirish xavfli).
func (m CourseModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	result, err := m.DB.Exec(`UPDATE courses SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
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

// LessonInfo internal endpointlar uchun yengil ko'rinish.
type LessonInfo struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	IsFree      bool    `json:"isFree"`
	CourseID    int64   `json:"courseId"`
	CourseTitle string  `json:"courseTitle"`
}

// LessonsForCourse kursning barcha darslarini tartib bilan qaytaradi
// (enrollment-service lesson_access to'ldirishi va currentLesson uchun).
func (m CourseModel) LessonsForCourse(courseID int64) ([]*LessonInfo, error) {
	return m.LessonsForCourses([]int64{courseID})
}

// LessonsForCourses bir nechta kursning darslarini modul/dars tartibida
// qaytaradi (me/courses currentLesson hisoblashi uchun batch).
func (m CourseModel) LessonsForCourses(courseIDs []int64) ([]*LessonInfo, error) {
	query := `
		SELECT l.id, l.title, l.price, l.is_free, c.id, c.title
		FROM lessons l
		JOIN modules m ON m.id = l.module_id
		JOIN courses c ON c.id = m.course_id
		WHERE c.id = ANY($1) AND c.deleted_at IS NULL
		ORDER BY c.id, m.position, m.id, l.position, l.id
	`

	return m.queryLessonInfos(query, pq.Array(courseIDs))
}

// LessonsByIDs alohida sotib olinayotgan darslar uchun (checkout).
func (m CourseModel) LessonsByIDs(ids []int64) ([]*LessonInfo, error) {
	query := `
		SELECT l.id, l.title, l.price, l.is_free, c.id, c.title
		FROM lessons l
		JOIN modules m ON m.id = l.module_id
		JOIN courses c ON c.id = m.course_id
		WHERE l.id = ANY($1) AND c.deleted_at IS NULL
	`

	return m.queryLessonInfos(query, pq.Array(ids))
}

// IncrementStudentCount enrollment-service yangi yozuv yaratganda chaqiriladi.
func (m CourseModel) IncrementStudentCount(courseID int64) error {
	_, err := m.DB.Exec(
		`UPDATE courses SET student_count = student_count + 1 WHERE id = $1`,
		courseID,
	)
	return err
}

// Stats admin panel uchun umumiy hisobot.
func (m CourseModel) Stats() (totalCourses, activeInstructors int, err error) {
	query := `
		SELECT count(*), count(DISTINCT instructor_id)
		FROM courses
		WHERE deleted_at IS NULL AND is_published = true
	`
	err = m.DB.QueryRow(query).Scan(&totalCourses, &activeInstructors)
	return totalCourses, activeInstructors, err
}

// CourseCountsByInstructor admin users ro'yxati uchun: user -> yaratgan kurslari.
func (m CourseModel) CourseCountsByInstructor(ids []int64) (map[int64]int, error) {
	query := `
		SELECT instructor_id, count(*)
		FROM courses
		WHERE instructor_id = ANY($1) AND deleted_at IS NULL
		GROUP BY instructor_id
	`

	rows, err := m.DB.Query(query, pq.Array(ids))
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

func (m CourseModel) queryLessonInfos(query string, arg any) ([]*LessonInfo, error) {
	rows, err := m.DB.Query(query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := []*LessonInfo{}

	for rows.Next() {
		var l LessonInfo
		err := rows.Scan(&l.ID, &l.Title, &l.Price, &l.IsFree, &l.CourseID, &l.CourseTitle)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, &l)
	}

	return lessons, rows.Err()
}

// InstructorStat kurslar jadvalidan hisoblangan instruktor statistikasi.
type InstructorStat struct {
	InstructorID int64
	CourseCount  int
	Students     int
	Rating       float64
}

func (m CourseModel) InstructorStats() ([]*InstructorStat, error) {
	query := `
		SELECT c.instructor_id, count(*), COALESCE(sum(c.student_count), 0),
		       COALESCE(round(avg(rv.avg_rating)::numeric, 1), 0)
		FROM courses c
		LEFT JOIN LATERAL (
		    SELECT avg(r.rating) AS avg_rating
		    FROM reviews r
		    WHERE r.course_id = c.id
		) rv ON true
		WHERE c.deleted_at IS NULL AND c.is_published = true
		GROUP BY c.instructor_id
		ORDER BY sum(c.student_count) DESC, count(*) DESC
	`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*InstructorStat{}

	for rows.Next() {
		var s InstructorStat
		err := rows.Scan(&s.InstructorID, &s.CourseCount, &s.Students, &s.Rating)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &s)
	}

	return stats, rows.Err()
}

func courseConstraintError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Constraint {
		case "courses_slug_key":
			return ErrDuplicateSlug
		case "courses_category_id_fkey":
			return ErrInvalidParent
		}
	}
	return err
}
