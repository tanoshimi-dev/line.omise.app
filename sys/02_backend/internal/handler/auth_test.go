package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestMe_UnauthenticatedReturns401(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/auth/me", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated GET /auth/me = %d, want 401", rec.Code)
	}
}

func TestMe_ReturnsLoggedInUser(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	user := testutil.CreateUser(t, pool, "admin")
	cookie := testutil.LoginCookieValue(t, pool, user.ID)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/auth/me", cookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /auth/me = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Role != "admin" {
		t.Errorf("role = %q, want admin", body.Role)
	}
}

func TestLogout_ClearsSessionSoSubsequentMeFails(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	user := testutil.CreateUser(t, pool, "reader")
	cookie := testutil.LoginCookieValue(t, pool, user.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/auth/logout", cookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /auth/logout = %d, want 204", rec.Code)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/auth/me", cookie, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /auth/me after logout = %d, want 401", rec.Code)
	}
}
