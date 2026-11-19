package notify

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(ctx context.Context, userID, title, body string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO notifications (id, user_id, title, body)
		VALUES ($1::uuid, $2::uuid, $3, $4)
	`, uuid.NewString(), userID, title, body)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID string) ([]Notification, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, user_id::text, title, body, read_at, created_at
		FROM notifications
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.ReadAt,
			&notification.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return notifications, nil
}

func (s *PostgresStore) MarkAllRead(ctx context.Context, userID string) error {
	if _, err := s.db.Exec(ctx, `
		UPDATE notifications SET read_at = now()
		WHERE user_id = $1::uuid AND read_at IS NULL
	`, userID); err != nil {
		return fmt.Errorf("mark notifications read: %w", err)
	}
	return nil
}

var _ Store = (*PostgresStore)(nil)
