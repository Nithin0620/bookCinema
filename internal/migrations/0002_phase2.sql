ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE showtimes
    ADD COLUMN IF NOT EXISTS premium_rows INTEGER NOT NULL DEFAULT 1 CHECK (premium_rows BETWEEN 1 AND 26),
    ADD COLUMN IF NOT EXISTS premium_price_cents INTEGER NOT NULL DEFAULT 0 CHECK (premium_price_cents >= 0),
    ADD COLUMN IF NOT EXISTS standard_price_cents INTEGER NOT NULL DEFAULT 0 CHECK (standard_price_cents >= 0);

UPDATE showtimes
SET premium_rows = 2, premium_price_cents = 1800, standard_price_cents = 1200
WHERE id = '11111111-1111-1111-1111-111111111111'
  AND premium_price_cents = 0 AND standard_price_cents = 0;

UPDATE showtimes
SET premium_rows = 1, premium_price_cents = 2000, standard_price_cents = 1500
WHERE id = '22222222-2222-2222-2222-222222222222'
  AND premium_price_cents = 0 AND standard_price_cents = 0;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'showtimes' AND column_name = 'price_cents'
    ) THEN
        EXECUTE '
            UPDATE showtimes
            SET premium_price_cents = price_cents,
                standard_price_cents = price_cents
            WHERE premium_price_cents = 0 AND standard_price_cents = 0
        ';
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY,
    booking_id UUID NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tickets_booking_id ON tickets(booking_id);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id, created_at DESC);

ALTER TABLE showtimes DROP COLUMN IF EXISTS price_cents;
