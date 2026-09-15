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
// endpoints: fetching a quiz, answering it one question at a time or
// submitting it all at once, and a logged-in Reader's history/progress. Per
// dev-plan-quiz-mode-selection, the reader picks per-attempt which of the
// two flows to use — any published quiz supports both. This is unrelated to
// the Phase 1 lesson-tied ExamHandler.
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

// ListQuizzes handles GET /api/quizzes — published quizzes/exams only, no
// login required. Added in dev-plan-2-4 (2-4.1) for the /learn/quiz list
// page; not part of dev-plan-2-3's endpoint list, but the same shape and
// visibility rule as ListCourses/ListUsecases. No questions are included
// here — use GET /api/quizzes/:slug for that.
func (h *QuizHandler) ListQuizzes(c *gin.Context) {
	quizzes, err := h.Quizzes.ListPublished(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	items := make([]gin.H, 0, len(quizzes))
	for i := range quizzes {
		items = append(items, quizListItemJSON(&quizzes[i]))
	}
	c.JSON(http.StatusOK, gin.H{"quizzes": items})
}

func quizListItemJSON(quiz *repository.Quiz) gin.H {
	return gin.H{
		"id":            strconv.FormatInt(quiz.ID, 10),
		"slug":          quiz.Slug,
		"title":         quiz.Title,
		"description":   quiz.Description,
		"passing_score": quiz.PassingScore,
	}
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

// AnswerQuestion handles POST /api/quiz-questions/:id/answer — grades one
// question immediately, for a reader who chose to answer question-by-
// question. Anyone can answer; the answer is only recorded when the caller
// is logged in (middleware.OptionalUser).
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

// SubmitQuiz handles POST /api/quizzes/:slug/submit — grades every question
// at once and returns the final score, for a reader who chose to submit the
// whole quiz together. Anyone can submit; the attempt is only recorded when
// the caller is logged in (middleware.OptionalUser).
func (h *QuizHandler) SubmitQuiz(c *gin.Context) {
	ctx := c.Request.Context()
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
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

// GetPracticeHistory handles GET /api/me/quizzes/:slug/history — a reader's
// question-by-question answer history for a quiz (empty if they've only
// ever submitted it as a whole), Reader login required.
func (h *QuizHandler) GetPracticeHistory(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
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
			"id":                  strconv.FormatInt(a.ID, 10),
			"question_id":         strconv.FormatInt(a.QuestionID, 10),
			"question_text":       q.QuestionText,
			"selected_choice_ids": formatIDs(a.SelectedChoiceIDs),
			"is_correct":          a.IsCorrect,
			"answered_at":         a.AnsweredAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"quiz_id": strconv.FormatInt(quiz.ID, 10), "history": items})
}

// ListAttempts handles GET /api/me/quizzes/:slug/attempts — a reader's
// whole-quiz submission history for a quiz (empty if they've only ever
// answered it question-by-question), Reader login required.
func (h *QuizHandler) ListAttempts(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
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

// DeleteAttempt handles DELETE /api/me/quizzes/:slug/attempts/:attemptId —
// deletes one of the caller's own exam-mode attempts (dev-plan-quiz-history-delete).
// 404s both when the attempt doesn't exist and when it belongs to someone
// else, matching GetAttempt's ownership check.
func (h *QuizHandler) DeleteAttempt(c *gin.Context) {
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
	if err := h.Quizzes.DeleteOwnAttempt(ctx, attemptID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteAllAttempts handles DELETE /api/me/quizzes/:slug/attempts — deletes
// every exam-mode attempt the caller has for this quiz (dev-plan-quiz-history-delete's
// bulk "clear history" action).
func (h *QuizHandler) DeleteAllAttempts(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	if err := h.Quizzes.DeleteAttemptsForQuiz(ctx, user.ID, quiz.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeletePracticeAnswer handles DELETE /api/me/quizzes/:slug/history/:answerId
// — deletes one of the caller's own question-by-question answers
// (dev-plan-quiz-history-delete).
func (h *QuizHandler) DeletePracticeAnswer(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	if _, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug")); err != nil {
		respondQuizError(c, err)
		return
	}
	answerID, ok := parseIDParam(c, "answerId")
	if !ok {
		return
	}
	if err := h.Quizzes.DeleteOwnPracticeAnswer(ctx, answerID, user.ID); err != nil {
		respondQuizError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteAllPracticeHistory handles DELETE /api/me/quizzes/:slug/history —
// deletes every question-by-question answer the caller has for this quiz
// (dev-plan-quiz-history-delete's bulk "clear history" action).
func (h *QuizHandler) DeleteAllPracticeHistory(c *gin.Context) {
	ctx := c.Request.Context()
	user := middleware.CurrentUser(c)
	quiz, err := h.Quizzes.GetPublishedBySlug(ctx, c.Param("slug"))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	if err := h.Quizzes.DeletePracticeHistoryForQuiz(ctx, user.ID, quiz.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetQuizProgress handles GET /api/me/quizzes/:slug/progress — Reader login
// required. Returns both question-by-question and whole-quiz-submission
// progress for the quiz, since a reader may have used either or both
// (dev-plan-quiz-mode-selection).
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
		"quiz_id":         strconv.FormatInt(quiz.ID, 10),
		"slug":            quiz.Slug,
		"title":           quiz.Title,
		"total_questions": len(questions),
		"passing_score":   quiz.PassingScore,
	}

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
	base["answered_count"] = len(latestByQuestion)
	base["correct_count"] = correct

	attempts, err := h.Quizzes.ListAttempts(ctx, userID, quiz.ID)
	if err != nil {
		return nil, err
	}
	base["attempt_count"] = len(attempts)
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
