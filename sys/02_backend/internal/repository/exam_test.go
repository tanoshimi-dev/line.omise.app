package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func setupLesson(t *testing.T, courses *repository.CourseRepository) int64 {
	t.Helper()
	ctx := context.Background()
	course, err := courses.Create(ctx, "course", "Course", "", 1, "published")
	if err != nil {
		t.Fatalf("Create course: %v", err)
	}
	lesson, err := courses.CreateLesson(ctx, course.ID, "lesson", "Lesson", "", 1, "published")
	if err != nil {
		t.Fatalf("CreateLesson: %v", err)
	}
	return lesson.ID
}

func TestExamCreate_OneExamPerLesson(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	lessonID := setupLesson(t, courses)

	if _, err := exams.Create(context.Background(), lessonID, "Quiz", 70); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := exams.Create(context.Background(), lessonID, "Quiz 2", 50); err == nil {
		t.Error("expected a unique-constraint error creating a second exam for the same lesson, got nil")
	}
}

func TestCreateQuestion_InsertsQuestionAndChoicesAtomically(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	ctx := context.Background()
	lessonID := setupLesson(t, courses)

	exam, err := exams.Create(ctx, lessonID, "Quiz", 70)
	if err != nil {
		t.Fatalf("Create exam: %v", err)
	}

	choices := []repository.ChoiceInput{
		{ChoiceText: "Right", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "Wrong", IsCorrect: false, SortOrder: 2},
	}
	question, inserted, err := exams.CreateQuestion(ctx, exam.ID, "Q1?", 1, choices)
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}
	if len(inserted) != 2 {
		t.Fatalf("inserted choices = %d, want 2", len(inserted))
	}

	allChoices, err := exams.ListChoicesForExam(ctx, exam.ID)
	if err != nil {
		t.Fatalf("ListChoicesForExam: %v", err)
	}
	if len(allChoices) != 2 {
		t.Errorf("choices for exam = %+v, want 2", allChoices)
	}
	for _, c := range allChoices {
		if c.QuestionID != question.ID {
			t.Errorf("choice %+v has wrong question_id, want %d", c, question.ID)
		}
	}
}

func TestUpdateQuestion_ReplacesChoicesEntirely(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	ctx := context.Background()
	lessonID := setupLesson(t, courses)

	exam, err := exams.Create(ctx, lessonID, "Quiz", 70)
	if err != nil {
		t.Fatalf("Create exam: %v", err)
	}
	question, _, err := exams.CreateQuestion(ctx, exam.ID, "Q1?", 1, []repository.ChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "B", IsCorrect: false, SortOrder: 2},
	})
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}

	_, newChoices, err := exams.UpdateQuestion(ctx, question.ID, "Q1 updated?", 1, []repository.ChoiceInput{
		{ChoiceText: "C", IsCorrect: false, SortOrder: 1},
		{ChoiceText: "D", IsCorrect: true, SortOrder: 2},
		{ChoiceText: "E", IsCorrect: false, SortOrder: 3},
	})
	if err != nil {
		t.Fatalf("UpdateQuestion: %v", err)
	}
	if len(newChoices) != 3 {
		t.Fatalf("choices after update = %d, want 3 (old ones replaced)", len(newChoices))
	}

	allChoices, err := exams.ListChoicesForExam(ctx, exam.ID)
	if err != nil {
		t.Fatalf("ListChoicesForExam: %v", err)
	}
	if len(allChoices) != 3 {
		t.Errorf("choices for exam after update = %d, want exactly 3 (old A/B must be gone)", len(allChoices))
	}
}

func TestDeleteQuestion_CascadesChoices(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	courses := repository.NewCourseRepository(pool)
	exams := repository.NewExamRepository(pool)
	ctx := context.Background()
	lessonID := setupLesson(t, courses)

	exam, err := exams.Create(ctx, lessonID, "Quiz", 70)
	if err != nil {
		t.Fatalf("Create exam: %v", err)
	}
	question, _, err := exams.CreateQuestion(ctx, exam.ID, "Q1?", 1, []repository.ChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
	})
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}

	if err := exams.DeleteQuestion(ctx, question.ID); err != nil {
		t.Fatalf("DeleteQuestion: %v", err)
	}

	choices, err := exams.ListChoicesForExam(ctx, exam.ID)
	if err != nil {
		t.Fatalf("ListChoicesForExam: %v", err)
	}
	if len(choices) != 0 {
		t.Errorf("choices after deleting question = %+v, want none (cascade)", choices)
	}
}

func TestDeleteQuestion_NotFoundForUnknownID(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	exams := repository.NewExamRepository(pool)

	err := exams.DeleteQuestion(context.Background(), 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
