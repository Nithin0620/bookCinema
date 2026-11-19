package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeStore struct {
	mu       sync.Mutex
	users    map[string]fakeUser
	byID     map[string]User
	sessions map[string]string
}

type fakeUser struct {
	user User
	hash string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:    map[string]fakeUser{},
		byID:     map[string]User{},
		sessions: map[string]string{},
	}
}

func (f *fakeStore) CreateUser(_ context.Context, id, email, passwordHash string, isAdmin bool) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.users[email]; exists {
		return User{}, ErrEmailTaken
	}
	user := User{ID: id, Email: email, CreatedAt: time.Now(), IsAdmin: isAdmin}
	f.users[email] = fakeUser{user: user, hash: passwordHash}
	f.byID[id] = user
	return user, nil
}

func (f *fakeStore) UserByEmail(_ context.Context, email string) (User, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	stored, exists := f.users[email]
	if !exists {
		return User{}, "", ErrUserNotFound
	}
	return stored.user, stored.hash, nil
}

func (f *fakeStore) CreateSession(_ context.Context, tokenHash, userID string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[tokenHash] = userID
	return nil
}

func (f *fakeStore) UserByTokenHash(_ context.Context, tokenHash string) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	userID, exists := f.sessions[tokenHash]
	if !exists {
		return User{}, ErrInvalidToken
	}
	return f.byID[userID], nil
}

func (f *fakeStore) DeleteSession(_ context.Context, tokenHash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, tokenHash)
	return nil
}

func TestRegisterLoginAuthenticateLogout(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, time.Hour, []string{"admin@example.com"})
	ctx := context.Background()

	user, token, err := service.Register(ctx, "  Alice@Example.com ", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("email was not normalized: %q", user.Email)
	}
	if token == "" {
		t.Fatal("register returned an empty token")
	}

	authenticated, _, err := service.Authenticate(ctx, token)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if authenticated.ID != user.ID {
		t.Fatalf("authenticated user mismatch: got %q want %q", authenticated.ID, user.ID)
	}

	loggedIn, loginToken, err := service.Login(ctx, "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loggedIn.ID != user.ID || loginToken == token {
		t.Fatalf("login returned unexpected user or token")
	}

	if _, _, err := service.Login(ctx, "alice@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: got %v want %v", err, ErrInvalidCredentials)
	}

	if err := service.Logout(ctx, token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, _, err := service.Authenticate(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("authenticate after logout: got %v want %v", err, ErrInvalidToken)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	service := NewService(newFakeStore(), time.Hour, nil)
	ctx := context.Background()

	if _, _, err := service.Register(ctx, "invalid", "password123"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("invalid email: got %v want %v", err, ErrInvalidEmail)
	}
	if _, _, err := service.Register(ctx, "alice@example.com", "short"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("weak password: got %v want %v", err, ErrWeakPassword)
	}
}

func TestRegisterPromotesConfiguredAdmin(t *testing.T) {
	service := NewService(newFakeStore(), time.Hour, []string{"admin@example.com"})
	user, _, err := service.Register(context.Background(), "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("register admin: %v", err)
	}
	if !user.IsAdmin {
		t.Fatal("configured admin email was not promoted")
	}

	regular, _, err := service.Register(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("register regular user: %v", err)
	}
	if regular.IsAdmin {
		t.Fatal("regular user was promoted")
	}
}
