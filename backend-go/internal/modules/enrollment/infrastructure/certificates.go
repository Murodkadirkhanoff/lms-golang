package infrastructure

import (
	"context"
	"errors"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CertificateRepository is the pgx-backed enrollment.CertificateRepository.
type CertificateRepository struct {
	pool *pgxpool.Pool
}

// NewCertificateRepository builds a CertificateRepository.
func NewCertificateRepository(pool *pgxpool.Pool) *CertificateRepository {
	return &CertificateRepository{pool: pool}
}

var _ application.CertificateRepository = (*CertificateRepository)(nil)

// Issue idempotently awards a certificate; the bool reports a fresh issue.
func (r *CertificateRepository) Issue(ctx context.Context, userID, courseID int64, courseTitle string) (bool, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO enrollment.certificates (user_id, course_id, course_title)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, course_id) DO NOTHING
		RETURNING id`, userID, courseID, courseTitle).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanCertificate(row pgx.Row) (domain.Certificate, error) {
	var c domain.Certificate
	err := row.Scan(&c.ID, &c.IssuedAt, &c.UserID, &c.CourseID, &c.CourseTitle)
	if err != nil {
		return domain.Certificate{}, err
	}
	c.Color = domain.ThumbnailColor(c.CourseID)
	return c, nil
}

// ListByUser returns a user's certificates, newest first.
func (r *CertificateRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, issued_at, user_id, course_id, course_title
		FROM enrollment.certificates
		WHERE user_id = $1
		ORDER BY issued_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Certificate{}
	for rows.Next() {
		c, err := scanCertificate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// FindForUser returns one certificate owned by the user.
func (r *CertificateRepository) FindForUser(ctx context.Context, id, userID int64) (*domain.Certificate, error) {
	c, err := scanCertificate(r.pool.QueryRow(ctx, `
		SELECT id, issued_at, user_id, course_id, course_title
		FROM enrollment.certificates
		WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// CountByUser counts a user's certificates.
func (r *CertificateRepository) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM enrollment.certificates WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}
