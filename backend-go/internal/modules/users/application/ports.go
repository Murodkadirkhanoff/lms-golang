// Package application holds the users use-case services and the ports
// (interfaces) they depend on. Concrete adapters live in infrastructure.
package application

import (
	"context"
	"time"

	"github.com/chashma/lms/internal/modules/users/domain"
)

// Repository is the persistence port for users.
type Repository interface {
	Insert(ctx context.Context, u *domain.User) error // domain.ErrDuplicateEmail on conflict
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByIDs(ctx context.Context, ids []int64) ([]domain.User, error)
	UpdatePassword(ctx context.Context, u *domain.User) error // domain.ErrEditConflict
	UpdateName(ctx context.Context, u *domain.User) error     // domain.ErrEditConflict
	UpdateRole(ctx context.Context, id int64, role string) error
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, page, pageSize int) ([]domain.User, int, error)
}

// TokenRepository stores hashed password-reset tokens.
type TokenRepository interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	UserIDForToken(ctx context.Context, plaintext string) (int64, error) // domain.ErrNotFound if none/expired
	DeleteAllForUser(ctx context.Context, userID int64) error
}

// TokenIssuer mints JWTs (satisfied by the platform TokenMaker).
type TokenIssuer interface {
	New(userID int64, role string) (string, error)
}

// Mailer delivers the password-reset link.
type Mailer interface {
	SendPasswordReset(to, token string)
}
