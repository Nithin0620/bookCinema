package booking

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBookingStore struct {
	db *pgxpool.Pool
}

func NewPostgresBookingStore(db *pgxpool.Pool) *PostgresBookingStore {
	return &PostgresBookingStore{db: db}
}

func (s *PostgresBookingStore) CreateMany(ctx context.Context, bookings []Booking) ([]Booking, error) {
	if len(bookings) == 0 {
		return []Booking{}, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin checkout: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, booking := range bookings {
		tag, err := tx.Exec(ctx, `
			INSERT INTO bookings (id, showtime_id, seat_id, user_id, price_cents, created_at)
			VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5, $6)
			ON CONFLICT (showtime_id, seat_id) DO NOTHING
		`, booking.ID, booking.ShowtimeID, booking.SeatID, booking.UserID, booking.PriceCents, booking.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert booking: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, ErrSeatAlreadyBooked
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit checkout: %w", err)
	}
	return bookings, nil
}

func (s *PostgresBookingStore) Exists(ctx context.Context, showtimeID, seatID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM bookings WHERE showtime_id = $1::uuid AND seat_id = $2
		)
	`, showtimeID, seatID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check booking: %w", err)
	}
	return exists, nil
}

func (s *PostgresBookingStore) ListByShowtime(ctx context.Context, showtimeID string) ([]Booking, error) {
	return s.list(ctx, `
		SELECT id::text, showtime_id::text, seat_id, user_id::text, price_cents, created_at
		FROM bookings WHERE showtime_id = $1::uuid ORDER BY seat_id
	`, showtimeID)
}

func (s *PostgresBookingStore) ListByUser(ctx context.Context, userID string) ([]Booking, error) {
	return s.list(ctx, `
		SELECT id::text, showtime_id::text, seat_id, user_id::text, price_cents, created_at
		FROM bookings WHERE user_id = $1::uuid ORDER BY created_at DESC
	`, userID)
}

func (s *PostgresBookingStore) list(ctx context.Context, query string, args ...any) ([]Booking, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings := []Booking{}
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(
			&booking.ID,
			&booking.ShowtimeID,
			&booking.SeatID,
			&booking.UserID,
			&booking.PriceCents,
			&booking.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}
	return bookings, nil
}

func (s *PostgresBookingStore) DeleteOwned(ctx context.Context, bookingID, userID string) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM bookings WHERE id = $1::uuid AND user_id = $2::uuid
	`, bookingID, userID)
	if err != nil {
		return fmt.Errorf("delete booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookingNotFound
	}
	return nil
}

var _ BookingStore = (*PostgresBookingStore)(nil)
