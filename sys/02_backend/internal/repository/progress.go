package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LessonProgress mirrors `user_lesson_progress` (dev-plan-02-database 2.5).
type LessonProgress struct {
	ID          int64
	UserID      int64
	LessonID    int64
	CompletedAt time.Time
}

// ExamResult mirrors `user_exam_results`.
type ExamResult struct {
	ID          int64
	UserID      int64
	ExamID      int64
	Score       int
	Passed      bool
	SubmittedAt time.Time
}

// ProgressRepository queries `user_lesson_progress` and `user_exam_results`.
type ProgressRepository struct {
	db *pgxpool.Pool
}

func NewProgressRepository(db *pgxpool.Pool) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// MarkLessonComplete upserts completion for (userID, lessonID), refreshing
// completed_at if it was already marked complete.
func (r *ProgressRepository) MarkLessonComplete(ctx context.Context, userID, lessonID int64) (*LessonProgress, error) {
	var p LessonProgress
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_lesson_progress (user_id, lesson_id) VALUES ($1, $2)
		ON CONFLICT (user_id, lesson_id) DO UPDATE SET completed_at = now()
		RETURNING id, user_id, lesson_id, completed_at
	`, userID, lessonID).Scan(&p.ID, &p.UserID, &p.LessonID, &p.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CompletedLessons returns completed_at for each of userID's completed
// lessons among lessonIDs, keyed by lesson_id.
func (r *ProgressRepository) CompletedLessons(ctx context.Context, userID int64, lessonIDs []int64) (map[int64]time.Time, error) {
	result := make(map[int64]time.Time)
	if len(lessonIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT lesson_id, completed_at FROM user_lesson_progress
		WHERE user_id = $1 AND lesson_id = ANY($2)
	`, userID, lessonIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lessonID int64
		var completedAt time.Time
		if err := rows.Scan(&lessonID, &completedAt); err != nil {
			return nil, err
		}
		result[lessonID] = completedAt
	}
	return result, rows.Err()
}

// CountCompletedByCourse returns how many of a course's lessons this user
// has completed, for the /api/me/progress summary.
func (r *ProgressRepository) CountCompletedByCourse(ctx context.Context, userID, courseID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM user_lesson_progress p
		JOIN lessons l ON l.id = p.lesson_id
		WHERE p.user_id = $1 AND l.course_id = $2
	`, userID, courseID).Scan(&count)
	return count, err
}

// SaveExamResult appends a new attempt. dev-plan-06 6.4 decision: history is
// kept (no upsert) so a user's attempts over time are preserved; course
// progress and /api/me/progress surface the latest attempt.
func (r *ProgressRepository) SaveExamResult(ctx context.Context, userID, examID int64, score int, passed bool) (*ExamResult, error) {
	var res ExamResult
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_exam_results (user_id, exam_id, score, passed) VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, exam_id, score, passed, submitted_at
	`, userID, examID, score, passed).Scan(&res.ID, &res.UserID, &res.ExamID, &res.Score, &res.Passed, &res.SubmittedAt)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// LatestExamResult returns a user's most recent attempt at an exam, or
// ErrNotFound if they haven't attempted it.
func (r *ProgressRepository) LatestExamResult(ctx context.Context, userID, examID int64) (*ExamResult, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, exam_id, score, passed, submitted_at FROM user_exam_results
		WHERE user_id = $1 AND exam_id = $2
		ORDER BY submitted_at DESC LIMIT 1
	`, userID, examID)

	var res ExamResult
	err := row.Scan(&res.ID, &res.UserID, &res.ExamID, &res.Score, &res.Passed, &res.SubmittedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &res, nil
}
