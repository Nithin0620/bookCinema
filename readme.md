# bookCinema

A cinema seat-booking app. Redis holds seats for two minutes with an atomic Lua/`SET NX` lock; Postgres stores users, sessions, movies, showtimes, confirmed bookings, tickets, and notifications. Seats are split into premium and standard sections with fixed prices. Checkout uses a one-time math puzzle instead of real payments. The API is Go, the frontend is React + Vite + TypeScript.

## Run everything

```bash
docker compose up --build
```

Open http://localhost:8080. Redis Commander is at http://localhost:8081.

The Go app applies all SQL files in `internal/migrations/` on startup and seeds two movies with one showtime each. The seeded showtimes have 2 and 1 premium rows respectively; the rest are standard. Register with an email listed in `ADMIN_EMAILS` to get admin access.

## Run locally

Start only the backing services:

```bash
docker compose up -d postgres redis redis-commander
```

Build the frontend and run the API:

```bash
npm --prefix web install
npm --prefix web run build
cp .env.example .env
go run ./cmd/api
```

For frontend hot reload, run `npm --prefix web run dev` and keep the API on `:8080`; Vite proxies `/api` to it.

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | HTTP listen address |
| `DATABASE_URL` | required | Postgres connection string |
| `REDIS_ADDR` | `localhost:6380` | Redis address (compose publishes 6380) |
| `HOLD_TTL` | `2m` | Seat hold duration |
| `SESSION_TTL` | `168h` | Login session duration |
| `COOKIE_SECURE` | `false` | Set `Secure` on the session cookie |
| `STATIC_DIR` | `web/dist` | Built React assets |
| `ADMIN_EMAILS` | empty | Comma-separated emails promoted to admin on registration |

## API

All routes are under `/api`.

| Method | Route | Auth | Description |
| --- | --- | --- | --- |
| POST | `/auth/register` | no | Create account, set session cookie |
| POST | `/auth/login` | no | Log in, set session cookie |
| POST | `/auth/logout` | yes | Revoke session |
| GET | `/auth/me` | yes | Current user |
| GET | `/movies` | no | Movie catalog |
| GET | `/showtimes?movie_id=` | no | Showtimes |
| GET | `/showtimes/{id}/seats` | no | Seat map (own holds marked when logged in) |
| POST | `/showtimes/{id}/seats/{seatID}/hold` | yes | Hold one seat |
| POST | `/payment/challenge` | yes | Get one-time confirmation puzzle |
| POST | `/checkout` | yes | Confirm holds with puzzle: `{ "session_ids": ["..."], "payment": { "id": "...", "answer": "..." } }` |
| DELETE | `/sessions/{id}` | yes | Release a hold |
| GET | `/me/bookings` | yes | Confirmed bookings |
| DELETE | `/bookings/{id}` | yes | Cancel a confirmed booking |
| GET | `/me/tickets` | yes | Tickets for the current user |
| GET | `/tickets/{id}/qr.png` | yes | Ticket QR code PNG |
| GET | `/me/notifications` | yes | Notifications for the current user |
| POST | `/me/notifications/read` | yes | Mark all notifications read |
| POST | `/admin/movies` | admin | Create a movie |
| POST | `/admin/showtimes` | admin | Schedule a showtime |
| GET | `/healthz` | no | Health check |

Errors use `{"error": "code"}` with real HTTP status codes.

## Design notes

- Ownership is enforced server-side from the session cookie; the browser never sends a `user_id`.
- Confirm inserts into Postgres inside a transaction with a unique `(showtime_id, seat_id)` constraint, so the database is the final authority even if two clients race.
- Redis holds are removed after a successful checkout; the seat map merges Redis holds with Postgres bookings.
- Expired holds are deleted by the frontend on the next poll; Redis TTL is the real expiry boundary.
- Puzzle answers are stored in Redis with a short TTL and deleted atomically on first verification, so each puzzle can only be used once.
- One ticket is created per confirmed booking; the QR code encodes a unique ticket code.
- Confirmation and cancellation write notifications for the booking user.

## Tests

```bash
gofmt -l .
go vet ./...
go test ./...
npm --prefix web run build
```

Integration test (requires Postgres and Redis on the compose ports):

```bash
DATABASE_URL='postgres://cinema:cinema@localhost:5433/cinema?sslmode=disable' \
REDIS_ADDR='localhost:6380' \
go test -tags=integration ./internal/booking -run TestIntegration -v
```

See `plan.md` for the agreed phase-1 scope; phase 2 adds section pricing, puzzle checkout, tickets, notifications, and admin CRUD.
