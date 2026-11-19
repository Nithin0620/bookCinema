package booking

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSeatAlreadyBooked = errors.New("seat is already taken")
	ErrSeatNotFound      = errors.New("seat does not exist")
	ErrShowtimeNotFound  = errors.New("showtime does not exist")
	ErrInvalidSession    = errors.New("invalid or expired session")
	ErrNotOwner          = errors.New("you do not own this session")
	ErrEmptyCheckout     = errors.New("no sessions selected")
	ErrBookingNotFound   = errors.New("booking not found")
)

type Hold struct {
	ID         string    `json:"id"`
	ShowtimeID string    `json:"showtime_id"`
	SeatID     string    `json:"seat_id"`
	UserID     string    `json:"user_id"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type Booking struct {
	ID         string    `json:"id"`
	ShowtimeID string    `json:"showtime_id"`
	SeatID     string    `json:"seat_id"`
	UserID     string    `json:"user_id"`
	PriceCents int       `json:"price_cents"`
	CreatedAt  time.Time `json:"created_at"`
}

type SeatStatus string

const (
	SeatAvailable SeatStatus = "available"
	SeatHeld      SeatStatus = "held"
	SeatConfirmed SeatStatus = "confirmed"
)

type SeatView struct {
	SeatID     string     `json:"seat_id"`
	Section    string     `json:"section"`
	PriceCents int        `json:"price_cents"`
	Status     SeatStatus `json:"status"`
	Mine       bool       `json:"mine"`
	SessionID  string     `json:"session_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type HoldStore interface {
	Create(ctx context.Context, hold Hold) (Hold, error)
	GetOwned(ctx context.Context, sessionID, userID string) (Hold, error)
	DeleteOwned(ctx context.Context, sessionID, userID string) error
	DeleteManyOwned(ctx context.Context, sessionIDs []string, userID string) error
	ListByShowtime(ctx context.Context, showtimeID string) ([]Hold, error)
}

type BookingStore interface {
	CreateMany(ctx context.Context, bookings []Booking) ([]Booking, error)
	Exists(ctx context.Context, showtimeID, seatID string) (bool, error)
	ListByShowtime(ctx context.Context, showtimeID string) ([]Booking, error)
	ListByUser(ctx context.Context, userID string) ([]Booking, error)
	DeleteOwned(ctx context.Context, bookingID, userID string) error
}
