package postgres

import (
	"context"
	"fmt"

	"github.com/Nithin0620/bookCinema/internal/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	schemas, err := migrations.Schemas()
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	if err := applyMigrations(ctx, pool, schemas); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, schemas []string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}

	const migrationLock int64 = 7241982736401
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLock); err != nil {
		conn.Release()
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationLock)
		conn.Release()
	}()

	for _, schema := range schemas {
		if _, err := conn.Exec(ctx, schema); err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}
	}
	return nil
}
