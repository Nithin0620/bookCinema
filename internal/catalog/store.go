package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

const showtimeColumns = `
	id::text, movie_id, starts_at, screen_name, rows, seats_per_row,
	premium_rows, premium_price_cents, standard_price_cents
`

func scanShowtime(row scanner) (Showtime, error) {
	var showtime Showtime
	err := row.Scan(
		&showtime.ID,
		&showtime.MovieID,
		&showtime.StartsAt,
		&showtime.ScreenName,
		&showtime.Rows,
		&showtime.SeatsPerRow,
		&showtime.PremiumRows,
		&showtime.PremiumPriceCents,
		&showtime.StandardPriceCents,
	)
	return showtime, err
}

func (s *PostgresStore) ListMovies(ctx context.Context) ([]Movie, error) {
	rows, err := s.db.Query(ctx, `SELECT id, title, description, duration_min, poster_url FROM movies ORDER BY title`)
	if err != nil {
		return nil, fmt.Errorf("list movies: %w", err)
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.DurationMin, &movie.PosterURL); err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}
		movies = append(movies, movie)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movies: %w", err)
	}
	return movies, nil
}

func (s *PostgresStore) MovieByID(ctx context.Context, id string) (Movie, error) {
	var movie Movie
	err := s.db.QueryRow(ctx, `
		SELECT id, title, description, duration_min, poster_url
		FROM movies WHERE id = $1
	`, id).Scan(&movie.ID, &movie.Title, &movie.Description, &movie.DurationMin, &movie.PosterURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return Movie{}, ErrNotFound
	}
	if err != nil {
		return Movie{}, fmt.Errorf("get movie: %w", err)
	}
	return movie, nil
}

func (s *PostgresStore) CreateMovie(ctx context.Context, movie Movie) (Movie, error) {
	err := s.db.QueryRow(ctx, `
		INSERT INTO movies (id, title, description, duration_min, poster_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, duration_min, poster_url
	`, movie.ID, movie.Title, movie.Description, movie.DurationMin, movie.PosterURL).Scan(
		&movie.ID, &movie.Title, &movie.Description, &movie.DurationMin, &movie.PosterURL,
	)
	if isUniqueViolation(err) {
		return Movie{}, ErrConflict
	}
	if err != nil {
		return Movie{}, fmt.Errorf("create movie: %w", err)
	}
	return movie, nil
}

func (s *PostgresStore) ListShowtimes(ctx context.Context, movieID string) ([]Showtime, error) {
	query := `SELECT ` + showtimeColumns + ` FROM showtimes`
	var args []any
	if movieID != "" {
		query += ` WHERE movie_id = $1`
		args = append(args, movieID)
	}
	query += ` ORDER BY starts_at`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list showtimes: %w", err)
	}
	defer rows.Close()

	showtimes := []Showtime{}
	for rows.Next() {
		showtime, err := scanShowtime(rows)
		if err != nil {
			return nil, fmt.Errorf("scan showtime: %w", err)
		}
		showtimes = append(showtimes, showtime)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate showtimes: %w", err)
	}
	return showtimes, nil
}

func (s *PostgresStore) ShowtimeByID(ctx context.Context, id string) (Showtime, error) {
	showtime, err := scanShowtime(s.db.QueryRow(ctx, `SELECT `+showtimeColumns+` FROM showtimes WHERE id = $1::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Showtime{}, ErrNotFound
	}
	if err != nil {
		return Showtime{}, fmt.Errorf("get showtime: %w", err)
	}
	return showtime, nil
}

func (s *PostgresStore) CreateShowtime(ctx context.Context, showtime Showtime) (Showtime, error) {
	if showtime.ID == "" {
		showtime.ID = uuid.NewString()
	}
	created, err := scanShowtime(s.db.QueryRow(ctx, `
		INSERT INTO showtimes (
			id, movie_id, starts_at, screen_name, rows, seats_per_row,
			premium_rows, premium_price_cents, standard_price_cents
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+showtimeColumns+`
	`,
		showtime.ID,
		showtime.MovieID,
		showtime.StartsAt,
		showtime.ScreenName,
		showtime.Rows,
		showtime.SeatsPerRow,
		showtime.PremiumRows,
		showtime.PremiumPriceCents,
		showtime.StandardPriceCents,
	))
	if isUniqueViolation(err) {
		return Showtime{}, ErrConflict
	}
	if err != nil {
		return Showtime{}, fmt.Errorf("create showtime: %w", err)
	}
	return created, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}

var _ Store = (*PostgresStore)(nil)
