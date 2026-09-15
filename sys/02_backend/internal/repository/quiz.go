package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Quiz mirrors the `quizzes` table (dev-plan-2-1-db-migration). A reader
// chooses per-attempt whether to answer question-by-question (immediate
// feedback) or submit the whole quiz at once (final score) — see
// dev-plan-quiz-mode-selection.
type Quiz struct {
	ID           int64
	Slug         string
	Title        string
	Description  string
	PassingScore *int
	Published    bool
}

// QuizQuestion mirrors `quiz_questions`.
type QuizQuestion struct {
	ID            int64
	QuizID        int64
	QuestionText  string
	AllowMultiple bool
	Explanation   string
	ReferenceURL  string
	SortOrder     int
}

// QuizChoice mirrors `quiz_choices`.
type QuizChoice struct {
	ID         int64
	QuestionID int64
	ChoiceText string
	IsCorrect  bool
	SortOrder  int
}

// QuizChoiceInput is one choice supplied when creating/updating a question.
type QuizChoiceInput struct {
	ChoiceText string
	IsCorrect  bool
	SortOrder  int
}

// QuizRepository queries `quizzes`, `quiz_questions` and `quiz_choices`.
type QuizRepository struct {
	db *pgxpool.Pool
}

func NewQuizRepository(db *pgxpool.Pool) *QuizRepository {
	return &QuizRepository{db: db}
}

// ListAll returns every quiz regardless of mode/published state, for the
// admin quiz list (dev-plan-2-2 2-2.1).
func (r *QuizRepository) ListAll(ctx context.Context) ([]Quiz, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), passing_score, published
		FROM quizzes
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []Quiz
	for rows.Next() {
		q, err := scanQuiz(rows)
		if err != nil {
			return nil, err
		}
		quizzes = append(quizzes, *q)
	}
	return quizzes, rows.Err()
}

func (r *QuizRepository) GetByID(ctx context.Context, id int64) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), passing_score, published
		FROM quizzes WHERE id = $1
	`, id)
	return scanQuiz(row)
}

// ListPublished returns published quizzes/exams ordered for display
// (dev-plan-2-3 2-3.4's "GET /api/me/quizzes/progress" summary lists every
// published quiz, mirroring CourseRepository.ListPublished).
func (r *QuizRepository) ListPublished(ctx context.Context) ([]Quiz, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), passing_score, published
		FROM quizzes WHERE published = true
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []Quiz
	for rows.Next() {
		q, err := scanQuiz(rows)
		if err != nil {
			return nil, err
		}
		quizzes = append(quizzes, *q)
	}
	return quizzes, rows.Err()
}

// GetPublishedBySlug returns a published quiz, or ErrNotFound if it doesn't
// exist or isn't published (dev-plan-2-3 2-3.1 — draft quizzes aren't
// reachable through the public API).
func (r *QuizRepository) GetPublishedBySlug(ctx context.Context, slug string) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, title, COALESCE(description, ''), passing_score, published
		FROM quizzes WHERE slug = $1 AND published = true
	`, slug)
	return scanQuiz(row)
}

func (r *QuizRepository) Create(ctx context.Context, slug, title, description string, passingScore *int, published bool) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO quizzes (slug, title, description, passing_score, published)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, slug, title, COALESCE(description, ''), passing_score, published
	`, slug, title, description, passingScore, published)
	return scanQuiz(row)
}

func (r *QuizRepository) Update(ctx context.Context, id int64, slug, title, description string, passingScore *int, published bool) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE quizzes SET slug = $2, title = $3, description = $4, passing_score = $5, published = $6
		WHERE id = $1
		RETURNING id, slug, title, COALESCE(description, ''), passing_score, published
	`, id, slug, title, description, passingScore, published)
	return scanQuiz(row)
}

