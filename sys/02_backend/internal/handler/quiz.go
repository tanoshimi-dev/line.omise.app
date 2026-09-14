package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// QuizHandler implements dev-plan-2-3-answer-scoring-api's public/Reader
// endpoints: fetching a quiz, answering it (practice mode) or submitting it
// (exam mode), and a logged-in Reader's history/progress. See
// dev-plan-2-2-admin-api's terminology note — this is unrelated to the
// Phase 1 lesson-tied ExamHandler.
type QuizHandler struct {
	Quizzes *repository.QuizRepository
}

type quizAnswerRequest struct {
	ChoiceIDs []int64 `json:"choice_ids"`
}

type quizSubmitAnswerRequest struct {
	QuestionID int64   `json:"question_id"`
	ChoiceIDs  []int64 `json:"choice_ids"`
}

type quizSubmitRequest struct {
	StartedAt time.Time                 `json:"started_at"`
	Answers   []quizSubmitAnswerRequest `json:"answers"`
}

// GetQuiz handles GET /api/quizzes/:slug — public, no login required. The
// response never includes is_correct, explanation or reference_url
// (dev-plan-2-3 2-3.1: no cheating).
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	ctx := c.Request.Context()
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	questions, err := h.Quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	choices, err := h.Quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, quizPublicJSON(quiz, questions, choices))
}

// AnswerQuestion handles POST /api/quiz-questions/:id/answer — practice mode
// only. Anyone can answer; the answer is only recorded when the caller is
// logged in (middleware.OptionalUser).
func (h *QuizHandler) AnswerQuestion(c *gin.Context) {
	ctx := c.Request.Context()
	questionID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	question, err := h.Quizzes.GetQuestionByID(ctx, questionID)
	if err != nil {
		respondQuizError(c, err)
		return
	}
	quiz, err := h.Quizzes.GetByID(ctx, question.QuizID)
	if err != nil || !quiz.Published {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if quiz.Mode != "practice" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this question belongs to an exam-mode quiz; submit via POST /api/quizzes/:slug/submit instead"})
		return
	}

	var req quizAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.ChoiceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "choice_ids is required"})
		return
	}

	choices, err := h.Quizzes.ListChoicesForQuestion(ctx, question.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	isCorrect, correctChoiceIDs, err := service.GradeQuizQuestion(question.AllowMultiple, choices, req.ChoiceIDs)
	if err != nil {
		respondQuizGradeError(c, err)
		return
	}

	if user := middleware.CurrentUser(c); user != nil {
		if _, err := h.Quizzes.SaveAnswer(ctx, user.ID, question.ID, nil, req.ChoiceIDs, isCorrect); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"question_id":         strconv.FormatInt(question.ID, 10),
		"is_correct":          isCorrect,
		"selected_choice_ids": formatIDs(req.ChoiceIDs),
		"correct_choice_ids":  formatIDs(correctChoiceIDs),
		"explanation":         question.Explanation,
		"reference_url":       question.ReferenceURL,
	})
}

