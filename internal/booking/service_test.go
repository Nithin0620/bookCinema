package booking

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Nithin0620/bookCinema/internal/catalog"
)

type fakeHolds struct {
	mu    sync.Mutex
	holds map[string]Hold
}

func newFakeHolds() *fakeHolds {
	return &fakeHolds{holds: map[string]Hold{}}
}

func (f *fakeHolds) Create(_ context.Context, hold Hold) (Hold, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.holds {
		if existing.ShowtimeID == hold.ShowtimeID && existing.SeatID == hold.SeatID {
			return Hold{}, ErrSeatAlreadyBooked
		}
	}
	f.holds[hold.ID] = hold
	return hold, nil
}

func (f *fakeHolds) GetOwned(_ context.Context, sessionID, userID string) (Hold, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	hold, exists := f.holds[sessionID]
	if !exists {
		return Hold{}, ErrInvalidSession
	}
	if hold.UserID != userID {
		return Hold{}, ErrNotOwner
	}
	return hold, nil
}

func (f *fakeHolds) DeleteOwned(ctx context.Context, sessionID, userID string) error {
	if _, err := f.GetOwned(ctx, sessionID, userID); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.holds, sessionID)
	return nil
}

func (f *fakeHolds) DeleteManyOwned(ctx context.Context, sessionIDs []string, userID string) error {
	for _, id := range sessionIDs {
		if err := f.DeleteOwned(ctx, id, userID); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeHolds) ListByShowtime(_ context.Context, showtimeID string) ([]Hold, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := []Hold{}
	for _, hold := range f.holds {
		if hold.ShowtimeID == showtimeID {
			result = append(result, hold)
		}
	}
	return result, nil
}

type fakeBookings struct {
	mu       sync.Mutex
	bookings map[string]Booking
}

func newFakeBookings() *fakeBookings {
	return &fakeBookings{bookings: map[string]Booking{}}
}

func bookingKey(showtimeID, seatID string) string {
	return showtimeID + ":" + seatID
}

func (f *fakeBookings) CreateMany(_ context.Context, bookings []Booking) ([]Booking, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, booking := range bookings {
		key := bookingKey(booking.ShowtimeID, booking.SeatID)
		if _, exists := f.bookings[key]; exists {
			return nil, ErrSeatAlreadyBooked
		}
	}
	for _, booking := range bookings {
		f.bookings[bookingKey(booking.ShowtimeID, booking.SeatID)] = booking
	}
	return bookings, nil
}

func (f *fakeBookings) Exists(_ context.Context, showtimeID, seatID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, exists := f.bookings[bookingKey(showtimeID, seatID)]
	return exists, nil
}

func (f *fakeBookings) ListByShowtime(_ context.Context, showtimeID string) ([]Booking, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := []Booking{}
	for _, booking := range f.bookings {
		if booking.ShowtimeID == showtimeID {
			result = append(result, booking)
		}
	}
	return result, nil
}

func (f *fakeBookings) ListByUser(_ context.Context, userID string) ([]Booking, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := []Booking{}
	for _, booking := range f.bookings {
		if booking.UserID == userID {
			result = append(result, booking)
		}
	}
	return result, nil
}

func (f *fakeBookings) DeleteOwned(_ context.Context, bookingID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for key, booking := range f.bookings {
		if booking.ID == bookingID && booking.UserID == userID {
			delete(f.bookings, key)
			return nil
		}
	}
	return ErrBookingNotFound
}

type fakeShowtimes struct {
	showtime catalog.Showtime
}

func (f fakeShowtimes) ShowtimeByID(_ context.Context, id string) (catalog.Showtime, error) {
	if id != f.showtime.ID {
		return catalog.Showtime{}, catalog.ErrNotFound
	}
	return f.showtime, nil
}

func newTestService() (*Service, *fakeHolds, *fakeBookings, catalog.Showtime) {
	showtime := catalog.Showtime{
		ID:                 "showtime-1",
		MovieID:            "inception",
		Rows:               3,
		SeatsPerRow:        4,
		PremiumRows:        1,
		PremiumPriceCents:  1500,
		StandardPriceCents: 1000,
	}
	holds := newFakeHolds()
	bookings := newFakeBookings()
	service := NewService(holds, bookings, fakeShowtimes{showtime: showtime}, 2*time.Minute)
	return service, holds, bookings, showtime
}

func TestConcurrentHoldsExactlyOneWinner(t *testing.T) {
	service, _, _, showtime := newTestService()
	ctx := context.Background()

	const attempts = 100
	results := make(chan error, attempts)
	var waitGroup sync.WaitGroup

	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			_, err := service.HoldSeat(ctx, showtime.ID, "A1", fmt.Sprintf("user-%d", index))
			results <- err
		}(i)
	}
	waitGroup.Wait()
	close(results)

	winners, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, ErrSeatAlreadyBooked):
			conflicts++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if winners != 1 || conflicts != attempts-1 {
		t.Fatalf("got %d winners and %d conflicts, want 1 and %d", winners, conflicts, attempts-1)
	}
}

