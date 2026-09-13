package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/service"
)

// ExamHandler implements dev-plan-06-exam-progress-api's exam endpoints.
type ExamHandler struct {
	Exams    *repository.ExamRepository
	Courses  *repository.CourseRepository
	Progress *repository.ProgressRepository
}

type examRequest struct {
	Title        string `json:"title"`
	PassingScore int    `json:"passing_score"`
}

type choiceRequest struct {
	ChoiceText string `json:"choice_text"`
	IsCorrect  bool   `json:"is_correct"`
	SortOrder  int    `json:"sort_order"`
}

type questionRequest struct {
	QuestionText string          `json:"question_text"`
	SortOrder    int             `json:"sort_order"`
	Choices      []choiceRequest `json:"choices"`
}

type submitAnswer struct {
	QuestionID int64 `json:"question_id"`
	ChoiceID   int64 `json:"choice_id"`
}

type submitRequest struct {
	Answers []submitAnswer `json:"answers"`
}

// AdminGetExamByLesson handles GET /api/admin/lessons/:id/exam — the admin
// view (includes is_correct, unlike the Reader-facing GET
// /api/lessons/:lessonId/exam), for the exam management UI
// (dev-plan-11-frontend-admin 11.5). 404 if the lesson has no exam yet.
// Uses :id (not :lessonId) to match the existing /admin/lessons/:id routes —
// Gin's router rejects two different wildcard names at the same path position.
func (h *ExamHandler) AdminGetExamByLesson(c *gin.Context) {
	lessonID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	exam, err := h.Exams.GetByLessonID(c.Request.Context(), lessonID)
	if err != nil {
		respondExamError(c, err)
		return
	}
	questions, err := h.Exams.ListQuestions(c.Request.Context(), exam.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	choices, err := h.Exams.ListChoicesForExam(c.Request.Context(), exam.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, examAdminJSON(exam, questions, choices))
}

// AdminCreateExam handles POST /api/admin/lessons/:lessonId/exam.
func (h *ExamHandler) AdminCreateExam(c *gin.Context) {
	lessonID, ok := parseIDParam(c, "lessonId")
	if !ok {
		return
	}
	if _, err := h.Courses.GetLessonByID(c.Request.Context(), lessonID); err != nil {
		respondExamError(c, err)
		return
	}

	var req examRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	exam, err := h.Exams.Create(c.Request.Context(), lessonID, req.Title, req.PassingScore)
	if err != nil {
		if errors.Is(service.AsExamAlreadyExists(err), service.ErrExamAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "lesson already has an exam"})
			return
		}
		respondExamError(c, err)
		return
	}
	c.JSON(http.StatusCreated, examAdminJSON(exam, nil, nil))
}

// AdminCreateQuestion handles POST /api/admin/exams/:examId/questions.
func (h *ExamHandler) AdminCreateQuestion(c *gin.Context) {
	examID, ok := parseIDParam(c, "examId")
	if !ok {
		return
	}
	if _, err := h.Exams.GetByID(c.Request.Context(), examID); err != nil {
		respondExamError(c, err)
		return
	}

	req, ok := bindQuestionRequest(c)
	if !ok {
		return
	}

	question, choices, err := h.Exams.CreateQuestion(c.Request.Context(), examID, req.QuestionText, req.SortOrder, toChoiceInputs(req.Choices))
	if err != nil {
		respondExamError(c, err)
		return
	}
	c.JSON(http.StatusCreated, questionAdminJSON(question, choices))
}

// AdminUpdateQuestion handles PUT /api/admin/questions/:id.
func (h *ExamHandler) AdminUpdateQuestion(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	req, ok := bindQuestionRequest(c)
	if !ok {
		return
	}

	question, choices, err := h.Exams.UpdateQuestion(c.Request.Context(), id, req.QuestionText, req.SortOrder, toChoiceInputs(req.Choices))
	if err != nil {
		respondExamError(c, err)
		return
	}
	c.JSON(http.StatusOK, questionAdminJSON(question, choices))
}

// AdminDeleteQuestion handles DELETE /api/admin/questions/:id.
func (h *ExamHandler) AdminDeleteQuestion(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Exams.DeleteQuestion(c.Request.Context(), id); err != nil {
		respondExamError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetExam handles GET /api/lessons/:lessonId/exam (Reader). The response
// never includes is_correct (dev-plan-06 6.2/完了条件: no cheating).
func (h *ExamHandler) GetExam(c *gin.Context) {
	exam, questions, choices, err := h.loadPublishedExam(c)
	if err != nil {
		respondExamError(c, err)
		return
	}
	c.JSON(http.StatusOK, examPublicJSON(exam, questions, choices))
}

// SubmitExam handles POST /api/lessons/:lessonId/exam/submit (Reader).
func (h *ExamHandler) SubmitExam(c *gin.Context) {
	user := middleware.CurrentUser(c)

	exam, questions, choices, err := h.loadPublishedExam(c)
	if err != nil {
		respondExamError(c, err)
		return
	}

	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	correctByQuestion := make(map[int64]int64, len(questions))
	validChoices := make(map[int64]map[int64]bool, len(questions))
	for _, q := range questions {
		validChoices[q.ID] = map[int64]bool{}
	}
	for _, ch := range choices {
		validChoices[ch.QuestionID][ch.ID] = true
		if ch.IsCorrect {
			correctByQuestion[ch.QuestionID] = ch.ID
		}
	}

	answers := make([]service.ExamAnswer, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, service.ExamAnswer{QuestionID: a.QuestionID, ChoiceID: a.ChoiceID})
	}

	score, passed, err := service.GradeExam(answers, correctByQuestion, validChoices, exam.PassingScore)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.Progress.SaveExamResult(c.Request.Context(), user.ID, exam.ID, score, passed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exam_id":      strconv.FormatInt(exam.ID, 10),
		"score":        result.Score,
		"passed":       result.Passed,
		"submitted_at": result.SubmittedAt,
	})
}

