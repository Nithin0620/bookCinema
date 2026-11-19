package notify

import (
	"context"
	"time"
)

type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type Store interface {
	Create(ctx context.Context, userID, title, body string) error
	ListByUser(ctx context.Context, userID string) ([]Notification, error)
	MarkAllRead(ctx context.Context, userID string) error
}
