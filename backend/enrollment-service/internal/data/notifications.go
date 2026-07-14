package data

import (
	"database/sql"
	"time"
)

// Notification JSON shakli frontend Notification tipiga mos (types/index.ts:154).
type Notification struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserID    int64     `json:"-"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
}

type NotificationModel struct {
	DB *sql.DB
}

func (m NotificationModel) Insert(userID int64, notifType, title, body string) error {
	_, err := m.DB.Exec(
		`INSERT INTO notifications (user_id, type, title, body) VALUES ($1, $2, $3, $4)`,
		userID, notifType, title, body,
	)
	return err
}

func (m NotificationModel) ListByUser(userID int64) ([]*Notification, error) {
	query := `
		SELECT id, created_at, user_id, type, title, body, read
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 50
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []*Notification{}

	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.CreatedAt, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Read)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

func (m NotificationModel) MarkAllRead(userID int64) error {
	_, err := m.DB.Exec(`UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`, userID)
	return err
}
