package handler_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

// createPublishedLessonWithExam creates a published course, a published
// lesson, an exam, and one two-choice question via the real admin
// endpoints, returning the lesson id as a string (matching how ids are
// serialized in JSON responses throughout the API).
func createPublishedLessonWithExam(t *testing.T, router *gin.Engine, adminCookie string) string {
	t.Helper()

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses", adminCookie,
		mustJSON(t, map[string]any{"slug": "course", "title": "Course", "status": "published"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create course = %d: %s", rec.Code, rec.Body.String())
	}
	courseID := decodeField(t, rec.Body.Bytes(), "id")

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/courses/"+courseID+"/lessons", adminCookie,
		mustJSON(t, map[string]any{"slug": "lesson", "title": "Lesson", "status": "published"}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create lesson = %d: %s", rec.Code, rec.Body.String())
	}
	lessonID := decodeField(t, rec.Body.Bytes(), "id")

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/lessons/"+lessonID+"/exam", adminCookie,
		mustJSON(t, map[string]any{"title": "Quiz", "passing_score": 70}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create exam = %d: %s", rec.Code, rec.Body.String())
	}
	examID := decodeField(t, rec.Body.Bytes(), "id")

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/exams/"+examID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "2+2?",
			"sort_order":    1,
			"choices": []map[string]any{
				{"choice_text": "4", "is_correct": true, "sort_order": 1},
				{"choice_text": "5", "is_correct": false, "sort_order": 2},
			},
		}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create question = %d: %s", rec.Code, rec.Body.String())
	}

	return lessonID
}

func decodeField(t *testing.T, body []byte, field string) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	v, ok := m[field].(string)
	if !ok {
		t.Fatalf("field %q not found or not a string in %s", field, body)
	}
	return v
}

func TestGetExam_NeverIncludesIsCorrect(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	lessonID := createPublishedLessonWithExam(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/lessons/"+lessonID+"/exam", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("Reader GET exam = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "is_correct") {
		t.Errorf("Reader-facing exam response leaked is_correct: %s", rec.Body.String())
	}
}

func TestGetExam_UnauthenticatedGets401(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	lessonID := createPublishedLessonWithExam(t, router, adminCookie)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/lessons/"+lessonID+"/exam", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated GET exam = %d, want 401", rec.Code)
	}
}

func TestSubmitExam_CorrectAnswerScores100AndPasses(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	lessonID := createPublishedLessonWithExam(t, router, adminCookie)

	// Fetch the admin view to learn the real question/choice ids.
	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/admin/lessons/"+lessonID+"/exam", adminCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin GET exam = %d: %s", rec.Code, rec.Body.String())
	}
	var adminExam struct {
		Questions []struct {
			ID      string `json:"id"`
			Choices []struct {
				ID        string `json:"id"`
				IsCorrect bool   `json:"is_correct"`
			} `json:"choices"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &adminExam); err != nil {
		t.Fatalf("unmarshal admin exam: %v", err)
	}
	if len(adminExam.Questions) == 0 {
		t.Fatal("fixture setup problem: no questions found")
	}
	questionID := mustParseInt(t, adminExam.Questions[0].ID)
	var correctChoiceID int64
	for _, c := range adminExam.Questions[0].Choices {
		if c.IsCorrect {
			correctChoiceID = mustParseInt(t, c.ID)
		}
	}
	if correctChoiceID == 0 {
		t.Fatal("fixture setup problem: no correct choice found")
	}

	submitBody := mustJSON(t, map[string]any{
		"answers": []map[string]any{{"question_id": questionID, "choice_id": correctChoiceID}},
	})
	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/lessons/"+lessonID+"/exam/submit", readerCookie, submitBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit exam = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Score  int  `json:"score"`
		Passed bool `json:"passed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	if result.Score != 100 || !result.Passed {
		t.Errorf("result = %+v, want score=100 passed=true", result)
	}
}

func TestSubmitExam_InvalidChoiceIsRejected(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)

	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	lessonID := createPublishedLessonWithExam(t, router, adminCookie)

	submitBody := mustJSON(t, map[string]any{
		"answers": []map[string]any{{"question_id": 999999, "choice_id": 999999}},
	})
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/lessons/"+lessonID+"/exam/submit", readerCookie, submitBody)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("submit with bogus ids = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func mustParseInt(t *testing.T, s string) int64 {
	t.Helper()
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.Fatalf("failed to parse %q as int64: %v", s, err)
	}
	return n
}
