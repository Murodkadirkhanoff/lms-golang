package data

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"lms.chashma.uz/internal/validator"
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

type Category struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Slug      string    `json:"slug"`
	NameUz    string    `json:"name_uz"`
	NameRu    string    `json:"name_ru"`
	NameEn    string    `json:"name_en"`
	ParentID  *int      `json:"parent_id,omitzero"`
	Version   int       `json:"version"`
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

	return nil
}

func (m CategoryModel) Get(id int) (*Category, error) {
	return nil, nil
}

func (m CategoryModel) Update(category *Category) error {
	return nil
}

func (m CategoryModel) Delete(id int) error {
	return nil
}
