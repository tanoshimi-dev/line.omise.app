// Package database manages the database connection pool.
//
// NOTE: the DB engine itself is decided in dev-plan-02-database.md (not yet
// implemented). Until then, this package only tracks whether DATABASE_URL is
// configured and reachable at the TCP level, so /health has something
// meaningful to report without hard-depending on a driver/schema that
// doesn't exist yet.
package database

import (
	"context"
	"net"
	"strings"
	"time"
)

// Status describes the current reachability of the database, as far as this
// stub can tell without a real driver.
type Status struct {
	Configured bool
	Reachable  bool
	Detail     string
}

// Check reports the database status for the given DATABASE_URL. It never
// returns an error itself — callers (e.g. the health handler) decide how to
// react to an unreachable or unconfigured database.
func Check(ctx context.Context, databaseURL string) Status {
	if databaseURL == "" {
		return Status{Configured: false, Detail: "DATABASE_URL not set (expected until dev-plan-02-database lands)"}
	}

	host, ok := hostPort(databaseURL)
	if !ok {
		return Status{Configured: true, Detail: "DATABASE_URL set but host:port could not be parsed"}
	}

	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return Status{Configured: true, Reachable: false, Detail: err.Error()}
	}
	_ = conn.Close()
	return Status{Configured: true, Reachable: true}
}

// hostPort extracts the host:port portion of a DSN shaped like
// postgres://user:pass@host:port/dbname. It's intentionally minimal — a real
// driver/connection pool replaces this in dev-plan-02-database.
func hostPort(dsn string) (string, bool) {
	afterScheme := dsn
	if _, rest, ok := strings.Cut(dsn, "://"); ok {
		afterScheme = rest
	}
	afterAt := afterScheme
	if idx := strings.LastIndex(afterScheme, "@"); idx != -1 {
		afterAt = afterScheme[idx+1:]
	}
	end := strings.IndexAny(afterAt, "/?")
	if end != -1 {
		afterAt = afterAt[:end]
	}
	if afterAt == "" || !strings.Contains(afterAt, ":") {
		return "", false
	}
	return afterAt, true
}
