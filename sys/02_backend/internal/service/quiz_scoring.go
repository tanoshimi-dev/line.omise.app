package service

import (
	"errors"
	"math"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

var (
	// ErrQuizModeMismatch is returned when a request hits the wrong
	// endpoint for a quiz's mode — the practice answer endpoint against an
	// exam-mode quiz, or the exam submit endpoint against a practice-mode
	// quiz (dev-plan-2-3 2-3.2/2-3.3).
	ErrQuizModeMismatch = errors.New("this endpoint does not apply to this quiz's mode")

	// ErrInvalidQuizAnswer is returned when a selected choice id doesn't
	// belong to the question being answered, or when an answer references a
	// question that doesn't belong to the quiz being submitted.
	ErrInvalidQuizAnswer = errors.New("selected choice does not belong to the question")

	// ErrTooManyChoicesSelected is returned when more than one choice is
	// selected for a question that doesn't allow multiple answers.
	ErrTooManyChoicesSelected = errors.New("this question does not allow multiple selected choices")
)

// GradeQuizQuestion grades selectedChoiceIDs against one question's choices
// (dev-plan-2-3 2-3.5: an exact match of the selected set against the
// correct-choice set is required — no partial credit). An empty
// selectedChoiceIDs is graded as incorrect rather than rejected, so exam-mode
// submission (2-3.3) can treat an unanswered question as simply wrong
// instead of failing the whole submission.
func GradeQuizQuestion(allowMultiple bool, choices []repository.QuizChoice, selectedChoiceIDs []int64) (isCorrect bool, correctChoiceIDs []int64, err error) {
	valid := make(map[int64]bool, len(choices))
	for _, c := range choices {
		valid[c.ID] = true
		if c.IsCorrect {
			correctChoiceIDs = append(correctChoiceIDs, c.ID)
		}
	}

	if !allowMultiple && len(selectedChoiceIDs) > 1 {
		return false, correctChoiceIDs, ErrTooManyChoicesSelected
	}

	selectedSet := make(map[int64]bool, len(selectedChoiceIDs))
	for _, id := range selectedChoiceIDs {
		if !valid[id] {
			return false, correctChoiceIDs, ErrInvalidQuizAnswer
		}
		selectedSet[id] = true
	}

	if len(selectedSet) == 0 || len(selectedSet) != len(correctChoiceIDs) {
		return false, correctChoiceIDs, nil
	}
	for _, id := range correctChoiceIDs {
		if !selectedSet[id] {
			return false, correctChoiceIDs, nil
		}
	}
	return true, correctChoiceIDs, nil
}

// QuizExamAnswer is one submitted answer in an exam-mode bulk submission.
type QuizExamAnswer struct {
	QuestionID int64
	ChoiceIDs  []int64
}

// QuizExamQuestionGrade is one question's grading result within an exam-mode
// submission — used both to compute the total score and to render the
// post-submission review (dev-plan-2-3 2-3.3).
type QuizExamQuestionGrade struct {
	QuestionID        int64
	IsCorrect         bool
	SelectedChoiceIDs []int64
	CorrectChoiceIDs  []int64
}

// QuizExamGrade is the outcome of grading a whole exam-mode submission.
type QuizExamGrade struct {
	Score          int
	TotalQuestions int
	// Passed is nil when the quiz has no passing_score configured.
	Passed    *bool
	Questions []QuizExamQuestionGrade
}

// GradeQuizExam grades every question in a quiz for one exam-mode
// submission. A question with no entry in answers (or an empty ChoiceIDs)
// counts as incorrect rather than causing an error (dev-plan-2-3 2-3.3's
// "未回答は不正解扱い" decision) — an error is only returned for answers that
// are actually invalid: a question id that isn't in the quiz, a choice id
// that doesn't belong to its question, or more than one choice selected for
// a question that doesn't allow it.
func GradeQuizExam(questions []repository.QuizQuestion, choicesByQuestion map[int64][]repository.QuizChoice, answers []QuizExamAnswer, passingScore *int) (QuizExamGrade, error) {
	choiceIDsByQuestion := make(map[int64][]int64, len(answers))
	for _, a := range answers {
		if _, ok := choicesByQuestion[a.QuestionID]; !ok {
			return QuizExamGrade{}, ErrInvalidQuizAnswer
		}
		choiceIDsByQuestion[a.QuestionID] = a.ChoiceIDs
	}

	grade := QuizExamGrade{TotalQuestions: len(questions)}
	for _, q := range questions {
		isCorrect, correctIDs, err := GradeQuizQuestion(q.AllowMultiple, choicesByQuestion[q.ID], choiceIDsByQuestion[q.ID])
		if err != nil {
			return QuizExamGrade{}, err
		}
		if isCorrect {
			grade.Score++
		}
		grade.Questions = append(grade.Questions, QuizExamQuestionGrade{
			QuestionID:        q.ID,
			IsCorrect:         isCorrect,
			SelectedChoiceIDs: choiceIDsByQuestion[q.ID],
			CorrectChoiceIDs:  correctIDs,
		})
	}

	if passingScore != nil && grade.TotalQuestions > 0 {
		percentage := int(math.Round(float64(grade.Score) / float64(grade.TotalQuestions) * 100))
		passed := percentage >= *passingScore
		grade.Passed = &passed
	}
	return grade, nil
}
