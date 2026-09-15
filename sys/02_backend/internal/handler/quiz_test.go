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

// createPublishedQuiz creates a published quiz with one single-choice
// question (choice "A" correct) via the Admin API, returning the quiz
// slug/id (as returned by the API, a string) and the question id as an
// int64 — request bodies for answer/submit use plain JSON numbers for
// question_id/choice_ids, matching the existing exam submit convention
// (LessonInteractive.tsx converts the string id back to a Number before
// sending), even though every *response* id is a string. passingScore is
// optional (nil = no pass/fail threshold) and independent of how the reader
// later chooses to answer (dev-plan-quiz-mode-selection: any quiz supports
// both question-by-question answering and whole-quiz submission).
func createPublishedQuiz(t *testing.T, router *gin.Engine, adminCookie, slug string, passingScore *int) (quizID string, questionID int64) {
	t.Helper()

	body := map[string]any{"slug": slug, "title": "Quiz", "published": true}
	if passingScore != nil {
		body["passing_score"] = *passingScore
	}
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie, mustJSON(t, body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create quiz = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	var quiz map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &quiz); err != nil {
		t.Fatalf("unmarshal quiz: %v", err)
	}
	quizID = quiz["id"].(string)

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes/"+quizID+"/questions", adminCookie,
		mustJSON(t, map[string]any{
			"question_text": "Q1?",
			"explanation":   "Because A is right.",
			"reference_url": "https://example.com",
			"choices": []map[string]any{
				{"choice_text": "A", "is_correct": true},
				{"choice_text": "B", "is_correct": false},
			},
		}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create question = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	var question map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &question); err != nil {
		t.Fatalf("unmarshal question: %v", err)
	}
	return quizID, mustParseInt64(t, question["id"].(string))
}

// choiceIDByText looks up a question's choice id by its text, via the Admin
// API's is_correct-carrying detail view, and returns it as an int64 (see
// createPublishedQuiz's doc comment on why answer/submit bodies use numbers).
func choiceIDByText(t *testing.T, router *gin.Engine, adminCookie, quizID, text string) int64 {
	t.Helper()
	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/admin/quizzes/"+quizID, adminCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get quiz = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Questions []struct {
			Choices []struct {
				ID         string `json:"id"`
				ChoiceText string `json:"choice_text"`
			} `json:"choices"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, ch := range body.Questions[0].Choices {
		if ch.ChoiceText == text {
			return mustParseInt64(t, ch.ID)
		}
	}
	t.Fatalf("no choice with text %q found", text)
	return 0
}

func intPtr(v int) *int { return &v }

func mustParseInt64(t *testing.T, s string) int64 {
	t.Helper()
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.Fatalf("parse id %q: %v", s, err)
	}
	return n
}

func TestListQuizzes_ExcludesDraft(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	createPublishedQuiz(t, router, adminCookie, "listed-quiz", nil)
	testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "draft-listed-quiz", "title": "Draft", "published": false}))

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/quizzes", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/quizzes = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Quizzes []map[string]any `json:"quizzes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Quizzes) != 1 || body.Quizzes[0]["slug"] != "listed-quiz" {
		t.Errorf("public quiz list = %+v, want only the published quiz", body.Quizzes)
	}
}

func TestGetQuiz_PublicResponseHidesAnswerAndExplanation(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	createPublishedQuiz(t, router, adminCookie, "public-quiz", nil)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/quizzes/public-quiz", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/quizzes/:slug = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if body == "" || strings.Contains(body, "is_correct") || strings.Contains(body, "explanation") || strings.Contains(body, "reference_url") {
		t.Errorf("public quiz response leaks answer data: %s", body)
	}
}

func TestGetQuiz_UnpublishedIs404(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	testutil.DoRequest(t, router, http.MethodPost, "/api/admin/quizzes", adminCookie,
		mustJSON(t, map[string]any{"slug": "draft-quiz", "title": "Draft", "published": false}))

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/quizzes/draft-quiz", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET unpublished quiz = %d, want 404", rec.Code)
	}
}

