package infrastructure

import (
	"context"
	"errors"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/domain"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CategoryRepository is the pgx-backed courses.CategoryRepository.
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository builds a CategoryRepository.
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

var _ application.CategoryRepository = (*CategoryRepository)(nil)

// Insert persists a new category (depth is set by a DB trigger).
func (r *CategoryRepository) Insert(ctx context.Context, c *domain.Category) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course.categories (slug, name_uz, name_ru, name_en, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version`,
		c.Slug, c.NameUz, c.NameRu, c.NameEn, c.ParentID,
	).Scan(&c.ID, &c.CreatedAt, &c.Version)
	if err != nil {
		return mapCategoryErr(err)
	}
	return nil
}

// FindByID returns an active category.
func (r *CategoryRepository) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	if id < 1 {
		return nil, domain.ErrNotFound
	}
	var c domain.Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, created_at, slug, name_uz, name_ru, name_en, parent_id, version
		FROM course.categories
		WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&c.ID, &c.CreatedAt, &c.Slug, &c.NameUz, &c.NameRu, &c.NameEn, &c.ParentID, &c.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// List returns all categories, with published-course counts rolled up so a
// parent's count includes its children's.
func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.created_at, c.slug, c.name_uz, c.name_ru, c.name_en, c.parent_id, c.version,
		       COALESCE(cc.n, 0) AS course_count
		FROM course.categories c
		LEFT JOIN (
		    SELECT category_id, COUNT(*) AS n
		    FROM course.courses
		    WHERE is_published = true AND deleted_at IS NULL AND category_id IS NOT NULL
		    GROUP BY category_id
		) cc ON cc.category_id = c.id
		WHERE c.deleted_at IS NULL
		ORDER BY depth, c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []domain.Category{}
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.CreatedAt, &c.Slug, &c.NameUz, &c.NameRu, &c.NameEn, &c.ParentID, &c.Version, &c.CourseCount); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	byID := map[int64]int{}
	for i, c := range categories {
		byID[c.ID] = i
	}
	for _, c := range categories {
		if c.ParentID != nil {
			if idx, ok := byID[*c.ParentID]; ok {
				categories[idx].CourseCount += c.CourseCount
			}
		}
	}
	return categories, nil
}

// Update persists category changes with optimistic locking.
func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	err := r.pool.QueryRow(ctx, `
		UPDATE course.categories
		SET slug = $1, name_uz = $2, name_ru = $3, name_en = $4, parent_id = $5, version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version`,
		c.Slug, c.NameUz, c.NameRu, c.NameEn, c.ParentID, c.ID, c.Version,
	).Scan(&c.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEditConflict
		}
		return mapCategoryErr(err)
	}
	return nil
}

// Delete removes a category.
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	if id < 1 {
		return domain.ErrNotFound
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM course.categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func mapCategoryErr(err error) error {
	switch database.Constraint(err) {
	case "categories_slug_key":
		return domain.ErrDuplicateSlug
	case "categories_parent_id_fkey":
		return domain.ErrInvalidParent
	case "max_category_depth":
		return domain.ErrMaxDepth
	}
	return err
}
