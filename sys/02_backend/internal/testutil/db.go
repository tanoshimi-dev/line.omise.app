// Package testutil provides shared test infrastructure for repository,
// service and handler tests (dev-plan-12-test-phase1 12.1): a real Postgres
// instance via testcontainers-go (migrated with the project's actual
// migration files, not a hand-maintained schema copy), plus helpers for
// creating authenticated test requests.
package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	once     sync.Once
	pool     *pgxpool.Pool
	initErr  error
	dbTables = []string{
		"user_quiz_answers", "user_quiz_attempts",
		"quiz_choices", "quiz_questions", "quizzes",
		"article_tags", "tags", "articles",
		"usecases",
		"sessions", "users",
	}
)

// TestDB returns a shared Postgres pool for the whole test binary run,
// starting the container and applying migrations only once (container
// startup takes several seconds — sharing it across a package's tests
// keeps the suite fast). Each test should call TruncateAll via t.Cleanup
// to avoid leaking rows into the next test.
func TestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	once.Do(func() {
		pool, initErr = startContainer()
	})
	if initErr != nil {
		t.Fatalf("testutil: failed to start test database: %v", initErr)
	}
	return pool
}

// TruncateAll clears every table so tests don't see rows left by others
// sharing the same TestDB container.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE "+strings.Join(dbTables, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("testutil: failed to truncate tables: %v", err)
	}
}

func startContainer() (*pgxpool.Pool, error) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("line_omise_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	if err := applyMigrations(connStr); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("connect pool: %w", err)
	}
	return pool, nil
}

// migrationsDir resolves sys/02_backend/migrations relative to this source
// file, so it works regardless of which package's test binary is running
// (each has a different working directory).
func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

// applyMigrations retries the initial connection: testcontainers' wait
// strategy reports the container "ready" based on a log line, but on
// Docker Desktop/Windows there's sometimes a brief window after that where
// Postgres accepts the TCP connection and then closes it (EOF) because it's
// still cycling through its startup restart. A few short retries clears it.
func applyMigrations(databaseURL string) error {
	var lastErr error
	for i := 0; i < 10; i++ {
		m, err := migrate.New("file://"+filepath.ToSlash(migrationsDir()), databaseURL)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		err = m.Up()
		m.Close()
		if err != nil && err != migrate.ErrNoChange {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}
