package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Quiz mirrors the `quizzes` table (dev-plan-2-1-db-migration). Mode
// distinguishes a practice-mode quiz (per-question immediate feedback) from
// an exam-mode quiz (bulk submit, final score) — see dev-plan-2-2-admin-api
// background note on terminology.
type Quiz struct {
	ID           int64
	Slug         string
	Title        string
	Description  string
	Mode         string
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
		SELECT id, slug, title, COALESCE(description, ''), mode, passing_score, published
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
		SELECT id, slug, title, COALESCE(description, ''), mode, passing_score, published
		FROM quizzes WHERE id = $1
	`, id)
	return scanQuiz(row)
}

func (r *QuizRepository) Create(ctx context.Context, slug, title, description, mode string, passingScore *int, published bool) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO quizzes (slug, title, description, mode, passing_score, published)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, slug, title, COALESCE(description, ''), mode, passing_score, published
	`, slug, title, description, mode, passingScore, published)
	return scanQuiz(row)
}

func (r *QuizRepository) Update(ctx context.Context, id int64, slug, title, description, mode string, passingScore *int, published bool) (*Quiz, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE quizzes SET slug = $2, title = $3, description = $4, mode = $5, passing_score = $6, published = $7
		WHERE id = $1
		RETURNING id, slug, title, COALESCE(description, ''), mode, passing_score, published
	`, id, slug, title, description, mode, passingScore, published)
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
	err := row.Scan(&q.ID, &q.Slug, &q.Title, &q.Description, &q.Mode, &q.PassingScore, &q.Published)
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
