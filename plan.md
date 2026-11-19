# bookCinema — Phase 1 Full-Stack Build Plan

## Current state
- Go HTTP service, Redis-only store, hardcoded movies, single self-contained `static/index.html`.
- Redis `SET NX` gives the useful concurrency guarantee: exactly one winner per seat hold.
- Gaps this plan closes: no auth, no durable DB, client-supplied `user_id`, no ownership check in `getSession`, no HTTP error semantics, port mismatch (app `:6379` vs compose `:6380`), empty README, stale in-memory stores.

## Target architecture
- Redis stays the source of truth for **short-lived holds** (2 min TTL).
- Postgres becomes the source of truth for **users, sessions, movies, showtimes, confirmed bookings**.
- App is a single Go binary serving JSON API + static frontend.

```
cmd/api/main.go
internal/config/          env parsing, defaults, validation
internal/platform/postgres/  pgx pool + health check
internal/platform/redis/     client + Lua scripts
internal/auth/            register/login/logout, sessions, bcrypt, middleware
internal/catalog/         movies, showtimes, seat layout
internal/booking/         domain, service, Redis hold store, Postgres booking store
internal/httpapi/         router, middleware, errors, DTOs, handlers
internal/migrations/      0001_init.sql (embedded, idempotent)
web/                      Vite + React + TypeScript frontend (built to web/dist)
tests/                    unit + integration helpers
```

## Database schema (`internal/migrations/0001_init.sql`)
- `users`: id uuid, email unique, password_hash, created_at
- `sessions`: token hash (primary key), user_id, expires_at
- `movies`: id slug, title, description, duration_min, poster_url
- `showtimes`: id uuid, movie_id, starts_at, screen_name, rows, seats_per_row, price_cents
- `bookings`: id uuid, showtime_id, seat_id, user_id, price_cents, created_at, **unique(showtime_id, seat_id)**

## Redis keys (showtime-scoped)
- `hold:{showtimeID}:{seatID}` → hold JSON, TTL 2 min
- `session:{sessionID}` → hold key, TTL 2 min
- Lua scripts:
  - `hold`: atomic `SET NX EX` + reverse index; returns session id + `expires_at`
  - `validate`: verify session ownership; returns hold payload without mutating
  - `delete`: remove hold + reverse index after successful checkout
  - `release`: ownership check + delete in one step
- Confirm flow: validate owner in Redis → insert bookings in one Postgres transaction (unique constraint is final authority) → delete holds in Redis. Seat map merges Redis holds + Postgres confirmed seats; Postgres wins.

## API endpoints (all under `/api`)
- Auth: `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/auth/me`
- Catalog: `GET /api/movies`, `GET /api/movies/{id}`, `GET /api/showtimes/{id}`, `GET /api/showtimes/{id}/seats`
- Booking: `POST /api/showtimes/{id}/seats/{seatID}/hold`, `POST /api/checkout` (`session_ids: []`), `DELETE /api/sessions/{id}`, `GET /api/me/bookings`, `DELETE /api/bookings/{id}`
- Ops: `GET /api/healthz`
- All errors use `{"error": "code_or_message"}` with real status codes (400/401/403/404/409/500).

## Auth
- Opaque random session token in `HttpOnly; SameSite=Lax` cookie (secure flag configurable), hashed in Postgres `sessions`.
- `bcrypt` password hashing. No JWT needed for phase 1; revocation is a DELETE.
- Middleware resolves user from cookie; handlers use `user.ID` only — never trust request bodies for identity.
- Ownership enforced in every session/booking operation.

## Frontend (React)
- Vite + React + TypeScript app in `web/`, built to `web/dist`, served by the Go binary at `/`.
- Vite dev server proxies `/api` to `http://localhost:8080`; same-origin cookie auth works in both modes.
- React Router pages: movies, showtime seat map, my bookings, login, register.
- `AuthProvider` resolves the current user via `GET /api/auth/me`; protected routes redirect to login.
- Movie → showtime → seat map → cart (multiple holds) → checkout flow.
- Hold countdown calls `DELETE /api/sessions/{id}` on expiry; 2s polling for live seat state (SSE/WebSocket deferred).
- Replace the old inline-JS `static/index.html` with the React build.

## Infra & tooling
- Multi-stage `Dockerfile`: Node stage builds `web/dist`, Go stage builds the binary, final image contains both.
- Compose services: `app`, `postgres`, `redis`, `redis-commander`; named volumes; healthchecks; `.env.example`.
- `config` reads `APP_ADDR`, `DATABASE_URL`, `REDIS_ADDR`, `HOLD_TTL`, `SESSION_TTL`, `COOKIE_SECURE`.
- Graceful shutdown, server timeouts, recover/logging middleware.
- README: setup, env vars, API overview, how to run tests.
- Delete or fix the stale `MemoryStore` / `ConcurentStore` so only one store contract exists.

## Tests
- Unit: auth service, booking service, handlers with in-memory fakes, error mapping.
- Integration (build tag `integration`): Postgres + Redis from compose; hold/concurrency, ownership, checkout transaction, expiry.
- Replace the 100k-goroutine live-Redis test with a bounded concurrency test that runs in CI.
- Verify with `go test ./...`, `go vet ./...`, `gofmt -l .`, and `npm --prefix web run build`.

## Out of scope for phase 1
Payments, tickets/QR, email/SMS, admin UI, websockets, seat pricing tiers, CI/CD pipeline.

## Estimate
~3–5 hours of assisted iteration (config/DB/auth → API → Redis flow → frontend → tests/docs), or ~3–4 days solo.

## Approval gate
I will not create or modify any code files until you approve this plan.
