//go:build integration

package booking

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Nithin0620/bookCinema/internal/auth"
	"github.com/Nithin0620/bookCinema/internal/catalog"
	"github.com/Nithin0620/bookCinema/internal/notify"
	"github.com/Nithin0620/bookCinema/internal/platform/postgres"
	redisplatform "github.com/Nithin0620/bookCinema/internal/platform/redis"
	"github.com/google/uuid"
)

func TestIntegrationHoldAndCheckout(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	if databaseURL == "" || redisAddr == "" {
		t.Skip("set DATABASE_URL and REDIS_ADDR to run integration tests")
	}

	ctx := context.Background()
	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	rdb, err := redisplatform.New(ctx, redisAddr)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	defer rdb.Close()

	userStore := auth.NewPostgresStore(db)
	user, err := userStore.CreateUser(ctx, uuid.NewString(), uuid.NewString()+"@example.com", "integration", false)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer db.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid`, user.ID)

	catalogStore := catalog.NewPostgresStore(db)
	showtime, err := catalogStore.ShowtimeByID(ctx, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("seeded showtime: %v", err)
	}

	bookingStore := NewPostgresBookingStore(db)
	service := NewService(NewRedisHoldStore(rdb, 2*time.Minute), bookingStore, catalogStore, 2*time.Minute)

	seats, err := service.ListSeats(ctx, showtime.ID, user.ID)
	if err != nil {
		t.Fatalf("list seats: %v", err)
	}
	seat := ""
	expectedPrice := 0
	for _, view := range seats {
		if view.Status == SeatAvailable {
			seat = view.SeatID
			expectedPrice = view.PriceCents
			break
		}
	}
	if seat == "" {
		t.Fatal("no available seat in seeded showtime")
	}

	hold, err := service.HoldSeat(ctx, showtime.ID, seat, user.ID)
	if err != nil {
		t.Fatalf("hold seat: %v", err)
	}

	bookings, err := service.Checkout(ctx, user.ID, []string{hold.ID})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if len(bookings) != 1 || bookings[0].SeatID != seat {
		t.Fatalf("unexpected bookings: %+v", bookings)
	}
	if bookings[0].PriceCents != expectedPrice {
		t.Fatalf("unexpected price: got %d want %d", bookings[0].PriceCents, expectedPrice)
	}

	notificationStore := notify.NewPostgresStore(db)
	if err := notificationStore.Create(ctx, user.ID, "Integration", "Booking confirmed"); err != nil {
		t.Fatalf("create notification: %v", err)
	}
	notifications, err := notificationStore.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifications) == 0 || notifications[0].Title != "Integration" {
		t.Fatalf("unexpected notifications: %+v", notifications)
	}

	if err := service.CancelBooking(ctx, bookings[0].ID, user.ID); err != nil {
		t.Fatalf("cleanup booking: %v", err)
	}
}
