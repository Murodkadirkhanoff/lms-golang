package infrastructure

import (
	"context"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationRepository is the pgx-backed enrollment.NotificationRepository.
type NotificationRepository struct {
	pool *pgxpool.Pool
}

// NewNotificationRepository builds a NotificationRepository.
func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

var _ application.NotificationRepository = (*NotificationRepository)(nil)

// Insert records a notification.
func (r *NotificationRepository) Insert(ctx context.Context, userID int64, ntype, title, body string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO enrollment.notifications (user_id, type, title, body)
		VALUES ($1, $2, $3, $4)`, userID, ntype, title, body)
	return err
}

// ListByUser returns a user's most recent notifications.
func (r *NotificationRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, user_id, type, title, body, read
		FROM enrollment.notifications
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.CreatedAt, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Read); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// MarkAllRead marks all a user's unread notifications read.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE enrollment.notifications SET read = true WHERE user_id = $1 AND read = false`, userID)
	return err
}

// MarkRead marks one of a user's notifications read.
func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID int64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE enrollment.notifications SET read = true WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
