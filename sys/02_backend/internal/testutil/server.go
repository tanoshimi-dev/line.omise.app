package testutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/config"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/server"
)

// NewRouter builds the real application router (internal/server.New) wired
// against the shared TestDB pool — the same router cmd/server/main.go runs,
// so handler tests exercise the actual auth middleware chain (401/403)
// rather than a mocked substitute.
func NewRouter(t *testing.T, pool *pgxpool.Pool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		SessionSecret: TestSessionSecret,
		Env:           "test",
	}
	dbPool := database.NewPoolForTest(pool)
	return server.New(cfg, dbPool)
}

// DoRequest performs an in-process HTTP request against router (no real
// network) and returns the recorded response. cookie may be "" for an
// unauthenticated request; body may be nil for a bodyless request.
func DoRequest(t *testing.T, router *gin.Engine, method, path, cookie string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		t.Fatalf("testutil: failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "line_omise_session", Value: cookie})
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