func TestAnswerQuestion_AnonymousGetsResultButNothingSaved(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "anon-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", "",
		mustJSON(t, map[string]any{"choice_ids": []int64{choiceA}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST answer = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["is_correct"] != true {
		t.Errorf("is_correct = %v, want true", body["is_correct"])
	}
	if body["explanation"] != "Because A is right." {
		t.Errorf("explanation = %v, want revealed after answering", body["explanation"])
	}

	var count int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM user_quiz_answers").Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("user_quiz_answers row count = %d, want 0 (anonymous answers must not be saved)", count)
	}
}

func TestAnswerQuestion_LoggedInSavesAnswerVisibleInHistory(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "history-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", readerCookie,
		mustJSON(t, map[string]any{"choice_ids": []int64{choiceA}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST answer = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/history-quiz/history", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET history = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		History []struct {
			IsCorrect bool `json:"is_correct"`
		} `json:"history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.History) != 1 || !body.History[0].IsCorrect {
		t.Fatalf("history = %+v, want one correct entry", body.History)
	}
}

// TestAnswerQuestion_WorksEvenWithPassingScoreSet confirms a quiz that has a
// passing_score (so it can also be taken as a whole-quiz submission) can
// still be answered question-by-question — the reader chooses per attempt,
// not the quiz record (dev-plan-quiz-mode-selection).
func TestAnswerQuestion_WorksEvenWithPassingScoreSet(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "exam-answer-quiz", intPtr(60))
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", "",
		mustJSON(t, map[string]any{"choice_ids": []int64{choiceA}}))
	if rec.Code != http.StatusOK {
		t.Errorf("answering a question on a quiz with passing_score set = %d, want 200: %s", rec.Code, rec.Body.String())
	}
}

// TestSubmitQuiz_WorksEvenWithoutPassingScore confirms a quiz with no
// passing_score can still be submitted as a whole (passed comes back nil,
// since there's no threshold to grade against) — the reader chooses per
// attempt (dev-plan-quiz-mode-selection).
func TestSubmitQuiz_WorksEvenWithoutPassingScore(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "no-passing-score-submit-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/no-passing-score-submit-quiz/submit", "",
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("submitting a quiz without passing_score = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["passed"] != nil {
		t.Errorf("passed = %v, want nil when the quiz has no passing_score", body["passed"])
	}
}

func TestSubmitQuiz_AnonymousGetsScoreButNothingSaved(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "anon-exam", intPtr(60))
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/anon-exam/submit", "",
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST submit = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["score"] != float64(1) || body["total_questions"] != float64(1) {
		t.Errorf("score/total = %v/%v, want 1/1", body["score"], body["total_questions"])
	}
	if body["passed"] != true {
		t.Errorf("passed = %v, want true (100%% >= 60%% passing_score)", body["passed"])
	}
	if body["attempt_id"] != nil {
		t.Errorf("attempt_id = %v, want nil for an anonymous submission", body["attempt_id"])
	}

	var count int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM user_quiz_attempts").Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("user_quiz_attempts row count = %d, want 0 (anonymous submissions must not be saved)", count)
	}
}

func TestSubmitQuiz_LoggedInSavesAttemptVisibleInAttemptsAndDetail(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "logged-in-exam", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/logged-in-exam/submit", readerCookie,
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST submit = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var submitBody map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &submitBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	attemptID := submitBody["attempt_id"].(string)

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/logged-in-exam/attempts", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET attempts = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var attemptsBody struct {
		Attempts []map[string]any `json:"attempts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &attemptsBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(attemptsBody.Attempts) != 1 {
		t.Fatalf("attempts = %+v, want 1 entry", attemptsBody.Attempts)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/logged-in-exam/attempts/"+attemptID, readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET attempt detail = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var detail struct {
		Questions []struct {
			IsCorrect   bool   `json:"is_correct"`
			Explanation string `json:"explanation"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(detail.Questions) != 1 || !detail.Questions[0].IsCorrect || detail.Questions[0].Explanation == "" {
		t.Errorf("attempt detail = %+v, want one correct question with explanation", detail.Questions)
	}
}

func TestGetAttempt_AnotherUsersAttemptIs404(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "isolation-exam", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	owner := testutil.CreateUser(t, pool, "reader")
	ownerCookie := testutil.LoginCookieValue(t, pool, owner.ID)
	other := testutil.CreateUser(t, pool, "reader")
	otherCookie := testutil.LoginCookieValue(t, pool, other.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/isolation-exam/submit", ownerCookie,
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST submit = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var submitBody map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &submitBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	attemptID := submitBody["attempt_id"].(string)

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/isolation-exam/attempts/"+attemptID, otherCookie, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("another user's attempt lookup = %d, want 404", rec.Code)
	}
}

func TestGetQuizProgress_PracticeMode(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "progress-practice-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")
	choiceB := choiceIDByText(t, router, adminCookie, quizID, "B")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	// Answer wrong first, then correctly — progress should reflect the
	// latest answer per question, not every historical attempt.
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", readerCookie,
		mustJSON(t, map[string]any{"choice_ids": []int64{choiceB}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST first (wrong) answer = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", readerCookie,
		mustJSON(t, map[string]any{"choice_ids": []int64{choiceA}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST answer = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/progress-practice-quiz/progress", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET progress = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["total_questions"] != float64(1) || body["answered_count"] != float64(1) || body["correct_count"] != float64(1) {
		t.Errorf("progress = %+v, want total=1 answered=1 correct=1 (latest answer was correct)", body)
	}
}

func TestGetQuizProgress_ExamMode(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "progress-exam-quiz", intPtr(60))
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/progress-exam-quiz/submit", readerCookie,
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST submit = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/progress-exam-quiz/progress", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET progress = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["attempt_count"] != float64(1) || body["best_score"] != float64(1) || body["latest_score"] != float64(1) {
		t.Errorf("progress = %+v, want attempt_count=1 best_score=1 latest_score=1", body)
	}
	if body["latest_passed"] != true {
		t.Errorf("latest_passed = %v, want true", body["latest_passed"])
	}
}

func TestGetMyQuizzesProgress_ListsPublishedQuizzes(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	createPublishedQuiz(t, router, adminCookie, "summary-practice-quiz", nil)
	createPublishedQuiz(t, router, adminCookie, "summary-exam-quiz", intPtr(60))

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/progress", readerCookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/me/quizzes/progress = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Quizzes []map[string]any `json:"quizzes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Quizzes) != 2 {
		t.Fatalf("quizzes = %+v, want 2 entries (never attempted, but every published quiz is listed)", body.Quizzes)
	}
}

func TestQuizHistoryAndAttempts_RequireLogin(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	createPublishedQuiz(t, router, adminCookie, "login-required-quiz", nil)

	rec := testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/login-required-quiz/history", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated GET history = %d, want 401", rec.Code)
	}
}

