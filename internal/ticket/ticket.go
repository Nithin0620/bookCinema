package ticket

import (
	"context"
	"time"

	"github.com/Nithin0620/bookCinema/internal/booking"
)

type Ticket struct {
	ID        string    `json:"id"`
	BookingID string    `json:"booking_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type Detail struct {
	Ticket  `json:"ticket"`
	Booking booking.Booking `json:"booking"`
}

type Store interface {
	CreateForBookings(ctx context.Context, bookings []booking.Booking) ([]Ticket, error)
	ListByUser(ctx context.Context, userID string) ([]Detail, error)
	GetOwned(ctx context.Context, ticketID, userID string) (Detail, error)
}
