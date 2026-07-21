// Package infrastructure contains the users context's pgx adapters. Only this
// package touches the auth schema.
package infrastructure

import (
	"context"
	"errors"

	"github.com/chashma/lms/internal/modules/users/application"
	"github.com/chashma/lms/internal/modules/users/domain"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository is the pgx-backed users.Repository.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var _ application.Repository = (*UserRepository)(nil)

// Insert creates a user, returning domain.ErrDuplicateEmail on a unique clash.
func (r *UserRepository) Insert(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO auth.users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`
	err := r.pool.QueryRow(ctx, q, u.Name, u.Email, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt, &u.Version)
	if err != nil {
		if database.Constraint(err) == "users_email_key" {
			return domain.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

// FindByEmail returns the active user with the given email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, created_at, name, email, password_hash, role, version
		FROM auth.users
		WHERE email = $1 AND deleted_at IS NULL`
	return r.scanOne(r.pool.QueryRow(ctx, q, email))
}

// FindByID returns the active user with the given id.
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	if id < 1 {
		return nil, domain.ErrNotFound
	}
	const q = `
		SELECT id, created_at, name, email, password_hash, role, version
		FROM auth.users
		WHERE id = $1 AND deleted_at IS NULL`
	return r.scanOne(r.pool.QueryRow(ctx, q, id))
}

// FindByIDs returns active users for the given ids.
func (r *UserRepository) FindByIDs(ctx context.Context, ids []int64) ([]domain.User, error) {
	if len(ids) == 0 {
		return []domain.User{}, nil
	}
	const q = `
		SELECT id, created_at, name, email, role
		FROM auth.users
		WHERE id = ANY($1) AND deleted_at IS NULL`
	rows, err := r.pool.Query(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.CreatedAt, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdatePassword updates the password with optimistic locking.
func (r *UserRepository) UpdatePassword(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE auth.users
		SET password_hash = $1, version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`
	if err := r.pool.QueryRow(ctx, q, u.PasswordHash, u.ID, u.Version).Scan(&u.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEditConflict
		}
		return err
	}
	return nil
}

// UpdateName updates the display name with optimistic locking.
func (r *UserRepository) UpdateName(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE auth.users
		SET name = $1, version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`
	if err := r.pool.QueryRow(ctx, q, u.Name, u.ID, u.Version).Scan(&u.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEditConflict
		}
		return err
	}
	return nil
}

// UpdateRole assigns a new role.
func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	const q = `
		UPDATE auth.users
		SET role = $1, version = version + 1
		WHERE id = $2 AND deleted_at IS NULL`
	tag, err := r.pool.Exec(ctx, q, role, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Count returns the number of active users.
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM auth.users WHERE deleted_at IS NULL`).Scan(&n)
	return n, err
}

// List returns a page of users ordered by id, plus the total count.
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	const q = `
		SELECT count(*) OVER() AS total, id, created_at, name, email, role
		FROM auth.users
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []domain.User{}
	total := 0
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&total, &u.ID, &u.CreatedAt, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *UserRepository) scanOne(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.CreatedAt, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