// SubmitQuiz handles POST /api/quizzes/:slug/submit — exam mode only.
// Anyone can submit; the attempt is only recorded when the caller is
// logged in (middleware.OptionalUser).
func (h *QuizHandler) SubmitQuiz(c *gin.Context) {
	ctx := c.Request.Context()
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	if quiz.Mode != "exam" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this quiz is practice mode; answer via POST /api/quiz-questions/:id/answer instead"})
		return
	}

	var req quizSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	questions, err := h.Quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if len(questions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quiz has no questions"})
		return
	}
	choices, err := h.Quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	choicesByQuestion := groupQuizChoicesByQuestion(choices)
	questionByID := make(map[int64]repository.QuizQuestion, len(questions))
	for _, q := range questions {
		questionByID[q.ID] = q
	}

	answers := make([]service.QuizExamAnswer, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, service.QuizExamAnswer{QuestionID: a.QuestionID, ChoiceIDs: a.ChoiceIDs})
	}

	grade, err := service.GradeQuizExam(questions, choicesByQuestion, answers, quiz.PassingScore)
	if err != nil {
		respondQuizGradeError(c, err)
		return
	}

	startedAt := req.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	var attemptID *int64
	if user := middleware.CurrentUser(c); user != nil {
		attempt, err := h.Quizzes.CreateAttempt(ctx, user.ID, quiz.ID, grade.Score, grade.TotalQuestions, grade.Passed, startedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		attemptID = &attempt.ID
		for _, qg := range grade.Questions {
			if _, err := h.Quizzes.SaveAnswer(ctx, user.ID, qg.QuestionID, attemptID, qg.SelectedChoiceIDs, qg.IsCorrect); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, quizSubmitResultJSON(quiz, grade, questionByID, attemptID))
}

// GetPracticeHistory handles GET /api/me/quizzes/:slug/history — practice
// mode only, Reader login required.
func (h *QuizHandler) GetPracticeHistory(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	if quiz.Mode != "practice" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "history is only available for practice-mode quizzes; use /attempts for exam mode"})
		return
	}

	questions, err := h.Quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	questionByID := make(map[int64]repository.QuizQuestion, len(questions))
	for _, q := range questions {
		questionByID[q.ID] = q
	}

	answers, err := h.Quizzes.ListPracticeHistory(ctx, user.ID, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	items := make([]gin.H, 0, len(answers))
	for _, a := range answers {
		q := questionByID[a.QuestionID]
		items = append(items, gin.H{
			"question_id":         strconv.FormatInt(a.QuestionID, 10),
			"question_text":       q.QuestionText,
			"selected_choice_ids": formatIDs(a.SelectedChoiceIDs),
			"is_correct":          a.IsCorrect,
			"answered_at":         a.AnsweredAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"quiz_id": strconv.FormatInt(quiz.ID, 10), "history": items})
}

// ListAttempts handles GET /api/me/quizzes/:slug/attempts — exam mode only,
// Reader login required.
func (h *QuizHandler) ListAttempts(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	if quiz.Mode != "exam" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attempts are only available for exam-mode quizzes; use /history for practice mode"})
		return
	}

	attempts, err := h.Quizzes.ListAttempts(ctx, user.ID, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	items := make([]gin.H, 0, len(attempts))
	for _, a := range attempts {
		items = append(items, quizAttemptSummaryJSON(&a))
	}
	c.JSON(http.StatusOK, gin.H{"quiz_id": strconv.FormatInt(quiz.ID, 10), "attempts": items})
}

// GetAttempt handles GET /api/me/quizzes/:slug/attempts/:attemptId — the
// per-question review of one past exam-mode submission, Reader login
// required. 404s (rather than 403) both when the attempt doesn't exist and
// when it belongs to someone else, so a guessed id can't confirm another
// user's attempt exists (dev-plan-2-3 2-3.5).
func (h *QuizHandler) GetAttempt(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	attemptID, ok := parseIDParam(c, "attemptId")
	if !ok {
		return
	}
	attempt, err := h.Quizzes.GetOwnAttempt(ctx, attemptID, user.ID)
	if err != nil || attempt.QuizID != quiz.ID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	questions, err := h.Quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	questionByID := make(map[int64]repository.QuizQuestion, len(questions))
	for _, q := range questions {
		questionByID[q.ID] = q
	}
	choices, err := h.Quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	choicesByQuestion := groupQuizChoicesByQuestion(choices)

	answers, err := h.Quizzes.ListAnswersForAttempt(ctx, attempt.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	questionItems := make([]gin.H, 0, len(answers))
	for _, a := range answers {
		q := questionByID[a.QuestionID]
		var correctIDs []int64
		for _, ch := range choicesByQuestion[a.QuestionID] {
			if ch.IsCorrect {
				correctIDs = append(correctIDs, ch.ID)
			}
		}
		questionItems = append(questionItems, gin.H{
			"question_id":         strconv.FormatInt(a.QuestionID, 10),
			"question_text":       q.QuestionText,
			"selected_choice_ids": formatIDs(a.SelectedChoiceIDs),
			"correct_choice_ids":  formatIDs(correctIDs),
			"is_correct":          a.IsCorrect,
			"explanation":         q.Explanation,
			"reference_url":       q.ReferenceURL,
		})
	}

	result := quizAttemptSummaryJSON(attempt)
	result["questions"] = questionItems
	c.JSON(http.StatusOK, result)
}

// GetQuizProgress handles GET /api/me/quizzes/:slug/progress — Reader login
// required. Shape depends on the quiz's mode (dev-plan-2-3 2-3.4).
func (h *QuizHandler) GetQuizProgress(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}

	progress, err := h.quizProgress(ctx, user.ID, quiz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, progress)
}

// GetMyQuizzesProgress handles GET /api/me/quizzes/progress — a summary
// across every published quiz/exam (dev-plan-2-3 2-3.4, マイページ一覧用,
// mirroring ProgressHandler.GetMyProgress which likewise lists every
// published course).
func (h *QuizHandler) GetMyQuizzesProgress(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)

	quizzes, err := h.Quizzes.ListPublished(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	items := make([]gin.H, 0, len(quizzes))
	for i := range quizzes {
		progress, err := h.quizProgress(ctx, user.ID, &quizzes[i])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		items = append(items, progress)
	}
	c.JSON(http.StatusOK, gin.H{"quizzes": items})
}

func (h *QuizHandler) quizProgress(ctx context.Context, userID int64, quiz *repository.Quiz) (gin.H, error) {
	questions, err := h.Quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		return nil, err
	}

	base := gin.H{
		"quiz_id": strconv.FormatInt(quiz.ID, 10),
		"slug":    quiz.Slug,
		"title":   quiz.Title,
		"mode":    quiz.Mode,
	}

	if quiz.Mode == "practice" {
		answers, err := h.Quizzes.ListPracticeHistory(ctx, userID, quiz.ID)
		if err != nil {
			return nil, err
		}
		latestByQuestion := make(map[int64]bool, len(questions))
		for _, a := range answers { // newest first; keep only the first (latest) per question
			if _, seen := latestByQuestion[a.QuestionID]; seen {
				continue
			}
			latestByQuestion[a.QuestionID] = a.IsCorrect
		}
		correct := 0
		for _, ok := range latestByQuestion {
			if ok {
				correct++
			}
		}
		base["total_questions"] = len(questions)
		base["answered_count"] = len(latestByQuestion)
		base["correct_count"] = correct
		return base, nil
	}

	attempts, err := h.Quizzes.ListAttempts(ctx, userID, quiz.ID)
	if err != nil {
		return nil, err
	}
	base["total_questions"] = len(questions)
	base["attempt_count"] = len(attempts)
	base["passing_score"] = quiz.PassingScore
	if len(attempts) > 0 {
		best := attempts[0]
		for _, a := range attempts {
			if a.Score > best.Score {
				best = a
			}
		}
		base["best_score"] = best.Score
		base["latest_score"] = attempts[0].Score
		base["latest_passed"] = attempts[0].Passed
	} else {
		base["best_score"] = nil
		base["latest_score"] = nil
		base["latest_passed"] = nil
	}
	return base, nil
}

