package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestPublicListCourses_ExcludesDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie,
		mustJSON(t, map[string]any{"slug": "published-course", "title": "Published", "status": "published"}))
	testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie,
		mustJSON(t, map[string]any{"slug": "draft-course", "title": "Draft", "status": "draft"}))

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/courses", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/courses = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Courses []map[string]any `json:"courses"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(body.Courses) != 1 || body.Courses[0]["slug"] != "published-course" {
		t.Errorf("public course list = %+v, want only the published course", body.Courses)
	}
}

// Regression test for the bug found while implementing dev-plan-05: RequireAdmin
// used to call RequireReader as a plain function, whose internal c.Next()
// cascaded into the real handler *before* the admin check ran — so a Reader
// could actually execute an admin-only write. This test reproduces the exact
// scenario and checks both the HTTP status AND that no row was created.
func TestAdminCreateCourse_ReaderIsRejectedAndNoRowIsCreated(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", readerCookie,
		mustJSON(t, map[string]any{"slug": "should-not-exist", "title": "Should Not Exist", "status": "draft"}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Reader POST /api/admin/courses = %d, want 403", rec.Code)
	}

	var count int
	err := pool.QueryRow(context.Background(), "SELECT count(*) FROM courses WHERE slug = 'should-not-exist'").Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("course row count = %d, want 0 (a 403 must never have a write side effect)", count)
	}
}

func TestAdminCreateCourse_UnauthenticatedGets401(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", "",
		mustJSON(t, map[string]any{"slug": "x", "title": "x", "status": "draft"}))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated POST /api/admin/courses = %d, want 401", rec.Code)
	}
}

func TestAdminCreateCourse_AdminSucceeds(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie,
		mustJSON(t, map[string]any{"slug": "new-course", "title": "New Course", "status": "draft"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin POST /api/admin/courses = %d, want 201: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateCourse_DuplicateSlugConflict(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	body := mustJSON(t, map[string]any{"slug": "dup", "title": "First", "status": "draft"})
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie,
		mustJSON(t, map[string]any{"slug": "dup", "title": "Second", "status": "draft"}))
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate slug create = %d, want 409", rec.Code)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	return b
}