// practiceHistoryEntryID answers questionID as choiceIDs via the given
// cookie, then fetches history and returns the id of the newest entry for
// that question (dev-plan-quiz-history-delete tests).
func practiceHistoryEntryID(t *testing.T, router *gin.Engine, cookie, slug string, questionID int64, choiceIDs []int64) string {
	t.Helper()
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quiz-questions/"+strconv.FormatInt(questionID, 10)+"/answer", cookie,
		mustJSON(t, map[string]any{"choice_ids": choiceIDs}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST answer = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/"+slug+"/history", cookie, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET history = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		History []struct {
			ID string `json:"id"`
		} `json:"history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.History) == 0 {
		t.Fatalf("history is empty after answering")
	}
	return body.History[0].ID
}

func TestDeletePracticeAnswer_OwnerCanDeleteOneEntry(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-history-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)
	answerID := practiceHistoryEntryID(t, router, readerCookie, "delete-history-quiz", questionID, []int64{choiceA})

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-history-quiz/history/"+answerID, readerCookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE history entry = %d, want 204: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-history-quiz/history", readerCookie, nil)
	var body struct {
		History []map[string]any `json:"history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.History) != 0 {
		t.Errorf("history after delete = %+v, want empty", body.History)
	}
}

func TestDeletePracticeAnswer_OtherUsersEntryIs404(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-history-isolation-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	owner := testutil.CreateUser(t, pool, "reader")
	ownerCookie := testutil.LoginCookieValue(t, pool, owner.ID)
	other := testutil.CreateUser(t, pool, "reader")
	otherCookie := testutil.LoginCookieValue(t, pool, other.ID)
	answerID := practiceHistoryEntryID(t, router, ownerCookie, "delete-history-isolation-quiz", questionID, []int64{choiceA})

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-history-isolation-quiz/history/"+answerID, otherCookie, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("another user's history entry delete = %d, want 404", rec.Code)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-history-isolation-quiz/history", ownerCookie, nil)
	var body struct {
		History []map[string]any `json:"history"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.History) != 1 {
		t.Errorf("owner's history after another user's failed delete = %+v, want 1 entry untouched", body.History)
	}
}

func TestDeleteAllPracticeHistory_ClearsOwnHistoryOnly(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-all-history-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	readerA := testutil.CreateUser(t, pool, "reader")
	readerACookie := testutil.LoginCookieValue(t, pool, readerA.ID)
	readerB := testutil.CreateUser(t, pool, "reader")
	readerBCookie := testutil.LoginCookieValue(t, pool, readerB.ID)
	practiceHistoryEntryID(t, router, readerACookie, "delete-all-history-quiz", questionID, []int64{choiceA})
	practiceHistoryEntryID(t, router, readerBCookie, "delete-all-history-quiz", questionID, []int64{choiceA})

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-all-history-quiz/history", readerACookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE all history = %d, want 204: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-all-history-quiz/history", readerACookie, nil)
	var aBody struct {
		History []map[string]any `json:"history"`
	}
	json.Unmarshal(rec.Body.Bytes(), &aBody)
	if len(aBody.History) != 0 {
		t.Errorf("reader A history after bulk delete = %+v, want empty", aBody.History)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-all-history-quiz/history", readerBCookie, nil)
	var bBody struct {
		History []map[string]any `json:"history"`
	}
	json.Unmarshal(rec.Body.Bytes(), &bBody)
	if len(bBody.History) != 1 {
		t.Errorf("reader B history after reader A's bulk delete = %+v, want 1 entry untouched", bBody.History)
	}
}

func TestDeleteAttempt_OwnerCanDeleteOneAttempt(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-attempt-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	reader := testutil.CreateUser(t, pool, "reader")
	readerCookie := testutil.LoginCookieValue(t, pool, reader.ID)
	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/delete-attempt-quiz/submit", readerCookie,
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	if rec.Code != http.StatusOK {
		t.Fatalf("POST submit = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var submitBody map[string]any
	json.Unmarshal(rec.Body.Bytes(), &submitBody)
	attemptID := submitBody["attempt_id"].(string)

	rec = testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-attempt-quiz/attempts/"+attemptID, readerCookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE attempt = %d, want 204: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-attempt-quiz/attempts", readerCookie, nil)
	var attemptsBody struct {
		Attempts []map[string]any `json:"attempts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &attemptsBody)
	if len(attemptsBody.Attempts) != 0 {
		t.Errorf("attempts after delete = %+v, want empty", attemptsBody.Attempts)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-attempt-quiz/attempts/"+attemptID, readerCookie, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET deleted attempt = %d, want 404 (cascade-deleted answers too)", rec.Code)
	}
}

