package booking

import (
	"context"
	"errors"
	"time"

	"github.com/Nithin0620/bookCinema/internal/catalog"
	"github.com/google/uuid"
)

type ShowtimeLookup interface {
	ShowtimeByID(ctx context.Context, id string) (catalog.Showtime, error)
}

type Service struct {
	holds      HoldStore
	bookings   BookingStore
	showtimes  ShowtimeLookup
	sessionTTL time.Duration
}

func NewService(holds HoldStore, bookings BookingStore, showtimes ShowtimeLookup, sessionTTL time.Duration) *Service {
	return &Service{
		holds:      holds,
		bookings:   bookings,
		showtimes:  showtimes,
		sessionTTL: sessionTTL,
	}
}

func (s *Service) HoldSeat(ctx context.Context, showtimeID, seatID, userID string) (Hold, error) {
	showtime, err := s.showtimes.ShowtimeByID(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			return Hold{}, ErrShowtimeNotFound
		}
		return Hold{}, err
	}
	if !showtime.ValidSeat(seatID) {
		return Hold{}, ErrSeatNotFound
	}

	exists, err := s.bookings.Exists(ctx, showtimeID, seatID)
	if err != nil {
		return Hold{}, err
	}
	if exists {
		return Hold{}, ErrSeatAlreadyBooked
	}

	return s.holds.Create(ctx, Hold{
		ID:         uuid.NewString(),
		ShowtimeID: showtimeID,
		SeatID:     seatID,
		UserID:     userID,
		ExpiresAt:  time.Now().Add(s.sessionTTL),
	})
}

func (s *Service) Checkout(ctx context.Context, userID string, sessionIDs []string) ([]Booking, error) {
	if len(sessionIDs) == 0 {
		return nil, ErrEmptyCheckout
	}

	seen := make(map[string]bool, len(sessionIDs))
	holds := make([]Hold, 0, len(sessionIDs))
	showtimeID := ""

	for _, id := range sessionIDs {
		if seen[id] {
			return nil, ErrInvalidSession
		}
		seen[id] = true

		hold, err := s.holds.GetOwned(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		if showtimeID == "" {
			showtimeID = hold.ShowtimeID
		} else if hold.ShowtimeID != showtimeID {
			return nil, ErrInvalidSession
		}
		holds = append(holds, hold)
	}

	showtime, err := s.showtimes.ShowtimeByID(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			return nil, ErrShowtimeNotFound
		}
		return nil, err
	}

	now := time.Now()
	bookings := make([]Booking, 0, len(holds))
	for _, hold := range holds {
		_, price := showtime.SectionForSeat(hold.SeatID)
		bookings = append(bookings, Booking{
			ID:         uuid.NewString(),
			ShowtimeID: showtime.ID,
			SeatID:     hold.SeatID,
			UserID:     userID,
			PriceCents: price,
			CreatedAt:  now,
		})
	}

	created, err := s.bookings.CreateMany(ctx, bookings)
	if err != nil {
		return nil, err
	}

	if err := s.holds.DeleteManyOwned(ctx, sessionIDs, userID); err != nil {
		return created, nil
	}
	return created, nil
}

func (s *Service) Release(ctx context.Context, sessionID, userID string) error {
	return s.holds.DeleteOwned(ctx, sessionID, userID)
}

func (s *Service) ListSeats(ctx context.Context, showtimeID, userID string) ([]SeatView, error) {
	showtime, err := s.showtimes.ShowtimeByID(ctx, showtimeID)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			return nil, ErrShowtimeNotFound
		}
		return nil, err
	}

	seats := make([]SeatView, 0, showtime.Rows*showtime.SeatsPerRow)
	for _, seatID := range showtime.Seats() {
		section, price := showtime.SectionForSeat(seatID)
		seats = append(seats, SeatView{
			SeatID:     seatID,
			Section:    section,
			PriceCents: price,
			Status:     SeatAvailable,
		})
	}
	bySeat := make(map[string]int, len(seats))
	for i, seat := range seats {
		bySeat[seat.SeatID] = i
	}

	holds, err := s.holds.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	for _, hold := range holds {
		index, ok := bySeat[hold.SeatID]
		if !ok {
			continue
		}
		expiresAt := hold.ExpiresAt
		seats[index] = SeatView{
			SeatID:     hold.SeatID,
			Section:    seats[index].Section,
			PriceCents: seats[index].PriceCents,
			Status:     SeatHeld,
			Mine:       hold.UserID == userID,
			SessionID:  hold.ID,
			ExpiresAt:  &expiresAt,
		}
	}

	bookings, err := s.bookings.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	for _, booking := range bookings {
		index, ok := bySeat[booking.SeatID]
		if !ok {
			continue
		}
		seats[index] = SeatView{
			SeatID:     booking.SeatID,
			Section:    seats[index].Section,
			PriceCents: seats[index].PriceCents,
			Status:     SeatConfirmed,
		}
	}

	return seats, nil
}

func (s *Service) ListUserBookings(ctx context.Context, userID string) ([]Booking, error) {
	return s.bookings.ListByUser(ctx, userID)
}

func (s *Service) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return s.bookings.DeleteOwned(ctx, bookingID, userID)
}

func (s *Service) MyHolds(ctx context.Context, userID string, showtimeID string) ([]Hold, error) {
	holds, err := s.holds.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	mine := []Hold{}
	for _, hold := range holds {
		if hold.UserID == userID {
			mine = append(mine, hold)
		}
	}
	return mine, nil
}
