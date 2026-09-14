package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// AdminQuizHandler implements dev-plan-2-2-admin-api's Admin CRUD endpoints
// for quizzes (practice mode) and exams (exam mode) — both stored as
// `quizzes` rows distinguished by `mode`. See that plan's terminology note:
// this is unrelated to the Phase 1 lesson-tied ExamHandler/`exams` table.
type AdminQuizHandler struct {
	Quizzes *repository.QuizRepository
}

type quizRequest struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Mode         string `json:"mode"`
	PassingScore *int   `json:"passing_score"`
	Published    bool   `json:"published"`
}

type quizChoiceRequest struct {
	ChoiceText string `json:"choice_text"`
	IsCorrect  bool   `json:"is_correct"`
	SortOrder  int    `json:"sort_order"`
}

type quizQuestionRequest struct {
	QuestionText  string              `json:"question_text"`
	AllowMultiple bool                `json:"allow_multiple"`
	Explanation   string              `json:"explanation"`
	ReferenceURL  string              `json:"reference_url"`
	SortOrder     int                 `json:"sort_order"`
	Choices       []quizChoiceRequest `json:"choices"`
}

// AdminListQuizzes handles GET /api/admin/quizzes — every quiz regardless of
// mode or published state (dev-plan-2-2 2-2.1).
func (h *AdminQuizHandler) AdminListQuizzes(c *gin.Context) {
	quizzes, err := h.Quizzes.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	items := make([]gin.H, 0, len(quizzes))
	for i := range quizzes {
		items = append(items, quizAdminJSON(&quizzes[i], nil, nil))
	}
	c.JSON(http.StatusOK, gin.H{"quizzes": items})
}