func TestCheckoutConfirmsHoldsAndReleasesThem(t *testing.T) {
	service, holds, bookings, showtime := newTestService()
	ctx := context.Background()

	first, err := service.HoldSeat(ctx, showtime.ID, "A1", "user-1")
	if err != nil {
		t.Fatalf("hold A1: %v", err)
	}
	second, err := service.HoldSeat(ctx, showtime.ID, "B1", "user-1")
	if err != nil {
		t.Fatalf("hold B1: %v", err)
	}

	confirmed, err := service.Checkout(ctx, "user-1", []string{first.ID, second.ID})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if len(confirmed) != 2 {
		t.Fatalf("got %d bookings, want 2", len(confirmed))
	}
	prices := map[string]int{}
	for _, booking := range confirmed {
		if booking.UserID != "user-1" {
			t.Fatalf("unexpected booking user: %+v", booking)
		}
		prices[booking.SeatID] = booking.PriceCents
	}
	if prices["A1"] != showtime.PremiumPriceCents || prices["B1"] != showtime.StandardPriceCents {
		t.Fatalf("unexpected section prices: %+v", prices)
	}
	if len(bookings.bookings) != 2 {
		t.Fatalf("got %d stored bookings, want 2", len(bookings.bookings))
	}
	if len(holds.holds) != 0 {
		t.Fatalf("holds were not removed after checkout: %d", len(holds.holds))
	}

	if _, err := service.Checkout(ctx, "user-1", []string{first.ID}); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("checkout reused session: got %v want %v", err, ErrInvalidSession)
	}
}

func TestListSeatsIncludesSectionAndPrice(t *testing.T) {
	service, _, _, showtime := newTestService()
	seats, err := service.ListSeats(context.Background(), showtime.ID, "user-1")
	if err != nil {
		t.Fatalf("list seats: %v", err)
	}
	if len(seats) != showtime.Rows*showtime.SeatsPerRow {
		t.Fatalf("got %d seats, want %d", len(seats), showtime.Rows*showtime.SeatsPerRow)
	}

	bySeat := map[string]SeatView{}
	for _, seat := range seats {
		bySeat[seat.SeatID] = seat
	}
	if bySeat["A1"].Section != "premium" || bySeat["A1"].PriceCents != showtime.PremiumPriceCents {
		t.Fatalf("unexpected premium seat: %+v", bySeat["A1"])
	}
	if bySeat["B1"].Section != "standard" || bySeat["B1"].PriceCents != showtime.StandardPriceCents {
		t.Fatalf("unexpected standard seat: %+v", bySeat["B1"])
	}
}

func TestCheckoutRejectsOtherUsersSession(t *testing.T) {
	service, _, _, showtime := newTestService()
	ctx := context.Background()

	hold, err := service.HoldSeat(ctx, showtime.ID, "B1", "user-1")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if _, err := service.Checkout(ctx, "user-2", []string{hold.ID}); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("got %v want %v", err, ErrNotOwner)
	}
}

func TestHoldSeatRejectsInvalidSeatAndMovie(t *testing.T) {
	service, _, bookings, showtime := newTestService()
	ctx := context.Background()

	if _, err := service.HoldSeat(ctx, showtime.ID, "Z9", "user-1"); !errors.Is(err, ErrSeatNotFound) {
		t.Fatalf("invalid seat: got %v want %v", err, ErrSeatNotFound)
	}
	if _, err := service.HoldSeat(ctx, "missing", "A1", "user-1"); !errors.Is(err, ErrShowtimeNotFound) {
		t.Fatalf("invalid showtime: got %v want %v", err, ErrShowtimeNotFound)
	}

	bookings.bookings[bookingKey(showtime.ID, "A1")] = Booking{ID: "booked", ShowtimeID: showtime.ID, SeatID: "A1"}
	if _, err := service.HoldSeat(ctx, showtime.ID, "A1", "user-1"); !errors.Is(err, ErrSeatAlreadyBooked) {
		t.Fatalf("confirmed seat: got %v want %v", err, ErrSeatAlreadyBooked)
	}
}

func TestListSeatsMarksOwnHolds(t *testing.T) {
	service, _, bookings, showtime := newTestService()
	ctx := context.Background()

	hold, err := service.HoldSeat(ctx, showtime.ID, "A1", "user-1")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	bookings.bookings[bookingKey(showtime.ID, "B1")] = Booking{ID: "b1", ShowtimeID: showtime.ID, SeatID: "B1"}

	seats, err := service.ListSeats(ctx, showtime.ID, "user-1")
	if err != nil {
		t.Fatalf("list seats: %v", err)
	}
	if len(seats) != showtime.Rows*showtime.SeatsPerRow {
		t.Fatalf("got %d seats, want %d", len(seats), showtime.Rows*showtime.SeatsPerRow)
	}

	byID := map[string]SeatView{}
	for _, seat := range seats {
		byID[seat.SeatID] = seat
	}
	if !byID["A1"].Mine || byID["A1"].SessionID != hold.ID || byID["A1"].Status != SeatHeld {
		t.Fatalf("own hold not marked correctly: %+v", byID["A1"])
	}
	if byID["B1"].Status != SeatConfirmed {
		t.Fatalf("confirmed seat not marked: %+v", byID["B1"])
	}
}
