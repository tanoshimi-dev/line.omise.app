package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func TestMarkLessonComplete_UpsertsRatherThanDuplicates(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	courses := repository.NewCourseRepository(pool)
	progress := repository.NewProgressRepository(pool)

	lessonID := setupLesson(t, courses)
	user := testutil.CreateUser(t, pool, "reader")

	first, err := progress.MarkLessonComplete(ctx, user.ID, lessonID)
	if err != nil {
		t.Fatalf("first MarkLessonComplete: %v", err)
	}
	second, err := progress.MarkLessonComplete(ctx, user.ID, lessonID)
	if err != nil {
		t.Fatalf("second MarkLessonComplete: %v", err)
	}

	completed, err := progress.CompletedLessons(ctx, user.ID, []int64{lessonID})
	if err != nil {
		t.Fatalf("CompletedLessons: %v", err)
	}
	if len(completed) != 1 {
		t.Errorf("completed lessons = %+v, want exactly one row (upsert, not duplicate)", completed)
	}
	if !second.CompletedAt.After(first.CompletedAt) && !second.CompletedAt.Equal(first.CompletedAt) {
		t.Errorf("second completed_at (%v) should be >= first (%v)", second.CompletedAt, first.CompletedAt)
	}
}

func TestCompletedLessons_OnlyReturnsThisUsersRows(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	courses := repository.NewCourseRepository(pool)
	progress := repository.NewProgressRepository(pool)

	lessonID := setupLesson(t, courses)
	userA := testutil.CreateUser(t, pool, "reader")
	userB := testutil.CreateUser(t, pool, "reader")

	if _, err := progress.MarkLessonComplete(ctx, userA.ID, lessonID); err != nil {
		t.Fatalf("MarkLessonComplete for userA: %v", err)
	}

	completedForB, err := progress.CompletedLessons(ctx, userB.ID, []int64{lessonID})
	if err != nil {
		t.Fatalf("CompletedLessons for userB: %v", err)
	}
	if len(completedForB) != 0 {
		t.Errorf("userB's completed lessons = %+v, want none (userA's progress must not leak)", completedForB)
	}
}

// dev-plan-06 6.4: exam history is kept, not overwritten; the progress view
// surfaces the latest attempt.
func TestLatestExamResult_ReturnsMostRecentAttempt(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	progress := repository.NewProgressRepository(pool)

	lessonID := setupLesson(t, courses)
	exam, err := exams.Create(ctx, lessonID, "Quiz", 70)
	if err != nil {
		t.Fatalf("Create exam: %v", err)
	}
	user := testutil.CreateUser(t, pool, "reader")

	if _, err := progress.SaveExamResult(ctx, user.ID, exam.ID, 100, true); err != nil {
		t.Fatalf("first SaveExamResult: %v", err)
	}
	if _, err := progress.SaveExamResult(ctx, user.ID, exam.ID, 50, false); err != nil {
		t.Fatalf("second SaveExamResult: %v", err)
	}

	latest, err := progress.LatestExamResult(ctx, user.ID, exam.ID)
	if err != nil {
		t.Fatalf("LatestExamResult: %v", err)
	}
	if latest.Score != 50 || latest.Passed {
		t.Errorf("latest result = %+v, want the second (most recent) attempt: score=50 passed=false", latest)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM user_exam_results WHERE user_id = $1", user.ID).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 2 {
		t.Errorf("stored exam result rows = %d, want 2 (history kept, not overwritten)", count)
	}
}

func TestLatestExamResult_NotFoundWhenNeverAttempted(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	ctx := context.Background()
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	progress := repository.NewProgressRepository(pool)

	lessonID := setupLesson(t, courses)
	exam, err := exams.Create(ctx, lessonID, "Quiz", 70)
	if err != nil {
		t.Fatalf("Create exam: %v", err)
	}
	user := testutil.CreateUser(t, pool, "reader")

	_, err = progress.LatestExamResult(ctx, user.ID, exam.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