func respondQuizGradeError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidQuizAnswer) || errors.Is(err, service.ErrTooManyChoicesSelected) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func formatIDs(ids []int64) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.FormatInt(id, 10))
	}
	return out
}

func quizPublicJSON(quiz *repository.Quiz, questions []repository.QuizQuestion, choices []repository.QuizChoice) gin.H {
	byQuestion := groupQuizChoicesByQuestion(choices)
	questionItems := make([]gin.H, 0, len(questions))
	for _, q := range questions {
		choiceItems := make([]gin.H, 0, len(byQuestion[q.ID]))
		for _, ch := range byQuestion[q.ID] {
			choiceItems = append(choiceItems, gin.H{
				"id":          strconv.FormatInt(ch.ID, 10),
				"choice_text": ch.ChoiceText,
				"sort_order":  ch.SortOrder,
			})
		}
		questionItems = append(questionItems, gin.H{
			"id":             strconv.FormatInt(q.ID, 10),
			"question_text":  q.QuestionText,
			"allow_multiple": q.AllowMultiple,
			"sort_order":     q.SortOrder,
			"choices":        choiceItems,
		})
	}
	return gin.H{
		"id":            strconv.FormatInt(quiz.ID, 10),
		"slug":          quiz.Slug,
		"title":         quiz.Title,
		"description":   quiz.Description,
		"mode":          quiz.Mode,
		"passing_score": quiz.PassingScore,
		"questions":     questionItems,
	}
}

func quizSubmitResultJSON(quiz *repository.Quiz, grade service.QuizExamGrade, questionByID map[int64]repository.QuizQuestion, attemptID *int64) gin.H {
	questionItems := make([]gin.H, 0, len(grade.Questions))
	for _, qg := range grade.Questions {
		q := questionByID[qg.QuestionID]
		questionItems = append(questionItems, gin.H{
			"question_id":         strconv.FormatInt(qg.QuestionID, 10),
			"question_text":       q.QuestionText,
			"selected_choice_ids": formatIDs(qg.SelectedChoiceIDs),
			"correct_choice_ids":  formatIDs(qg.CorrectChoiceIDs),
			"is_correct":          qg.IsCorrect,
			"explanation":         q.Explanation,
			"reference_url":       q.ReferenceURL,
		})
	}
	result := gin.H{
		"quiz_id":         strconv.FormatInt(quiz.ID, 10),
		"score":           grade.Score,
		"total_questions": grade.TotalQuestions,
		"passed":          grade.Passed,
		"questions":       questionItems,
	}
	if attemptID != nil {
		result["attempt_id"] = strconv.FormatInt(*attemptID, 10)
	} else {
		result["attempt_id"] = nil
	}
	return result
}

func quizAttemptSummaryJSON(a *repository.UserQuizAttempt) gin.H {
	return gin.H{
		"id":              strconv.FormatInt(a.ID, 10),
		"score":           a.Score,
		"total_questions": a.TotalQuestions,
		"passed":          a.Passed,
		"started_at":      a.StartedAt,
		"submitted_at":    a.SubmittedAt,
	}
}