// loadPublishedExam resolves :lessonId to its exam, requiring both the
// lesson and its course to be published — a Reader shouldn't be able to
// probe draft content's exam by guessing lesson ids.
func (h *ExamHandler) loadPublishedExam(c *gin.Context) (*repository.Exam, []repository.ExamQuestion, []repository.ExamChoice, error) {
	lessonID, ok := parseIDParam(c, "lessonId")
	if !ok {
		return nil, nil, nil, errAlreadyResponded
	}

	lesson, err := h.Courses.GetLessonByID(c.Request.Context(), lessonID)
	if err != nil {
		return nil, nil, nil, err
	}
	if lesson.Status != "published" {
		return nil, nil, nil, repository.ErrNotFound
	}
	course, err := h.Courses.GetByID(c.Request.Context(), lesson.CourseID)
	if err != nil || course.Status != "published" {
		return nil, nil, nil, repository.ErrNotFound
	}

	exam, err := h.Exams.GetByLessonID(c.Request.Context(), lessonID)
	if err != nil {
		return nil, nil, nil, err
	}
	questions, err := h.Exams.ListQuestions(c.Request.Context(), exam.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	choices, err := h.Exams.ListChoicesForExam(c.Request.Context(), exam.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return exam, questions, choices, nil
}

// errAlreadyResponded signals that parseIDParam already wrote the response
// (400); callers just need to stop without writing again.
var errAlreadyResponded = errors.New("response already written")

func bindQuestionRequest(c *gin.Context) (questionRequest, bool) {
	var req questionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.QuestionText == "" || len(req.Choices) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question_text and at least one choice are required"})
		return req, false
	}
	return req, true
}

func toChoiceInputs(choices []choiceRequest) []repository.ChoiceInput {
	out := make([]repository.ChoiceInput, 0, len(choices))
	for _, c := range choices {
		out = append(out, repository.ChoiceInput{ChoiceText: c.ChoiceText, IsCorrect: c.IsCorrect, SortOrder: c.SortOrder})
	}
	return out
}

func respondExamError(c *gin.Context, err error) {
	if errors.Is(err, errAlreadyResponded) {
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}

func examAdminJSON(exam *repository.Exam, questions []repository.ExamQuestion, choices []repository.ExamChoice) gin.H {
	h := gin.H{
		"id":            strconv.FormatInt(exam.ID, 10),
		"lesson_id":     strconv.FormatInt(exam.LessonID, 10),
		"title":         exam.Title,
		"passing_score": exam.PassingScore,
	}
	if questions != nil {
		h["questions"] = questionsAdminJSON(questions, choices)
	}
	return h
}

func questionsAdminJSON(questions []repository.ExamQuestion, choices []repository.ExamChoice) []gin.H {
	byQuestion := groupChoicesByQuestion(choices)
	items := make([]gin.H, 0, len(questions))
	for _, q := range questions {
		items = append(items, questionAdminJSON(&q, byQuestion[q.ID]))
	}
	return items
}

func questionAdminJSON(q *repository.ExamQuestion, choices []repository.ExamChoice) gin.H {
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
		"id":            strconv.FormatInt(q.ID, 10),
		"exam_id":       strconv.FormatInt(q.ExamID, 10),
		"question_text": q.QuestionText,
		"sort_order":    q.SortOrder,
		"choices":       choiceItems,
	}
}

// examPublicJSON is the Reader-facing shape — no is_correct anywhere.
func examPublicJSON(exam *repository.Exam, questions []repository.ExamQuestion, choices []repository.ExamChoice) gin.H {
	byQuestion := groupChoicesByQuestion(choices)
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
			"id":            strconv.FormatInt(q.ID, 10),
			"question_text": q.QuestionText,
			"sort_order":    q.SortOrder,
			"choices":       choiceItems,
		})
	}
	return gin.H{
		"id":            strconv.FormatInt(exam.ID, 10),
		"lesson_id":     strconv.FormatInt(exam.LessonID, 10),
		"title":         exam.Title,
		"passing_score": exam.PassingScore,
		"questions":     questionItems,
	}
}

func groupChoicesByQuestion(choices []repository.ExamChoice) map[int64][]repository.ExamChoice {
	byQuestion := make(map[int64][]repository.ExamChoice)
	for _, ch := range choices {
		byQuestion[ch.QuestionID] = append(byQuestion[ch.QuestionID], ch)
	}
	return byQuestion
}
