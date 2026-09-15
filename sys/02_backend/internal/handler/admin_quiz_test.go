package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestAdminCreateQuiz_ReaderIsRejected(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", readerCookie,
		mustJSON(t, map[string]any{"slug": "should-not-exist", "title": "x"}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Reader POST /api/admin/quizzes = %d, want 403", rec.Code)
	}
}

func TestAdminCreateQuiz_UnauthenticatedGets401(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", "",
		mustJSON(t, map[string]any{"slug": "x", "title": "x"}))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated POST /api/admin/quizzes = %d, want 401", rec.Code)
	}
}

func TestAdminCreateQuiz_WithoutPassingScore(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "line-quiz", "title": "LINE運用クイズ", "published": true}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin POST /api/admin/quizzes = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["passing_score"] != nil {
		t.Errorf("passing_score = %v, want nil when not set", body["passing_score"])
	}
}

func TestAdminCreateQuiz_WithPassingScore(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "basic-line-exam", "title": "基礎LINE検定", "passing_score": 80, "published": true}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin POST /api/admin/quizzes = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["passing_score"] != float64(80) {
		t.Errorf("passing_score = %v, want 80", body["passing_score"])
	}
}

func TestAdminCreateQuiz_DuplicateSlugConflict(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)

	body := mustJSON(t, map[string]any{"slug": "dup-quiz", "title": "First"})
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create = %d, want 201: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "dup-quiz", "title": "Second"}))
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate slug create = %d, want 409", rec.Code)
	}
}

func createTestQuiz(t *testing.T, router *gin.Engine, adminCookie string) string {
	t.Helper()
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "quiz", "title": "Quiz", "published": true}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quiz = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return body["id"].(string)
}

func TestAdminCreateQuestion_TooFewChoicesRejected(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "Q1?",
			"explanation":   "exp",
			"choices":       []map[string]any{{"choice_text": "A", "is_correct": true}},
		}))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("create question with 1 choice = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateQuestion_NoCorrectChoiceRejected(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "Q1?",
			"explanation":   "exp",
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": false},
				{"choice_text": "B", "is_correct": false},
			},
		}))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("create question with no correct choice = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateQuestion_MultipleCorrectRejectedWhenNotAllowMultiple(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text":  "Q1?",
			"explanation":    "exp",
			"allow_multiple": false,
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": true},
				{"choice_text": "B", "is_correct": true},
			},
		}))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("create question with 2 correct choices and allow_multiple=false = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateQuestion_MultipleCorrectAllowedWhenAllowMultiple(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text":  "Q1?",
			"explanation":    "exp",
			"allow_multiple": true,
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": true},
				{"choice_text": "B", "is_correct": true},
			},
		}))
	if rec.Code != http.StatusCreated {
		t.Errorf("create question with 2 correct choices and allow_multiple=true = %d, want 201: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminGetQuiz_EmbedsQuestionsWithIsCorrect(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "Q1?",
			"explanation":   "exp",
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": true},
				{"choice_text": "B", "is_correct": false},
			},
		}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create question = %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/admin/quizzes/"+quizID, adminCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/admin/quizzes/:id = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Questions []struct {
			Choices []struct {
				IsCorrect bool `json:"is_correct"`
			} `json:"choices"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Questions) != 1 || len(body.Questions[0].Choices) != 2 {
		t.Fatalf("unexpected shape: %+v", body)
	}
	if !body.Questions[0].Choices[0].IsCorrect {
		t.Errorf("admin response should carry is_correct=true for the first choice")
	}
}

func TestAdminDeleteQuiz_CascadesQuestions(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID := createTestQuiz(t, router, adminCookie)

	testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "Q1?",
			"explanation":   "exp",
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": true},
				{"choice_text": "B", "is_correct": false},
			},
		}))

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/api/admin/quizzes/"+quizID, adminCookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/admin/quizzes/:id = %d, want 204: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/admin/quizzes/"+quizID, adminCookie, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET deleted quiz = %d, want 404", rec.Code)
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
