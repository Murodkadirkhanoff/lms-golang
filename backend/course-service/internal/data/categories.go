package data

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"lms.chashma.uz/pkg/validator"
)

var (
	ErrDuplicateSlug    = errors.New("duplicate slug")
	ErrInvalidParent    = errors.New("invalid parent category")
	ErrMaxDepthExceeded = errors.New("max category depth exceeded")
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a name into a URL-friendly slug, e.g. "My Category!" -> "my-category".
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonSlugChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// Category JSON kalitlari frontend Category tipiga mos (camelCase).
type Category struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Slug      string    `json:"slug"`
	NameUz    string    `json:"nameUz"`
	NameRu    string    `json:"nameRu"`
	NameEn    string    `json:"nameEn"`
	ParentID  *int64    `json:"parentId"`
	Version   int       `json:"-"`
}

func ValidateCategory(v *validator.Validator, category *Category) {
	v.Check(category.NameUz != "", "name_uz", "must be provided")
	v.Check(len(category.NameUz) <= 100, "name_uz", "must not be more than 100 bytes long")

	v.Check(category.NameRu != "", "name_ru", "must be provided")
	v.Check(len(category.NameRu) <= 100, "name_ru", "must not be more than 100 bytes long")

	v.Check(category.NameEn != "", "name_en", "must be provided")
	v.Check(len(category.NameEn) <= 100, "name_en", "must not be more than 100 bytes long")
}

type CategoryModel struct {
	DB *sql.DB
}

func (m CategoryModel) Insert(category *Category) error {
	query := `
		INSERT INTO categories (slug, name_uz, name_ru, name_en, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version
	`

	args := []any{category.Slug, category.NameUz, category.NameRu, category.NameEn, category.ParentID}

	err := m.DB.QueryRow(query, args...).Scan(&category.ID, &category.CreatedAt, &category.Version)
	if err != nil {
		return categoryConstraintError(err)
	}

	return nil
}

func (m CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, slug, name_uz, name_ru, name_en, parent_id, version
		FROM categories
		WHERE id = $1 AND deleted_at IS NULL
	`

	var category Category

	err := m.DB.QueryRow(query, id).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.Slug,
		&category.NameUz,
		&category.NameRu,
		&category.NameEn,
		&category.ParentID,
		&category.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &category, nil
}

func (m CategoryModel) List() ([]*Category, error) {
	query := `
		SELECT id, created_at, slug, name_uz, name_ru, name_en, parent_id, version
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY depth, id
	`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []*Category{}

	for rows.Next() {
		var category Category
		err := rows.Scan(
			&category.ID,
			&category.CreatedAt,
			&category.Slug,
			&category.NameUz,
			&category.NameRu,
			&category.NameEn,
			&category.ParentID,
			&category.Version,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	return categories, rows.Err()
}

func (m CategoryModel) Update(category *Category) error {
	query := `
		UPDATE categories
		SET slug = $1, name_uz = $2, name_ru = $3, name_en = $4, parent_id = $5, version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version
	`

	args := []any{
		category.Slug,
		category.NameUz,
		category.NameRu,
		category.NameEn,
		category.ParentID,
		category.ID,
		category.Version,
	}

	err := m.DB.QueryRow(query, args...).Scan(&category.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return categoryConstraintError(err)
		}
	}

	return nil
}

func (m CategoryModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM categories WHERE id = $1`

	result, err := m.DB.Exec(query, id)
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

// categoryConstraintError maps Postgres constraint violations to domain errors.
func categoryConstraintError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Constraint {
		case "categories_slug_key":
			return ErrDuplicateSlug
		case "categories_parent_id_fkey":
			return ErrInvalidParent
		case "max_category_depth":
			return ErrMaxDepthExceeded
		}
	}
	return err
}