// AdminGetQuiz handles GET /api/admin/quizzes/:id, embedding its questions
// (with is_correct) for the edit form (dev-plan-2-2 2-2.4).
func (h *AdminQuizHandler) AdminGetQuiz(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	quiz, err := h.Quizzes.GetByID(c.Request.Context(), id)
	if err != nil {
		respondQuizError(c, err)
		return
	}
	questions, err := h.Quizzes.ListQuestions(c.Request.Context(), quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	choices, err := h.Quizzes.ListChoicesForQuiz(c.Request.Context(), quiz.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, quizAdminJSON(quiz, questions, choices))
}

// AdminCreateQuiz handles POST /api/admin/quizzes.
func (h *AdminQuizHandler) AdminCreateQuiz(c *gin.Context) {
	req, ok := bindQuizRequest(c)
	if !ok {
		return
	}

	quiz, err := h.Quizzes.Create(c.Request.Context(), req.Slug, req.Title, req.Description, req.Mode, req.PassingScore, req.Published)
	if err != nil {
		respondQuizWriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, quizAdminJSON(quiz, nil, nil))
}

// AdminUpdateQuiz handles PUT /api/admin/quizzes/:id.
func (h *AdminQuizHandler) AdminUpdateQuiz(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	req, ok := bindQuizRequest(c)
	if !ok {
		return
	}

	quiz, err := h.Quizzes.Update(c.Request.Context(), id, req.Slug, req.Title, req.Description, req.Mode, req.PassingScore, req.Published)
	if err != nil {
		respondQuizWriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, quizAdminJSON(quiz, nil, nil))
}

// AdminDeleteQuiz handles DELETE /api/admin/quizzes/:id — cascades to its
// questions/choices (and any recorded attempts/answers once Step 2-3 lands).
func (h *AdminQuizHandler) AdminDeleteQuiz(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Quizzes.Delete(c.Request.Context(), id); err != nil {
		respondQuizError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminCreateQuestion handles POST /api/admin/quizzes/:quizId/questions.
func (h *AdminQuizHandler) AdminCreateQuestion(c *gin.Context) {
	quizID, ok := parseIDParam(c, "quizId")
	if !ok {
		return
	}
	if _, err := h.Quizzes.GetByID(c.Request.Context(), quizID); err != nil {
		respondQuizError(c, err)
		return
	}

	req, ok := bindQuizQuestionRequest(c)
	if !ok {
		return
	}

	question, choices, err := h.Quizzes.CreateQuestion(c.Request.Context(), quizID, req.QuestionText, req.AllowMultiple, req.Explanation, req.ReferenceURL, req.SortOrder, toQuizChoiceInputs(req.Choices))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	c.JSON(http.StatusCreated, quizQuestionAdminJSON(question, choices))
}

// AdminUpdateQuestion handles PUT /api/admin/quiz-questions/:id.
func (h *AdminQuizHandler) AdminUpdateQuestion(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	req, ok := bindQuizQuestionRequest(c)
	if !ok {
		return
	}

	question, choices, err := h.Quizzes.UpdateQuestion(c.Request.Context(), id, req.QuestionText, req.AllowMultiple, req.Explanation, req.ReferenceURL, req.SortOrder, toQuizChoiceInputs(req.Choices))
	if err != nil {
		respondQuizError(c, err)
		return
	}
	c.JSON(http.StatusOK, quizQuestionAdminJSON(question, choices))
}

// AdminDeleteQuestion handles DELETE /api/admin/quiz-questions/:id.
func (h *AdminQuizHandler) AdminDeleteQuestion(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Quizzes.DeleteQuestion(c.Request.Context(), id); err != nil {
		respondQuizError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func bindQuizRequest(c *gin.Context) (quizRequest, bool) {
	var req quizRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Slug == "" || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and title are required"})
		return req, false
	}
	if err := service.ValidateQuizMode(req.Mode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	if err := service.ValidatePassingScore(req.Mode, req.PassingScore); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	return req, true
}

func bindQuizQuestionRequest(c *gin.Context) (quizQuestionRequest, bool) {
	var req quizQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.QuestionText == "" || req.Explanation == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question_text and explanation are required"})
		return req, false
	}
	if err := service.ValidateQuestionChoices(req.AllowMultiple, toQuizChoiceInputs(req.Choices)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	return req, true
}

func toQuizChoiceInputs(choices []quizChoiceRequest) []repository.QuizChoiceInput {
	out := make([]repository.QuizChoiceInput, 0, len(choices))
	for _, c := range choices {
		out = append(out, repository.QuizChoiceInput{ChoiceText: c.ChoiceText, IsCorrect: c.IsCorrect, SortOrder: c.SortOrder})
	}
	return out
}

func respondQuizError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func respondQuizWriteError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if errors.Is(service.AsDuplicateSlug(err), service.ErrDuplicateSlug) {
		c.JSON(http.StatusConflict, gin.H{"error": "slug already in use"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func quizAdminJSON(quiz *repository.Quiz, questions []repository.QuizQuestion, choices []repository.QuizChoice) gin.H {
	h := gin.H{
		"id":            strconv.FormatInt(quiz.ID, 10),
		"slug":          quiz.Slug,
		"title":         quiz.Title,
		"description":   quiz.Description,
		"mode":          quiz.Mode,
		"passing_score": quiz.PassingScore,
		"published":     quiz.Published,
	}
	if questions != nil {
		h["questions"] = quizQuestionsAdminJSON(questions, choices)
	}
	return h
}

func quizQuestionsAdminJSON(questions []repository.QuizQuestion, choices []repository.QuizChoice) []gin.H {
	byQuestion := groupQuizChoicesByQuestion(choices)
	items := make([]gin.H, 0, len(questions))
	for _, q := range questions {
		items = append(items, quizQuestionAdminJSON(&q, byQuestion[q.ID]))
	}
	return items
}

func quizQuestionAdminJSON(q *repository.QuizQuestion, choices []repository.QuizChoice) gin.H {
	choiceItems := make([]gin.H, 0, len(choices))
	for _, ch := range choices {
		choiceItems = append(choiceItems, gin.H{
			"id":          strconv.FormatInt(ch.ID, 10),
			"choice_text": ch.ChoiceText,
			"is_correct":  ch.IsCorrect,
			"sort_order":  ch.SortOrder,
		})
	}
	return gin.H{
		"id":             strconv.FormatInt(q.ID, 10),
		"quiz_id":        strconv.FormatInt(q.QuizID, 10),
		"question_text":  q.QuestionText,
		"allow_multiple": q.AllowMultiple,
		"explanation":    q.Explanation,
		"reference_url":  q.ReferenceURL,
		"sort_order":     q.SortOrder,
		"choices":        choiceItems,
	}
}

func groupQuizChoicesByQuestion(choices []repository.QuizChoice) map[int64][]repository.QuizChoice {
	byQuestion := make(map[int64][]repository.QuizChoice)
	for _, ch := range choices {
		byQuestion[ch.QuestionID] = append(byQuestion[ch.QuestionID], ch)
	}
	return byQuestion
}
