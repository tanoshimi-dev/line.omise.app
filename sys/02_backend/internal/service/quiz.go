package service

import (
	"errors"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
)

var (
	// ErrTooFewChoices is returned when a question has fewer than two
	// choices (dev-plan-2-2-admin-api 2-2.3).
	ErrTooFewChoices = errors.New("each question needs at least two choices")

	// ErrNoCorrectChoice is returned when no choice is marked correct.
	ErrNoCorrectChoice = errors.New("at least one choice must be marked correct")

	// ErrMultipleCorrectNotAllowed is returned when allow_multiple is false
	// but more than one choice is marked correct.
	ErrMultipleCorrectNotAllowed = errors.New("allow_multiple is false but more than one choice is marked correct")
)

// ValidateQuestionChoices checks a question's choices against dev-plan-2-2
// 2-2.3: at least two choices, at least one correct, and — unless the
// question allows multiple answers — at most one correct.
func ValidateQuestionChoices(allowMultiple bool, choices []repository.QuizChoiceInput) error {
	if len(choices) < 2 {
		return ErrTooFewChoices
	}
	correctCount := 0
	for _, c := range choices {
		if c.IsCorrect {
			correctCount++
		}
	}
	if correctCount == 0 {
		return ErrNoCorrectChoice
	}
	if !allowMultiple && correctCount > 1 {
		return ErrMultipleCorrectNotAllowed
	}
	return nil
}
