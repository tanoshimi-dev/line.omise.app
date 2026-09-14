package service

import (
	"errors"
	"testing"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

// Question with a single correct choice (10) among 10/11.
func fixtureSingleChoiceQuestion() []repository.QuizChoice {
	return []repository.QuizChoice{
		{ID: 10, IsCorrect: true},
		{ID: 11, IsCorrect: false},
	}
}

// Question with two correct choices (20, 21) among 20/21/22.
func fixtureMultiChoiceQuestion() []repository.QuizChoice {
	return []repository.QuizChoice{
		{ID: 20, IsCorrect: true},
		{ID: 21, IsCorrect: true},
		{ID: 22, IsCorrect: false},
	}
}

func TestGradeQuizQuestion_SingleChoiceCorrect(t *testing.T) {
	isCorrect, correctIDs, err := GradeQuizQuestion(false, fixtureSingleChoiceQuestion(), []int64{10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isCorrect {
		t.Error("isCorrect = false, want true")
	}
	if len(correctIDs) != 1 || correctIDs[0] != 10 {
		t.Errorf("correctIDs = %v, want [10]", correctIDs)
	}
}

func TestGradeQuizQuestion_SingleChoiceIncorrect(t *testing.T) {
	isCorrect, _, err := GradeQuizQuestion(false, fixtureSingleChoiceQuestion(), []int64{11})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isCorrect {
		t.Error("isCorrect = true, want false")
	}
}

func TestGradeQuizQuestion_Unanswered(t *testing.T) {
	isCorrect, _, err := GradeQuizQuestion(false, fixtureSingleChoiceQuestion(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isCorrect {
		t.Error("isCorrect = true, want false (unanswered)")
	}
}

func TestGradeQuizQuestion_TooManyChoicesWhenNotAllowMultiple(t *testing.T) {
	_, _, err := GradeQuizQuestion(false, fixtureMultiChoiceQuestion(), []int64{20, 21})
	if !errors.Is(err, ErrTooManyChoicesSelected) {
		t.Errorf("err = %v, want ErrTooManyChoicesSelected", err)
	}
}

func TestGradeQuizQuestion_InvalidChoiceID(t *testing.T) {
	_, _, err := GradeQuizQuestion(false, fixtureSingleChoiceQuestion(), []int64{999})
	if !errors.Is(err, ErrInvalidQuizAnswer) {
		t.Errorf("err = %v, want ErrInvalidQuizAnswer", err)
	}
}

func TestGradeQuizQuestion_MultiChoiceExactSetRequired(t *testing.T) {
	choices := fixtureMultiChoiceQuestion()

	if isCorrect, _, err := GradeQuizQuestion(true, choices, []int64{20, 21}); err != nil || !isCorrect {
		t.Errorf("full correct set: isCorrect = %v, err = %v, want true/nil", isCorrect, err)
	}
	if isCorrect, _, err := GradeQuizQuestion(true, choices, []int64{20}); err != nil || isCorrect {
		t.Errorf("partial correct set: isCorrect = %v, err = %v, want false/nil (no partial credit)", isCorrect, err)
	}
	if isCorrect, _, err := GradeQuizQuestion(true, choices, []int64{20, 21, 22}); err != nil || isCorrect {
		t.Errorf("correct set plus a wrong choice: isCorrect = %v, err = %v, want false/nil", isCorrect, err)
	}
}

func TestGradeQuizExam_ScoreAndPassing(t *testing.T) {
	questions := []repository.QuizQuestion{{ID: 1, AllowMultiple: false}, {ID: 2, AllowMultiple: false}}
	choicesByQuestion := map[int64][]repository.QuizChoice{
		1: fixtureSingleChoiceQuestion(),
		2: {{ID: 20, IsCorrect: true}, {ID: 21, IsCorrect: false}},
	}
	answers := []QuizExamAnswer{{QuestionID: 1, ChoiceIDs: []int64{10}}, {QuestionID: 2, ChoiceIDs: []int64{21}}}
	passingScore := 60

	grade, err := GradeQuizExam(questions, choicesByQuestion, answers, &passingScore)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if grade.Score != 1 || grade.TotalQuestions != 2 {
		t.Fatalf("score/total = %d/%d, want 1/2", grade.Score, grade.TotalQuestions)
	}
	if grade.Passed == nil || *grade.Passed {
		t.Errorf("passed = %v, want false (50%% < 60%%)", grade.Passed)
	}
}

func TestGradeQuizExam_UnansweredQuestionCountsIncorrect(t *testing.T) {
	questions := []repository.QuizQuestion{{ID: 1, AllowMultiple: false}, {ID: 2, AllowMultiple: false}}
	choicesByQuestion := map[int64][]repository.QuizChoice{
		1: fixtureSingleChoiceQuestion(),
		2: {{ID: 20, IsCorrect: true}, {ID: 21, IsCorrect: false}},
	}
	answers := []QuizExamAnswer{{QuestionID: 1, ChoiceIDs: []int64{10}}} // question 2 never answered

	grade, err := GradeQuizExam(questions, choicesByQuestion, answers, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if grade.Score != 1 {
		t.Errorf("score = %d, want 1 (unanswered question 2 counts as incorrect)", grade.Score)
	}
	if grade.Passed != nil {
		t.Errorf("passed = %v, want nil (no passing_score configured)", grade.Passed)
	}
}

func TestGradeQuizExam_InvalidQuestionID(t *testing.T) {
	questions := []repository.QuizQuestion{{ID: 1, AllowMultiple: false}}
	choicesByQuestion := map[int64][]repository.QuizChoice{1: fixtureSingleChoiceQuestion()}
	answers := []QuizExamAnswer{{QuestionID: 999, ChoiceIDs: []int64{10}}}

	_, err := GradeQuizExam(questions, choicesByQuestion, answers, nil)
	if !errors.Is(err, ErrInvalidQuizAnswer) {
		t.Errorf("err = %v, want ErrInvalidQuizAnswer", err)
	}
}

func TestGradeQuizExam_PassingScoreBoundary(t *testing.T) {
	questions := []repository.QuizQuestion{{ID: 1}, {ID: 2}}
	choicesByQuestion := map[int64][]repository.QuizChoice{
		1: fixtureSingleChoiceQuestion(),
		2: {{ID: 20, IsCorrect: true}, {ID: 21, IsCorrect: false}},
	}
	answers := []QuizExamAnswer{{QuestionID: 1, ChoiceIDs: []int64{10}}, {QuestionID: 2, ChoiceIDs: []int64{21}}} // 50%

	at50 := 50
	grade, _ := GradeQuizExam(questions, choicesByQuestion, answers, &at50)
	if grade.Passed == nil || !*grade.Passed {
		t.Error("score equal to passing score should pass (>=), got false")
	}

	at51 := 51
	grade, _ = GradeQuizExam(questions, choicesByQuestion, answers, &at51)
	if grade.Passed == nil || *grade.Passed {
		t.Error("score below passing score should fail, got true")
	}
}
