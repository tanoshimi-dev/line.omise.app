package service

import (
	"errors"
	"testing"
)

// Two questions, each with two choices; question 1's correct choice is 10,
// question 2's is 20.
func fixtureExam() (correctByQuestion map[int64]int64, validChoicesByQuestion map[int64]map[int64]bool) {
	correctByQuestion = map[int64]int64{1: 10, 2: 20}
	validChoicesByQuestion = map[int64]map[int64]bool{
		1: {10: true, 11: true},
		2: {20: true, 21: true},
	}
	return
}

func TestGradeExam_AllCorrect(t *testing.T) {
	correct, valid := fixtureExam()
	answers := []ExamAnswer{{QuestionID: 1, ChoiceID: 10}, {QuestionID: 2, ChoiceID: 20}}

	score, passed, err := GradeExam(answers, correct, valid, 70)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 100 {
		t.Errorf("score = %d, want 100", score)
	}
	if !passed {
		t.Error("passed = false, want true")
	}
}

func TestGradeExam_PartialCorrect(t *testing.T) {
	correct, valid := fixtureExam()
	answers := []ExamAnswer{{QuestionID: 1, ChoiceID: 10}, {QuestionID: 2, ChoiceID: 21}} // 1 of 2 correct

	score, passed, err := GradeExam(answers, correct, valid, 70)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 50 {
		t.Errorf("score = %d, want 50", score)
	}
	if passed {
		t.Error("passed = true, want false (50 < passing score 70)")
	}
}

func TestGradeExam_UnansweredQuestionCountsIncorrect(t *testing.T) {
	correct, valid := fixtureExam()
	answers := []ExamAnswer{{QuestionID: 1, ChoiceID: 10}} // question 2 never answered

	score, _, err := GradeExam(answers, correct, valid, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 50 {
		t.Errorf("score = %d, want 50 (unanswered question 2 counts as incorrect)", score)
	}
}

func TestGradeExam_DuplicateAnswerKeepsFirst(t *testing.T) {
	correct, valid := fixtureExam()
	// Two answers for question 1: first correct, second (should be ignored) wrong.
	answers := []ExamAnswer{
		{QuestionID: 1, ChoiceID: 10},
		{QuestionID: 1, ChoiceID: 11},
		{QuestionID: 2, ChoiceID: 20},
	}

	score, passed, err := GradeExam(answers, correct, valid, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 100 || !passed {
		t.Errorf("score = %d passed = %v, want 100/true (first answer for a repeated question should count)", score, passed)
	}
}

func TestGradeExam_InvalidQuestionID(t *testing.T) {
	correct, valid := fixtureExam()
	answers := []ExamAnswer{{QuestionID: 999, ChoiceID: 10}}

	_, _, err := GradeExam(answers, correct, valid, 0)
	if !errors.Is(err, ErrInvalidAnswer) {
		t.Errorf("err = %v, want ErrInvalidAnswer", err)
	}
}

func TestGradeExam_ChoiceDoesNotBelongToQuestion(t *testing.T) {
	correct, valid := fixtureExam()
	// Choice 20 belongs to question 2, not question 1.
	answers := []ExamAnswer{{QuestionID: 1, ChoiceID: 20}}

	_, _, err := GradeExam(answers, correct, valid, 0)
	if !errors.Is(err, ErrInvalidAnswer) {
		t.Errorf("err = %v, want ErrInvalidAnswer", err)
	}
}

func TestGradeExam_NoQuestions(t *testing.T) {
	_, _, err := GradeExam(nil, map[int64]int64{}, map[int64]map[int64]bool{}, 0)
	if !errors.Is(err, ErrNoQuestions) {
		t.Errorf("err = %v, want ErrNoQuestions", err)
	}
}

func TestGradeExam_PassingScoreBoundary(t *testing.T) {
	correct, valid := fixtureExam()
	answers := []ExamAnswer{{QuestionID: 1, ChoiceID: 10}, {QuestionID: 2, ChoiceID: 21}} // score = 50

	_, passedAt50, _ := GradeExam(answers, correct, valid, 50)
	if !passedAt50 {
		t.Error("score equal to passing score should pass (>=), got false")
	}

	_, passedAt51, _ := GradeExam(answers, correct, valid, 51)
	if passedAt51 {
		t.Error("score below passing score should fail, got true")
	}
}
