package catalog

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

type Movie struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DurationMin int    `json:"duration_min"`
	PosterURL   string `json:"poster_url"`
}

type Showtime struct {
	ID                 string    `json:"id"`
	MovieID            string    `json:"movie_id"`
	StartsAt           time.Time `json:"starts_at"`
	ScreenName         string    `json:"screen_name"`
	Rows               int       `json:"rows"`
	SeatsPerRow        int       `json:"seats_per_row"`
	PremiumRows        int       `json:"premium_rows"`
	PremiumPriceCents  int       `json:"premium_price_cents"`
	StandardPriceCents int       `json:"standard_price_cents"`
}

type Store interface {
	ListMovies(ctx context.Context) ([]Movie, error)
	MovieByID(ctx context.Context, id string) (Movie, error)
	CreateMovie(ctx context.Context, movie Movie) (Movie, error)
	ListShowtimes(ctx context.Context, movieID string) ([]Showtime, error)
	ShowtimeByID(ctx context.Context, id string) (Showtime, error)
	CreateShowtime(ctx context.Context, showtime Showtime) (Showtime, error)
}

func (s Showtime) ValidSeat(seatID string) bool {
	if len(seatID) < 2 {
		return false
	}
	row := int(seatID[0] - 'A')
	column, err := strconv.Atoi(seatID[1:])
	if err != nil {
		return false
	}
	return row >= 0 && row < s.Rows && column >= 1 && column <= s.SeatsPerRow
}

func (s Showtime) Seats() []string {
	seats := make([]string, 0, s.Rows*s.SeatsPerRow)
	for row := 0; row < s.Rows; row++ {
		for column := 1; column <= s.SeatsPerRow; column++ {
			seats = append(seats, fmt.Sprintf("%c%d", 'A'+row, column))
		}
	}
	return seats
}

func (s Showtime) SectionForSeat(seatID string) (string, int) {
	if len(seatID) > 0 && int(seatID[0]-'A') < s.PremiumRows {
		return "premium", s.PremiumPriceCents
	}
	return "standard", s.StandardPriceCents
}
