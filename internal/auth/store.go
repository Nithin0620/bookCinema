package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateUser(ctx context.Context, id, email, passwordHash string, isAdmin bool) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, is_admin)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, created_at, is_admin
	`, id, email, passwordHash, isAdmin).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.IsAdmin)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) UserByEmail(ctx context.Context, email string) (User, string, error) {
	var (
		user         User
		passwordHash string
	)
	err := s.db.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at, is_admin
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrUserNotFound
	}
	if err != nil {
		return User{}, "", fmt.Errorf("select user: %w", err)
	}
	return user, passwordHash, nil
}

func (s *PostgresStore) CreateSession(ctx context.Context, tokenHash, userID string, expiresAt time.Time) error {
	if _, err := s.db.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, tokenHash, userID, expiresAt); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *PostgresStore) UserByTokenHash(ctx context.Context, tokenHash string) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.created_at, u.is_admin
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()
	`, tokenHash).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidToken
	}
	if err != nil {
		return User{}, fmt.Errorf("select session: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) DeleteSession(ctx context.Context, tokenHash string) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
