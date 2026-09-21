package handler_test

import (
	"context"
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

func TestDeleteMe_UnauthenticatedReturns401(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/auth/me", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated DELETE /auth/me = %d, want 401", rec.Code)
	}
}

func TestDeleteMe_DeletesOnlyCurrentUsersDataAndInvalidatesSessions(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	ctx := context.Background()

	user := testutil.CreateUser(t, pool, "reader")
	cookie := testutil.LoginCookieValue(t, pool, user.ID)
	// A second session proves that account deletion invalidates all devices,
	// not merely the cookie used for this request.
	_ = testutil.LoginCookieValue(t, pool, user.ID)
	otherUser := testutil.CreateUser(t, pool, "reader")
	otherCookie := testutil.LoginCookieValue(t, pool, otherUser.ID)

	var quizID, questionID, choiceID, attemptID int64
	if err := pool.QueryRow(ctx, `INSERT INTO quizzes (slug, title, published) VALUES ('account-delete-test', 'Test', true) RETURNING id`).Scan(&quizID); err != nil {
		t.Fatalf("insert quiz: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO quiz_questions (quiz_id, question_text, explanation) VALUES ($1, 'Question', 'Explanation') RETURNING id`, quizID).Scan(&questionID); err != nil {
		t.Fatalf("insert question: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO quiz_choices (question_id, choice_text, is_correct) VALUES ($1, 'Choice', true) RETURNING id`, questionID).Scan(&choiceID); err != nil {
		t.Fatalf("insert choice: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO user_quiz_attempts (user_id, quiz_id, score, total_questions, started_at) VALUES ($1, $2, 1, 1, now()) RETURNING id`, user.ID, quizID).Scan(&attemptID); err != nil {
		t.Fatalf("insert attempt: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_quiz_answers (user_id, question_id, attempt_id, selected_choice_ids, is_correct)
		VALUES ($1, $2, NULL, ARRAY[$3::bigint], true), ($1, $2, $4, ARRAY[$3::bigint], true), ($5, $2, NULL, ARRAY[$3::bigint], true)
	`, user.ID, questionID, choiceID, attemptID, otherUser.ID); err != nil {
		t.Fatalf("insert answers: %v", err)
	}

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/auth/me", cookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /auth/me = %d, want 204: %s", rec.Code, rec.Body.String())
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 1 || cookies[0].Name != "line_omise_session" || cookies[0].MaxAge >= 0 {
		t.Errorf("DELETE /auth/me did not clear session cookie: %+v", cookies)
	}

	for _, check := range []struct {
		name  string
		query string
		args  []any
		want  int
	}{
		{"deleted user", "SELECT count(*) FROM users WHERE id = $1", []any{user.ID}, 0},
		{"deleted sessions", "SELECT count(*) FROM sessions WHERE user_id = $1", []any{user.ID}, 0},
		{"deleted attempts", "SELECT count(*) FROM user_quiz_attempts WHERE user_id = $1", []any{user.ID}, 0},
		{"deleted answers", "SELECT count(*) FROM user_quiz_answers WHERE user_id = $1", []any{user.ID}, 0},
		{"other user remains", "SELECT count(*) FROM users WHERE id = $1", []any{otherUser.ID}, 1},
		{"other answer remains", "SELECT count(*) FROM user_quiz_answers WHERE user_id = $1", []any{otherUser.ID}, 1},
	} {
		var got int
		if err := pool.QueryRow(ctx, check.query, check.args...).Scan(&got); err != nil {
			t.Fatalf("%s query: %v", check.name, err)
		}
		if got != check.want {
			t.Errorf("%s = %d, want %d", check.name, got, check.want)
		}
	}

	if rec := testutil.DoRequest(t, router, http.MethodGet, "/auth/me", cookie, nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /auth/me after deletion = %d, want 401", rec.Code)
	}
	if rec := testutil.DoRequest(t, router, http.MethodGet, "/auth/me", otherCookie, nil); rec.Code != http.StatusOK {
		t.Errorf("GET /auth/me for other user = %d, want 200", rec.Code)
	}
}
