package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidToken       = errors.New("invalid or expired session")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	IsAdmin   bool      `json:"is_admin"`
}

type Store interface {
	CreateUser(ctx context.Context, id, email, passwordHash string, isAdmin bool) (User, error)
	UserByEmail(ctx context.Context, email string) (User, string, error)
	CreateSession(ctx context.Context, tokenHash, userID string, expiresAt time.Time) error
	UserByTokenHash(ctx context.Context, tokenHash string) (User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}
