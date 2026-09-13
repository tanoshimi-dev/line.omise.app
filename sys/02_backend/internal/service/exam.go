package service

import (
	"errors"
	"math"
)

var (
	// ErrExamAlreadyExists is returned when a lesson already has an exam
	// (dev-plan-02-database migration 005 adds a UNIQUE(lesson_id) constraint
	// so each lesson has at most one exam, matching the plan's singular
	// "/lessons/:lessonId/exam" routes).
	ErrExamAlreadyExists = errors.New("lesson already has an exam")

	// ErrInvalidAnswer is returned when a submitted answer references a
	// question or choice that doesn't belong to the exam being submitted.
	ErrInvalidAnswer = errors.New("invalid answer")

	// ErrNoQuestions is returned when grading an exam that has no questions.
	ErrNoQuestions = errors.New("exam has no questions")
)

// AsExamAlreadyExists returns ErrExamAlreadyExists if err is a violation of
// the exams.lesson_id unique constraint, and the original err otherwise.
func AsExamAlreadyExists(err error) error {
	if isUniqueViolation(err, "exams_lesson_id_key") {
		return ErrExamAlreadyExists
	}
	return err
}

// ExamAnswer is one submitted answer (dev-plan-06 6.2).
type ExamAnswer struct {
	QuestionID int64
	ChoiceID   int64
}

// GradeExam scores answers against the exam's questions/choices.
//
// correctChoiceByQuestion maps question_id -> its single correct choice_id
// (dev-plan-06-exam-progress-api assumes one correct choice per question,
// matching the plan's { question_id, choice_id } answer shape).
// validChoicesByQuestion maps question_id -> the set of choice_ids that
// actually belong to that question, used to reject a mismatched pairing
// (e.g. a choice_id from a different question) with ErrInvalidAnswer rather
// than silently mis-scoring it.
//
// Score is a percentage (0-100); passed is score >= passingScore. A
// question with no submitted answer counts as incorrect. Answering the same
// question twice keeps only the first answer for that question.
func GradeExam(answers []ExamAnswer, correctChoiceByQuestion map[int64]int64, validChoicesByQuestion map[int64]map[int64]bool, passingScore int) (score int, passed bool, err error) {
	totalQuestions := len(validChoicesByQuestion)
	if totalQuestions == 0 {
		return 0, false, ErrNoQuestions
	}

	answered := make(map[int64]bool, len(answers))
	correct := 0
	for _, a := range answers {
		validChoices, ok := validChoicesByQuestion[a.QuestionID]
		if !ok {
			return 0, false, ErrInvalidAnswer
		}
		if !validChoices[a.ChoiceID] {
			return 0, false, ErrInvalidAnswer
		}
		if answered[a.QuestionID] {
			continue
		}
		answered[a.QuestionID] = true
		if correctChoiceByQuestion[a.QuestionID] == a.ChoiceID {
			correct++
		}
	}

	score = int(math.Round(float64(correct) / float64(totalQuestions) * 100))
	passed = score >= passingScore
	return score, passed, nil
}
