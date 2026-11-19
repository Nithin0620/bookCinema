package ticket

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/Nithin0620/bookCinema/internal/booking"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("ticket not found")

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateForBookings(ctx context.Context, bookings []booking.Booking) ([]Ticket, error) {
	tickets := make([]Ticket, 0, len(bookings))
	for _, reservation := range bookings {
		code, err := newCode()
		if err != nil {
			return nil, err
		}
		var ticket Ticket
		err = s.db.QueryRow(ctx, `
			INSERT INTO tickets (id, booking_id, code)
			VALUES ($1::uuid, $2::uuid, $3)
			RETURNING id::text, booking_id::text, code, created_at
		`, uuid.NewString(), reservation.ID, code).Scan(
			&ticket.ID,
			&ticket.BookingID,
			&ticket.Code,
			&ticket.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("create ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID string) ([]Detail, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			t.id::text, t.booking_id::text, t.code, t.created_at,
			b.id::text, b.showtime_id::text, b.seat_id, b.user_id::text, b.price_cents, b.created_at
		FROM tickets t
		JOIN bookings b ON b.id = t.booking_id
		WHERE b.user_id = $1::uuid
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	details := []Detail{}
	for rows.Next() {
		var detail Detail
		if err := rows.Scan(
			&detail.Ticket.ID,
			&detail.Ticket.BookingID,
			&detail.Ticket.Code,
			&detail.Ticket.CreatedAt,
			&detail.Booking.ID,
			&detail.Booking.ShowtimeID,
			&detail.Booking.SeatID,
			&detail.Booking.UserID,
			&detail.Booking.PriceCents,
			&detail.Booking.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tickets: %w", err)
	}
	return details, nil
}

func (s *PostgresStore) GetOwned(ctx context.Context, ticketID, userID string) (Detail, error) {
	var detail Detail
	err := s.db.QueryRow(ctx, `
		SELECT
			t.id::text, t.booking_id::text, t.code, t.created_at,
			b.id::text, b.showtime_id::text, b.seat_id, b.user_id::text, b.price_cents, b.created_at
		FROM tickets t
		JOIN bookings b ON b.id = t.booking_id
		WHERE t.id = $1::uuid AND b.user_id = $2::uuid
	`, ticketID, userID).Scan(
		&detail.Ticket.ID,
		&detail.Ticket.BookingID,
		&detail.Ticket.Code,
		&detail.Ticket.CreatedAt,
		&detail.Booking.ID,
		&detail.Booking.ShowtimeID,
		&detail.Booking.SeatID,
		&detail.Booking.UserID,
		&detail.Booking.PriceCents,
		&detail.Booking.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Detail{}, ErrNotFound
	}
	if err != nil {
		return Detail{}, fmt.Errorf("get ticket: %w", err)
	}
	return detail, nil
}

func newCode() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

var _ Store = (*PostgresStore)(nil)
