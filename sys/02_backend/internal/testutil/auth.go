package testutil

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

// TestSessionSecret is used by every test that needs to sign a session
// cookie — arbitrary, since tests never talk to a real deployment.
const TestSessionSecret = "test-session-secret"

var userCounter atomic.Int64

// CreateUser inserts a test user with the given role ("admin" or "reader")
// and a unique provider_user_id, so multiple calls within one test never
// collide.
func CreateUser(t *testing.T, pool *pgxpool.Pool, role string) *repository.User {
	t.Helper()
	n := userCounter.Add(1)
	repo := repository.NewUserRepository(pool)
	user, err := repo.UpsertByProvider(context.Background(), "google",
		fmt.Sprintf("test-user-%d", n),
		fmt.Sprintf("test-user-%d@example.com", n),
		fmt.Sprintf("Test User %d", n),
		"",
		role == "admin",
	)
	if err != nil {
		t.Fatalf("testutil: failed to create test user: %v", err)
	}
	return user
}

// LoginCookieValue creates a session for userID and returns the signed
// cookie value expected by internal/middleware — pass it as the
// line_omise_session cookie on a test request.
func LoginCookieValue(t *testing.T, pool *pgxpool.Pool, userID int64) string {
	t.Helper()
	sessions := repository.NewSessionRepository(pool)
	sess, err := sessions.Create(context.Background(), userID)
	if err != nil {
		t.Fatalf("testutil: failed to create test session: %v", err)
	}
	return session.Sign(sess.ID, TestSessionSecret)
}
