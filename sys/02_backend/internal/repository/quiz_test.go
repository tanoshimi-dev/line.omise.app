package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/testutil"
)

func intPtr(v int) *int { return &v }

func TestQuizCreate_DuplicateSlugFails(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	if _, err := quizzes.Create(ctx, "line-quiz", "LINE Quiz", "", "practice", nil, true); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := quizzes.Create(ctx, "line-quiz", "Another", "", "practice", nil, true); err == nil {
		t.Error("expected a unique-constraint error creating a second quiz with the same slug, got nil")
	}
}

func TestQuizCreate_PassingScoreRoundTripsAsNil(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "practice-quiz", "Practice Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if quiz.PassingScore != nil {
		t.Errorf("PassingScore = %v, want nil", *quiz.PassingScore)
	}

	fetched, err := quizzes.GetByID(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.PassingScore != nil {
		t.Errorf("fetched PassingScore = %v, want nil", *fetched.PassingScore)
	}
}

func TestQuizCreate_PassingScoreRoundTripsAsValue(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "exam-quiz", "Exam Quiz", "", "exam", intPtr(80), true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if quiz.PassingScore == nil || *quiz.PassingScore != 80 {
		t.Fatalf("PassingScore = %v, want 80", quiz.PassingScore)
	}

	fetched, err := quizzes.GetByID(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.PassingScore == nil || *fetched.PassingScore != 80 {
		t.Fatalf("fetched PassingScore = %v, want 80", fetched.PassingScore)
	}
}

func TestQuizCreateQuestion_InsertsQuestionAndChoicesAtomically(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "quiz", "Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	choices := []repository.QuizChoiceInput{
		{ChoiceText: "Right", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "Wrong", IsCorrect: false, SortOrder: 2},
	}
	question, inserted, err := quizzes.CreateQuestion(ctx, quiz.ID, "Q1?", false, "Because.", "https://example.com", 1, choices)
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}
	if len(inserted) != 2 {
		t.Fatalf("inserted choices = %d, want 2", len(inserted))
	}
	if question.ReferenceURL != "https://example.com" {
		t.Errorf("ReferenceURL = %q, want https://example.com", question.ReferenceURL)
	}

	allChoices, err := quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("ListChoicesForQuiz: %v", err)
	}
	if len(allChoices) != 2 {
		t.Errorf("choices for quiz = %+v, want 2", allChoices)
	}
}

func TestQuizCreateQuestion_EmptyReferenceURLRoundTripsAsEmptyString(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "quiz", "Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	question, _, err := quizzes.CreateQuestion(ctx, quiz.ID, "Q1?", false, "Because.", "", 1, []repository.QuizChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "B", IsCorrect: false, SortOrder: 2},
	})
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}
	if question.ReferenceURL != "" {
		t.Errorf("ReferenceURL = %q, want empty", question.ReferenceURL)
	}
}

func TestQuizUpdateQuestion_ReplacesChoicesEntirely(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "quiz", "Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	question, _, err := quizzes.CreateQuestion(ctx, quiz.ID, "Q1?", false, "Because.", "", 1, []repository.QuizChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "B", IsCorrect: false, SortOrder: 2},
	})
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}

	_, newChoices, err := quizzes.UpdateQuestion(ctx, question.ID, "Q1 updated?", true, "Updated.", "", 1, []repository.QuizChoiceInput{
		{ChoiceText: "C", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "D", IsCorrect: true, SortOrder: 2},
		{ChoiceText: "E", IsCorrect: false, SortOrder: 3},
	})
	if err != nil {
		t.Fatalf("UpdateQuestion: %v", err)
	}
	if len(newChoices) != 3 {
		t.Fatalf("choices after update = %d, want 3 (old ones replaced)", len(newChoices))
	}

	allChoices, err := quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("ListChoicesForQuiz: %v", err)
	}
	if len(allChoices) != 3 {
		t.Errorf("choices for quiz after update = %d, want exactly 3 (old A/B must be gone)", len(allChoices))
	}
}

func TestQuizDeleteQuestion_CascadesChoices(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "quiz", "Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	question, _, err := quizzes.CreateQuestion(ctx, quiz.ID, "Q1?", false, "Because.", "", 1, []repository.QuizChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "B", IsCorrect: false, SortOrder: 2},
	})
	if err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}

	if err := quizzes.DeleteQuestion(ctx, question.ID); err != nil {
		t.Fatalf("DeleteQuestion: %v", err)
	}

	choices, err := quizzes.ListChoicesForQuiz(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("ListChoicesForQuiz: %v", err)
	}
	if len(choices) != 0 {
		t.Errorf("choices after deleting question = %+v, want none (cascade)", choices)
	}
}

func TestQuizDeleteQuestion_NotFoundForUnknownID(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)

	err := quizzes.DeleteQuestion(context.Background(), 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestQuizDelete_CascadesQuestionsAndChoices(t *testing.T) {
	pool := testutil.TestDB(t)
	t.Cleanup(func() { testutil.TruncateAll(t, pool) })
	quizzes := repository.NewQuizRepository(pool)
	ctx := context.Background()

	quiz, err := quizzes.Create(ctx, "quiz", "Quiz", "", "practice", nil, true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := quizzes.CreateQuestion(ctx, quiz.ID, "Q1?", false, "Because.", "", 1, []repository.QuizChoiceInput{
		{ChoiceText: "A", IsCorrect: true, SortOrder: 1},
		{ChoiceText: "B", IsCorrect: false, SortOrder: 2},
	}); err != nil {
		t.Fatalf("CreateQuestion: %v", err)
	}

	if err := quizzes.Delete(ctx, quiz.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	questions, err := quizzes.ListQuestions(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("ListQuestions: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("questions after deleting quiz = %+v, want none (cascade)", questions)
	}
}