func TestDeleteAttempt_OtherUsersAttemptIs404(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-attempt-isolation-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	owner := testutil.CreateUser(t, pool, "reader")
	ownerCookie := testutil.LoginCookieValue(t, pool, owner.ID)
	other := testutil.CreateUser(t, pool, "reader")
	otherCookie := testutil.LoginCookieValue(t, pool, other.ID)

	rec := testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/delete-attempt-isolation-quiz/submit", ownerCookie,
		mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	var submitBody map[string]any
	json.Unmarshal(rec.Body.Bytes(), &submitBody)
	attemptID := submitBody["attempt_id"].(string)

	rec = testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-attempt-isolation-quiz/attempts/"+attemptID, otherCookie, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("another user's attempt delete = %d, want 404", rec.Code)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-attempt-isolation-quiz/attempts", ownerCookie, nil)
	var attemptsBody struct {
		Attempts []map[string]any `json:"attempts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &attemptsBody)
	if len(attemptsBody.Attempts) != 1 {
		t.Errorf("owner's attempts after another user's failed delete = %+v, want 1 entry untouched", attemptsBody.Attempts)
	}
}

func TestDeleteAllAttempts_ClearsOwnAttemptsOnly(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	router := testutil.NewRouter(t, pool)
	admin := testutil.CreateUser(t, pool, "admin")
	adminCookie := testutil.LoginCookieValue(t, pool, admin.ID)
	quizID, questionID := createPublishedQuiz(t, router, adminCookie, "delete-all-attempts-quiz", nil)
	choiceA := choiceIDByText(t, router, adminCookie, quizID, "A")

	readerA := testutil.CreateUser(t, pool, "reader")
	readerACookie := testutil.LoginCookieValue(t, pool, readerA.ID)
	readerB := testutil.CreateUser(t, pool, "reader")
	readerBCookie := testutil.LoginCookieValue(t, pool, readerB.ID)
	for _, cookie := range []string{readerACookie, readerBCookie} {
		testutil.DoRequest(t, router, http.MethodPost, "/api/quizzes/delete-all-attempts-quiz/submit", cookie,
			mustJSON(t, map[string]any{"answers": []map[string]any{{"question_id": questionID, "choice_ids": []int64{choiceA}}}}))
	}

	rec := testutil.DoRequest(t, router, http.MethodDelete, "/api/me/quizzes/delete-all-attempts-quiz/attempts", readerACookie, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE all attempts = %d, want 204: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-all-attempts-quiz/attempts", readerACookie, nil)
	var aBody struct {
		Attempts []map[string]any `json:"attempts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &aBody)
	if len(aBody.Attempts) != 0 {
		t.Errorf("reader A attempts after bulk delete = %+v, want empty", aBody.Attempts)
	}

	rec = testutil.DoRequest(t, router, http.MethodGet, "/api/me/quizzes/delete-all-attempts-quiz/attempts", readerBCookie, nil)
	var bBody struct {
		Attempts []map[string]any `json:"attempts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &bBody)
	if len(bBody.Attempts) != 1 {
		t.Errorf("reader B attempts after reader A's bulk delete = %+v, want 1 entry untouched", bBody.Attempts)
	}
}
