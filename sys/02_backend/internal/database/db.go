// Package database manages the PostgreSQL connection pool (dev-plan-02-database).
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps a pgx connection pool. A Pool with no underlying pgxpool.Pool
// (Configured() == false) is valid and represents "no DATABASE_URL set" —
// callers such as /health degrade gracefully instead of failing to start.
type Pool struct {
	pool *pgxpool.Pool
}

// Connect creates a connection pool for the given DSN. Connecting is lazy
// (pgxpool does not dial until first use), so an unreachable database does
// not fail startup — Check reports reachability for /health instead.
func Connect(ctx context.Context, databaseURL string) (*Pool, error) {
	if databaseURL == "" {
		return &Pool{}, nil
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Pool{pool: pool}, nil
}

// Configured reports whether a DATABASE_URL was provided.
func (p *Pool) Configured() bool {
	return p != nil && p.pool != nil
}

// DB returns the underlying pgx pool for use by repositories (dev-plan-05/06).
func (p *Pool) DB() *pgxpool.Pool {
	return p.pool
}

// Close releases all pooled connections. Safe to call on an unconfigured Pool.
func (p *Pool) Close() {
	if p.Configured() {
		p.pool.Close()
	}
}

// Status describes the current reachability of the database.
type Status struct {
	Configured bool
	Reachable  bool
	Detail     string
}

// Check pings the database, bounded by a short timeout so /health stays fast.
func Check(ctx context.Context, p *Pool) Status {
	if !p.Configured() {
		return Status{Detail: "DATABASE_URL not set"}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := p.pool.Ping(ctx); err != nil {
		return Status{Configured: true, Reachable: false, Detail: err.Error()}
	}
	return Status{Configured: true, Reachable: true}
}