func (r *QuizRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM quizzes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListQuestions returns a quiz's questions ordered for display.
func (r *QuizRepository) ListQuestions(ctx context.Context, quizID int64) ([]QuizQuestion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, quiz_id, question_text, allow_multiple, explanation, COALESCE(reference_url, ''), sort_order
		FROM quiz_questions
		WHERE quiz_id = $1 ORDER BY sort_order, id
	`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []QuizQuestion
	for rows.Next() {
		var q QuizQuestion
		if err := rows.Scan(&q.ID, &q.QuizID, &q.QuestionText, &q.AllowMultiple, &q.Explanation, &q.ReferenceURL, &q.SortOrder); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetQuestionByID returns a single question, for the practice-mode answer
// endpoint (dev-plan-2-3 2-3.2) which is addressed by question id rather
// than by quiz slug.
func (r *QuizRepository) GetQuestionByID(ctx context.Context, id int64) (*QuizQuestion, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, quiz_id, question_text, allow_multiple, explanation, COALESCE(reference_url, ''), sort_order
		FROM quiz_questions WHERE id = $1
	`, id)
	var q QuizQuestion
	if err := row.Scan(&q.ID, &q.QuizID, &q.QuestionText, &q.AllowMultiple, &q.Explanation, &q.ReferenceURL, &q.SortOrder); err != nil {
		return nil, wrapNotFound(err)
	}
	return &q, nil
}

// ListChoicesForQuestion returns one question's choices, for grading a
// practice-mode answer (dev-plan-2-3 2-3.2).
func (r *QuizRepository) ListChoicesForQuestion(ctx context.Context, questionID int64) ([]QuizChoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, question_id, choice_text, is_correct, sort_order
		FROM quiz_choices WHERE question_id = $1
		ORDER BY sort_order, id
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []QuizChoice
	for rows.Next() {
		var ch QuizChoice
		if err := rows.Scan(&ch.ID, &ch.QuestionID, &ch.ChoiceText, &ch.IsCorrect, &ch.SortOrder); err != nil {
			return nil, err
		}
		choices = append(choices, ch)
	}
	return choices, rows.Err()
}

// ListChoicesForQuiz returns every choice for every question in a quiz, in
// one query — used to embed choices (including is_correct) in the admin
// quiz detail response.
func (r *QuizRepository) ListChoicesForQuiz(ctx context.Context, quizID int64) ([]QuizChoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.question_id, c.choice_text, c.is_correct, c.sort_order
		FROM quiz_choices c
		JOIN quiz_questions q ON q.id = c.question_id
		WHERE q.quiz_id = $1
		ORDER BY c.question_id, c.sort_order, c.id
	`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []QuizChoice
	for rows.Next() {
		var ch QuizChoice
		if err := rows.Scan(&ch.ID, &ch.QuestionID, &ch.ChoiceText, &ch.IsCorrect, &ch.SortOrder); err != nil {
			return nil, err
		}
		choices = append(choices, ch)
	}
	return choices, rows.Err()
}

// CreateQuestion inserts a question and its choices atomically.
func (r *QuizRepository) CreateQuestion(ctx context.Context, quizID int64, text string, allowMultiple bool, explanation, referenceURL string, sortOrder int, choices []QuizChoiceInput) (*QuizQuestion, []QuizChoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var q QuizQuestion
	err = tx.QueryRow(ctx, `
		INSERT INTO quiz_questions (quiz_id, question_text, allow_multiple, explanation, reference_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, quiz_id, question_text, allow_multiple, explanation, COALESCE(reference_url, ''), sort_order
	`, quizID, text, allowMultiple, explanation, nullableString(referenceURL), sortOrder).
		Scan(&q.ID, &q.QuizID, &q.QuestionText, &q.AllowMultiple, &q.Explanation, &q.ReferenceURL, &q.SortOrder)
	if err != nil {
		return nil, nil, wrapNotFound(err)
	}

	inserted, err := insertQuizChoices(ctx, tx, q.ID, choices)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &q, inserted, nil
}

// UpdateQuestion updates a question's fields and fully replaces its choices
// (matching the PUT-replaces-the-resource convention used for exam
// questions — dev-plan-06-exam-progress-api).
func (r *QuizRepository) UpdateQuestion(ctx context.Context, id int64, text string, allowMultiple bool, explanation, referenceURL string, sortOrder int, choices []QuizChoiceInput) (*QuizQuestion, []QuizChoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var q QuizQuestion
	err = tx.QueryRow(ctx, `
		UPDATE quiz_questions
		SET question_text = $2, allow_multiple = $3, explanation = $4, reference_url = $5, sort_order = $6
		WHERE id = $1
		RETURNING id, quiz_id, question_text, allow_multiple, explanation, COALESCE(reference_url, ''), sort_order
	`, id, text, allowMultiple, explanation, nullableString(referenceURL), sortOrder).
		Scan(&q.ID, &q.QuizID, &q.QuestionText, &q.AllowMultiple, &q.Explanation, &q.ReferenceURL, &q.SortOrder)
	if err != nil {
		return nil, nil, wrapNotFound(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM quiz_choices WHERE question_id = $1`, id); err != nil {
		return nil, nil, err
	}

	inserted, err := insertQuizChoices(ctx, tx, id, choices)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &q, inserted, nil
}

