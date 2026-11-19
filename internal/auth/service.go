package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store       Store
	sessionTTL  time.Duration
	adminEmails map[string]bool
}

func NewService(store Store, sessionTTL time.Duration, adminEmails []string) *Service {
	admins := make(map[string]bool, len(adminEmails))
	for _, email := range adminEmails {
		admins[normalizeEmail(email)] = true
	}
	return &Service{store: store, sessionTTL: sessionTTL, adminEmails: admins}
}

func (s *Service) Register(ctx context.Context, email, password string) (User, string, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return User{}, "", ErrInvalidEmail
	}
	if len(password) < 8 {
		return User{}, "", ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", err
	}

	user, err := s.store.CreateUser(ctx, uuid.NewString(), email, string(hash), s.adminEmails[email])
	if err != nil {
		return User{}, "", err
	}

	token, err := newToken()
	if err != nil {
		return User{}, "", err
	}
	if err := s.store.CreateSession(ctx, hashToken(token), user.ID, time.Now().Add(s.sessionTTL)); err != nil {
		return User{}, "", err
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	email = normalizeEmail(email)
	user, hash, err := s.store.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return User{}, "", ErrInvalidCredentials
		}
		return User{}, "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := newToken()
	if err != nil {
		return User{}, "", err
	}
	if err := s.store.CreateSession(ctx, hashToken(token), user.ID, time.Now().Add(s.sessionTTL)); err != nil {
		return User{}, "", err
	}

	return user, token, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, string, error) {
	if token == "" {
		return User{}, "", ErrInvalidToken
	}
	tokenHash := hashToken(token)
	user, err := s.store.UserByTokenHash(ctx, tokenHash)
	if err != nil {
		return User{}, "", err
	}
	return user, tokenHash, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, hashToken(token))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	at := strings.IndexByte(email, '@')
	return at > 0 && at < len(email)-1 && strings.Contains(email[at+1:], ".")
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
