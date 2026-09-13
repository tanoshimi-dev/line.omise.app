package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Exam mirrors the `exams` table (dev-plan-02-database 2.4). Each lesson has
// at most one exam (migration 005 adds UNIQUE(lesson_id)).
type Exam struct {
	ID           int64
	LessonID     int64
	Title        string
	PassingScore int
}

// ExamQuestion mirrors `exam_questions`.
type ExamQuestion struct {
	ID           int64
	ExamID       int64
	QuestionText string
	SortOrder    int
}

// ExamChoice mirrors `exam_choices`.
type ExamChoice struct {
	ID         int64
	QuestionID int64
	ChoiceText string
	IsCorrect  bool
	SortOrder  int
}

// ChoiceInput is one choice supplied when creating/updating a question.
type ChoiceInput struct {
	ChoiceText string
	IsCorrect  bool
	SortOrder  int
}

// ExamRepository queries `exams`, `exam_questions` and `exam_choices`.
type ExamRepository struct {
	db *pgxpool.Pool
}

func NewExamRepository(db *pgxpool.Pool) *ExamRepository {
	return &ExamRepository{db: db}
}

func (r *ExamRepository) GetByLessonID(ctx context.Context, lessonID int64) (*Exam, error) {
	row := r.db.QueryRow(ctx, `SELECT id, lesson_id, title, passing_score FROM exams WHERE lesson_id = $1`, lessonID)
	return scanExam(row)
}

func (r *ExamRepository) GetByID(ctx context.Context, id int64) (*Exam, error) {
	row := r.db.QueryRow(ctx, `SELECT id, lesson_id, title, passing_score FROM exams WHERE id = $1`, id)
	return scanExam(row)
}

// Create inserts an exam for a lesson. Returns AsExamAlreadyExists-wrapped
// error (via the raw unique-violation, translated by the caller/service) if
// the lesson already has one.
func (r *ExamRepository) Create(ctx context.Context, lessonID int64, title string, passingScore int) (*Exam, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO exams (lesson_id, title, passing_score) VALUES ($1, $2, $3)
		RETURNING id, lesson_id, title, passing_score
	`, lessonID, title, passingScore)
	return scanExam(row)
}

// ListQuestions returns an exam's questions ordered for display.
func (r *ExamRepository) ListQuestions(ctx context.Context, examID int64) ([]ExamQuestion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, exam_id, question_text, sort_order FROM exam_questions
		WHERE exam_id = $1 ORDER BY sort_order, id
	`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []ExamQuestion
	for rows.Next() {
		var q ExamQuestion
		if err := rows.Scan(&q.ID, &q.ExamID, &q.QuestionText, &q.SortOrder); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

func (r *ExamRepository) GetQuestionByID(ctx context.Context, id int64) (*ExamQuestion, error) {
	row := r.db.QueryRow(ctx, `SELECT id, exam_id, question_text, sort_order FROM exam_questions WHERE id = $1`, id)
	var q ExamQuestion
	if err := row.Scan(&q.ID, &q.ExamID, &q.QuestionText, &q.SortOrder); err != nil {
		return nil, wrapNotFound(err)
	}
	return &q, nil
}

// ListChoicesForExam returns every choice for every question in an exam, in
// one query — used both for the public (is_correct-hidden) response and for
// grading (which needs is_correct).
func (r *ExamRepository) ListChoicesForExam(ctx context.Context, examID int64) ([]ExamChoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.question_id, c.choice_text, c.is_correct, c.sort_order
		FROM exam_choices c
		JOIN exam_questions q ON q.id = c.question_id
		WHERE q.exam_id = $1
		ORDER BY c.question_id, c.sort_order, c.id
	`, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []ExamChoice
	for rows.Next() {
		var ch ExamChoice
		if err := rows.Scan(&ch.ID, &ch.QuestionID, &ch.ChoiceText, &ch.IsCorrect, &ch.SortOrder); err != nil {
			return nil, err
		}
		choices = append(choices, ch)
	}
	return choices, rows.Err()
}

// CreateQuestion inserts a question and its choices atomically.
func (r *ExamRepository) CreateQuestion(ctx context.Context, examID int64, text string, sortOrder int, choices []ChoiceInput) (*ExamQuestion, []ExamChoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var q ExamQuestion
	err = tx.QueryRow(ctx, `
		INSERT INTO exam_questions (exam_id, question_text, sort_order) VALUES ($1, $2, $3)
		RETURNING id, exam_id, question_text, sort_order
	`, examID, text, sortOrder).Scan(&q.ID, &q.ExamID, &q.QuestionText, &q.SortOrder)
	if err != nil {
		return nil, nil, wrapNotFound(err)
	}

	inserted, err := insertChoices(ctx, tx, q.ID, choices)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &q, inserted, nil
}

// UpdateQuestion updates a question's text/order and fully replaces its
// choices (matching the PUT-replaces-the-resource convention used
// throughout dev-plan-05-content-api).
func (r *ExamRepository) UpdateQuestion(ctx context.Context, id int64, text string, sortOrder int, choices []ChoiceInput) (*ExamQuestion, []ExamChoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var q ExamQuestion
	err = tx.QueryRow(ctx, `
		UPDATE exam_questions SET question_text = $2, sort_order = $3 WHERE id = $1
		RETURNING id, exam_id, question_text, sort_order
	`, id, text, sortOrder).Scan(&q.ID, &q.ExamID, &q.QuestionText, &q.SortOrder)
	if err != nil {
		return nil, nil, wrapNotFound(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM exam_choices WHERE question_id = $1`, id); err != nil {
		return nil, nil, err
	}

	inserted, err := insertChoices(ctx, tx, id, choices)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &q, inserted, nil
}

func (r *ExamRepository) DeleteQuestion(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM exam_questions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func insertChoices(ctx context.Context, tx pgx.Tx, questionID int64, choices []ChoiceInput) ([]ExamChoice, error) {
	inserted := make([]ExamChoice, 0, len(choices))
	for _, in := range choices {
		var ch ExamChoice
		err := tx.QueryRow(ctx, `
			INSERT INTO exam_choices (question_id, choice_text, is_correct, sort_order)
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

func scanExam(row rowScanner) (*Exam, error) {
	var e Exam
	err := row.Scan(&e.ID, &e.LessonID, &e.Title, &e.PassingScore)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &e, nil
}
