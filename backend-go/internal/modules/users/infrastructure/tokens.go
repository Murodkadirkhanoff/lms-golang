package infrastructure

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"

	"github.com/chashma/lms/internal/modules/users/application"
	"github.com/chashma/lms/internal/modules/users/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository stores SHA-256 hashes of password-reset tokens; the
// plaintext (Base32, no padding) is only ever returned to the caller.
type TokenRepository struct {
	pool *pgxpool.Pool
}

// NewTokenRepository builds a TokenRepository.
func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

var _ application.TokenRepository = (*TokenRepository)(nil)

var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)

// Create mints a token, stores its hash, and returns the plaintext.
func (r *TokenRepository) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	plaintext := base32NoPad.EncodeToString(raw)
	hash := sha256.Sum256([]byte(plaintext))

	const q = `
		INSERT INTO auth.password_reset_tokens (hash, user_id, expiry)
		VALUES ($1, $2, $3)`
	if _, err := r.pool.Exec(ctx, q, hash[:], userID, time.Now().Add(ttl)); err != nil {
		return "", err
	}
	return plaintext, nil
}

// UserIDForToken returns the user id for a valid, unexpired token.
func (r *TokenRepository) UserIDForToken(ctx context.Context, plaintext string) (int64, error) {
	hash := sha256.Sum256([]byte(plaintext))
	const q = `
		SELECT user_id
		FROM auth.password_reset_tokens
		WHERE hash = $1 AND expiry > NOW()`
	var userID int64
	if err := r.pool.QueryRow(ctx, q, hash[:]).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, err
	}
	return userID, nil
}

// DeleteAllForUser removes every reset token for the user.
func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM auth.password_reset_tokens WHERE user_id = $1`, userID)
	return err
}
