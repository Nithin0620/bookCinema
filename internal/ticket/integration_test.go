//go:build integration

package ticket

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Nithin0620/bookCinema/internal/auth"
	"github.com/Nithin0620/bookCinema/internal/booking"
	"github.com/Nithin0620/bookCinema/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestIntegrationTicketRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set DATABASE_URL to run integration tests")
	}

	ctx := context.Background()
	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	userStore := auth.NewPostgresStore(db)
	user, err := userStore.CreateUser(ctx, uuid.NewString(), uuid.NewString()+"@example.com", "integration", false)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer db.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid`, user.ID)

	created, err := booking.NewPostgresBookingStore(db).CreateMany(ctx, []booking.Booking{
		{
			ID:         uuid.NewString(),
			ShowtimeID: "11111111-1111-1111-1111-111111111111",
			SeatID:     "Z" + uuid.NewString()[:8],
			UserID:     user.ID,
			PriceCents: 1800,
			CreatedAt:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}
	defer db.Exec(ctx, `DELETE FROM bookings WHERE id = $1::uuid`, created[0].ID)

	store := NewPostgresStore(db)
	tickets, err := store.CreateForBookings(ctx, created)
	if err != nil {
		t.Fatalf("create tickets: %v", err)
	}
	if len(tickets) != 1 || tickets[0].Code == "" {
		t.Fatalf("unexpected tickets: %+v", tickets)
	}
	if _, err := QRCode(tickets[0].Code); err != nil {
		t.Fatalf("generate QR: %v", err)
	}

	listed, err := store.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if len(listed) != 1 || listed[0].Ticket.ID != tickets[0].ID {
		t.Fatalf("unexpected listed tickets: %+v", listed)
	}

	owned, err := store.GetOwned(ctx, tickets[0].ID, user.ID)
	if err != nil {
		t.Fatalf("get owned ticket: %v", err)
	}
	if owned.Booking.ID != created[0].ID {
		t.Fatalf("unexpected ticket booking: %+v", owned.Booking)
	}
}
