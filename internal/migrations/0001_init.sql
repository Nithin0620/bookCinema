-- Phase 1 schema. Safe to run on every startup.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

CREATE TABLE IF NOT EXISTS movies (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_min INTEGER NOT NULL DEFAULT 0,
    poster_url TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS showtimes (
    id UUID PRIMARY KEY,
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    screen_name TEXT NOT NULL,
    rows INTEGER NOT NULL CHECK (rows BETWEEN 1 AND 26),
    seats_per_row INTEGER NOT NULL CHECK (seats_per_row BETWEEN 1 AND 30),
    price_cents INTEGER NOT NULL CHECK (price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS idx_showtimes_movie_id ON showtimes(movie_id, starts_at);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    showtime_id UUID NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
    seat_id TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    price_cents INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bookings_showtime_seat_key UNIQUE (showtime_id, seat_id)
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bookings_showtime_id ON bookings(showtime_id);

INSERT INTO movies (id, title, description, duration_min, poster_url) VALUES
    ('inception', 'Inception', 'A thief who steals corporate secrets through dream-sharing technology.', 148, ''),
    ('dune-part-two', 'Dune: Part Two', 'Paul Atreides unites with the Fremen to wage war against the conspirators.', 166, '')
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'showtimes' AND column_name = 'price_cents'
    ) THEN
        EXECUTE '
            INSERT INTO showtimes (id, movie_id, starts_at, screen_name, rows, seats_per_row, price_cents) VALUES
                (''11111111-1111-1111-1111-111111111111'', ''inception'', now() + interval ''2 hours'', ''Screen 1'', 5, 8, 1200),
                (''22222222-2222-2222-2222-222222222222'', ''dune-part-two'', now() + interval ''4 hours'', ''Screen 2'', 4, 6, 1500)
            ON CONFLICT (id) DO NOTHING
        ';
    ELSE
        INSERT INTO showtimes (
            id, movie_id, starts_at, screen_name, rows, seats_per_row,
            premium_rows, premium_price_cents, standard_price_cents
        ) VALUES
            ('11111111-1111-1111-1111-111111111111', 'inception', now() + interval '2 hours', 'Screen 1', 5, 8, 2, 1800, 1200),
            ('22222222-2222-2222-2222-222222222222', 'dune-part-two', now() + interval '4 hours', 'Screen 2', 4, 6, 1, 2000, 1500)
        ON CONFLICT (id) DO NOTHING;
    END IF;
END
$$;
