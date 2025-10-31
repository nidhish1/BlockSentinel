package util

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectPostgresWithBackoff attempts to create a pgx pool and ping the database
// with exponential backoff up to maxWait. Returns a ready-to-use pool or error.
func ConnectPostgresWithBackoff(ctx context.Context, dsn string, maxWait time.Duration) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	wait := 500 * time.Millisecond
	started := time.Now()

	for {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				err = pingErr
				pool.Close()
			}
		}

		if time.Since(started) >= maxWait {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}

		if wait < 5*time.Second {
			wait *= 2
		}
	}
}