func (r *QuizRepository) DeleteQuestion(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM quiz_questions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func insertQuizChoices(ctx context.Context, tx pgx.Tx, questionID int64, choices []QuizChoiceInput) ([]QuizChoice, error) {
	inserted := make([]QuizChoice, 0, len(choices))
	for _, in := range choices {
		var ch QuizChoice
		err := tx.QueryRow(ctx, `
			INSERT INTO quiz_choices (question_id, choice_text, is_correct, sort_order)
			VALUES ($1, $2, $3, $4)
			RETURNING id, question_id, choice_text, is_correct, sort_order
		`, questionID, in.ChoiceText, in.IsCorrect, in.SortOrder).Scan(&ch.ID, &ch.QuestionID, &ch.ChoiceText, &ch.IsCorrect, &ch.SortOrder)
		if err != nil {
			return nil, err
		}
		inserted = append(inserted, ch)
	}
	return inserted, nil
}

func scanQuiz(row rowScanner) (*Quiz, error) {
	var q Quiz
	err := row.Scan(&q.ID, &q.Slug, &q.Title, &q.Description, &q.PassingScore, &q.Published)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &q, nil
}

// nullableString returns nil for an empty string so an optional text column
// (reference_url) is stored as SQL NULL rather than an empty string.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// UserQuizAttempt mirrors `user_quiz_attempts` — one exam-mode submission
// (dev-plan-2-1-db-migration). Passed is nil when the quiz has no
// passing_score set.
type UserQuizAttempt struct {
	ID             int64
	UserID         int64
	QuizID         int64
	Score          int
	TotalQuestions int
	Passed         *bool
	StartedAt      time.Time
	SubmittedAt    time.Time
}

// UserQuizAnswer mirrors `user_quiz_answers`. AttemptID is nil for a
// practice-mode answer and set for an exam-mode answer recorded as part of
// a CreateAttempt submission.
type UserQuizAnswer struct {
	ID                int64
	UserID            int64
	QuestionID        int64
	AttemptID         *int64
	SelectedChoiceIDs []int64
	IsCorrect         bool
	AnsweredAt        time.Time
}

// SaveAnswer records one answered question — attemptID is nil for a
// practice-mode answer (dev-plan-2-3 2-3.2), set for an exam-mode answer
// belonging to a CreateAttempt submission (dev-plan-2-3 2-3.3).
func (r *QuizRepository) SaveAnswer(ctx context.Context, userID, questionID int64, attemptID *int64, selectedChoiceIDs []int64, isCorrect bool) (*UserQuizAnswer, error) {
	var a UserQuizAnswer
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_quiz_answers (user_id, question_id, attempt_id, selected_choice_ids, is_correct)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, question_id, attempt_id, selected_choice_ids, is_correct, answered_at
	`, userID, questionID, attemptID, selectedChoiceIDs, isCorrect).
		Scan(&a.ID, &a.UserID, &a.QuestionID, &a.AttemptID, &a.SelectedChoiceIDs, &a.IsCorrect, &a.AnsweredAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateAttempt inserts one exam-mode submission (dev-plan-2-3 2-3.3).
func (r *QuizRepository) CreateAttempt(ctx context.Context, userID, quizID int64, score, totalQuestions int, passed *bool, startedAt time.Time) (*UserQuizAttempt, error) {
	var a UserQuizAttempt
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_quiz_attempts (user_id, quiz_id, score, total_questions, passed, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, quiz_id, score, total_questions, passed, started_at, submitted_at
	`, userID, quizID, score, totalQuestions, passed, startedAt).
		Scan(&a.ID, &a.UserID, &a.QuizID, &a.Score, &a.TotalQuestions, &a.Passed, &a.StartedAt, &a.SubmittedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListPracticeHistory returns a user's practice-mode answers (attempt_id
// IS NULL) for a quiz, newest first (dev-plan-2-3 2-3.4).
func (r *QuizRepository) ListPracticeHistory(ctx context.Context, userID, quizID int64) ([]UserQuizAnswer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.user_id, a.question_id, a.attempt_id, a.selected_choice_ids, a.is_correct, a.answered_at
		FROM user_quiz_answers a
		JOIN quiz_questions q ON q.id = a.question_id
		WHERE a.user_id = $1 AND q.quiz_id = $2 AND a.attempt_id IS NULL
		ORDER BY a.answered_at DESC
	`, userID, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanQuizAnswers(rows)
}

// ListAttempts returns a user's exam-mode attempts for a quiz, newest first
// (dev-plan-2-3 2-3.4).
func (r *QuizRepository) ListAttempts(ctx context.Context, userID, quizID int64) ([]UserQuizAttempt, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, quiz_id, score, total_questions, passed, started_at, submitted_at
		FROM user_quiz_attempts
		WHERE user_id = $1 AND quiz_id = $2
		ORDER BY submitted_at DESC
	`, userID, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []UserQuizAttempt
	for rows.Next() {
		var a UserQuizAttempt
		if err := rows.Scan(&a.ID, &a.UserID, &a.QuizID, &a.Score, &a.TotalQuestions, &a.Passed, &a.StartedAt, &a.SubmittedAt); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

// GetOwnAttempt returns attemptID, but only if it belongs to userID —
// scoping ownership at the query level so a wrong owner gets the same
// ErrNotFound as a nonexistent id, never leaking whether the attempt exists
// (dev-plan-2-3 2-3.5).
func (r *QuizRepository) GetOwnAttempt(ctx context.Context, attemptID, userID int64) (*UserQuizAttempt, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, quiz_id, score, total_questions, passed, started_at, submitted_at
		FROM user_quiz_attempts WHERE id = $1 AND user_id = $2
	`, attemptID, userID)
	var a UserQuizAttempt
	err := row.Scan(&a.ID, &a.UserID, &a.QuizID, &a.Score, &a.TotalQuestions, &a.Passed, &a.StartedAt, &a.SubmittedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &a, nil
}

// DeleteOwnAttempt deletes one exam-mode attempt, but only if it belongs to
// userID — same ownership-scoped-at-the-query-level pattern as
// GetOwnAttempt (dev-plan-quiz-history-delete). Its user_quiz_answers rows
// cascade-delete via the FK (007_quiz.up.sql).
func (r *QuizRepository) DeleteOwnAttempt(ctx context.Context, attemptID, userID int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM user_quiz_attempts WHERE id = $1 AND user_id = $2`, attemptID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAttemptsForQuiz deletes every exam-mode attempt userID has for
// quizID (dev-plan-quiz-history-delete's bulk "clear history" action).
// Deleting an already-empty history is not an error.
func (r *QuizRepository) DeleteAttemptsForQuiz(ctx context.Context, userID, quizID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_quiz_attempts WHERE user_id = $1 AND quiz_id = $2`, userID, quizID)
	return err
}

// DeleteOwnPracticeAnswer deletes one question-by-question answer (attempt_id
// IS NULL), but only if it belongs to userID.
func (r *QuizRepository) DeleteOwnPracticeAnswer(ctx context.Context, answerID, userID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM user_quiz_answers WHERE id = $1 AND user_id = $2 AND attempt_id IS NULL
	`, answerID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeletePracticeHistoryForQuiz deletes every question-by-question answer
// userID has for quizID. Deleting an already-empty history is not an error.
func (r *QuizRepository) DeletePracticeHistoryForQuiz(ctx context.Context, userID, quizID int64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM user_quiz_answers
		WHERE user_id = $1 AND attempt_id IS NULL
		AND question_id IN (SELECT id FROM quiz_questions WHERE quiz_id = $2)
	`, userID, quizID)
	return err
}

// ListAnswersForAttempt returns every answer recorded as part of one
// exam-mode attempt (dev-plan-2-3 2-3.4's attempt detail view).
func (r *QuizRepository) ListAnswersForAttempt(ctx context.Context, attemptID int64) ([]UserQuizAnswer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, question_id, attempt_id, selected_choice_ids, is_correct, answered_at
		FROM user_quiz_answers WHERE attempt_id = $1
		ORDER BY id
	`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanQuizAnswers(rows)
}

func scanQuizAnswers(rows pgx.Rows) ([]UserQuizAnswer, error) {
	var answers []UserQuizAnswer
	for rows.Next() {
		var a UserQuizAnswer
		if err := rows.Scan(&a.ID, &a.UserID, &a.QuestionID, &a.AttemptID, &a.SelectedChoiceIDs, &a.IsCorrect, &a.AnsweredAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}
