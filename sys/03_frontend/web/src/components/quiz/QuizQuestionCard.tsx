'use client'

// Single-mode quiz runner (dev-plan-2-4-frontend-quiz-ui 2-4.2): shows one
// question at a time, grades it immediately via POST /api/quiz-questions/:id
// /answer, and reveals the explanation. Learners can move between questions
// without answering, matching the exam-mode navigation behavior.
// Works whether the caller is logged in or not (dev-plan-2-3 2-3.2) — only
// the login prompt at the end differs. Any quiz can be run this way — the
// reader picks single mode vs exam mode per attempt (dev-plan-quiz-mode-selection).
//
// Selection/result state lives here, keyed by question id, rather than
// inside QuestionStep, so navigating back with 前の問題へ
// (dev-plan-quiz-question-pagination) redisplays an already-graded
// question's result without re-submitting it.

import { useState } from 'react'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import LoginPrompt from '@/components/learn/LoginPrompt'
import type { Quiz, QuizQuestion, QuizAnswerResult } from '@/lib/types'

export default function QuizQuestionCard({ quiz }: { quiz: Quiz }) {
  const { user } = useAuth()
  const [index, setIndex] = useState(0)
  const [selectedByQuestion, setSelectedByQuestion] = useState<Record<string, string[]>>({})
  const [resultByQuestion, setResultByQuestion] = useState<Record<string, QuizAnswerResult>>({})
  const [finished, setFinished] = useState(false)
  const questions = quiz.questions

  if (questions.length === 0) {
    return <p className="text-gray-500">このクイズにはまだ設問がありません。</p>
  }

  if (finished) {
    const correctCount = Object.values(resultByQuestion).filter((r) => r.is_correct).length
    return (
      <div className="rounded-2xl border border-gray-100 bg-white p-8 text-center shadow-sm">
        <p className="text-2xl font-bold text-gray-900">
          {questions.length}問中 {correctCount}問正解しました
        </p>
        {!user && (
          <div className="mt-6">
            <LoginPrompt message="ログインすると解答履歴・正答率が記録されます。" />
          </div>
        )}
      </div>
    )
  }

  const question = questions[index]
  const isLast = index === questions.length - 1

  return (
    <QuestionStep
      key={question.id}
      question={question}
      questionNumber={index + 1}
      totalQuestions={questions.length}
      selected={selectedByQuestion[question.id] ?? []}
      result={resultByQuestion[question.id] ?? null}
      isLast={isLast}
      onSelectedChange={(selected) => setSelectedByQuestion((prev) => ({ ...prev, [question.id]: selected }))}
      onAnswered={(result) => setResultByQuestion((prev) => ({ ...prev, [question.id]: result }))}
      onPrev={index > 0 ? () => setIndex(index - 1) : undefined}
      onNext={() => {
        if (isLast) {
          setFinished(true)
        } else {
          setIndex(index + 1)
        }
      }}
    />
  )
}

function QuestionStep({
  question,
  questionNumber,
  totalQuestions,
  selected,
  result,
  isLast,
  onSelectedChange,
  onAnswered,
  onPrev,
  onNext,
}: {
  question: QuizQuestion
  questionNumber: number
  totalQuestions: number
  selected: string[]
  result: QuizAnswerResult | null
  isLast: boolean
  onSelectedChange: (selected: string[]) => void
  onAnswered: (result: QuizAnswerResult) => void
  onPrev?: () => void
  onNext: () => void
}) {
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const toggle = (choiceId: string) => {
    if (result) return
    if (question.allow_multiple) {
      onSelectedChange(selected.includes(choiceId) ? selected.filter((id) => id !== choiceId) : [...selected, choiceId])
    } else {
      onSelectedChange([choiceId])
    }
  }

  const submit = async () => {
    if (selected.length === 0) return
    setSubmitting(true)
    setError(null)
    try {
      const res = await api.post<QuizAnswerResult>(`/api/quiz-questions/${question.id}/answer`, {
        choice_ids: selected.map(Number),
      })
      onAnswered(res)
    } catch {
      setError('採点に失敗しました。')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <p className="text-sm text-gray-400">
        {questionNumber} / {totalQuestions}問
      </p>
      <fieldset className="mt-2" disabled={!!result}>
        <legend className="font-medium text-gray-900">{question.question_text}</legend>
        <div className="mt-3 space-y-2">
          {question.choices.map((choice) => {
            const isSelected = selected.includes(choice.id)
            const isCorrectChoice = result?.correct_choice_ids.includes(choice.id) ?? false
            const isWrongSelected = !!result && isSelected && !isCorrectChoice
            const stateClass = result
              ? isCorrectChoice
                ? 'border-line-green bg-green-50 text-line-green'
                : isWrongSelected
                  ? 'border-red-300 bg-red-50 text-red-700'
                  : 'border-gray-200 text-gray-500'
              : 'border-gray-200 text-gray-700'
            return (
              <label key={choice.id} className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-sm ${stateClass}`}>
                <input
                  type={question.allow_multiple ? 'checkbox' : 'radio'}
                  name={`question-${question.id}`}
                  checked={isSelected}
                  onChange={() => toggle(choice.id)}
                  className="accent-line-green"
                />
                {choice.choice_text}
              </label>
            )
          })}
        </div>
      </fieldset>

      {error && <p className="mt-3 text-sm text-red-600">{error}</p>}

      {result && (
        <div className="mt-4 space-y-3">
          <p className={`font-semibold ${result.is_correct ? 'text-line-green' : 'text-red-600'}`}>
            {result.is_correct ? '正解です！' : '不正解です'}
          </p>
          <p className="whitespace-pre-line text-sm text-gray-600">{result.explanation}</p>
          {result.reference_url && (
            <a href={result.reference_url} target="_blank" rel="noreferrer" className="inline-block text-sm text-line-green hover:underline">
              参考リンク
            </a>
          )}
        </div>
      )}

      <div className="mt-4 flex items-center gap-3">
        {onPrev && (
          <button
            type="button"
            onClick={onPrev}
            className="rounded-full border-2 border-line-green px-6 py-2.5 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white"
          >
            前の問題へ
          </button>
        )}
        {!result && (
          <button
            onClick={() => void submit()}
            disabled={selected.length === 0 || submitting}
            className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
          >
            {submitting ? '採点中…' : '解答する'}
          </button>
        )}
        <button
          type="button"
          onClick={onNext}
          className="rounded-full border-2 border-line-green px-6 py-2.5 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white"
        >
          {isLast ? '結果を見る' : '次の問題へ'}
        </button>
      </div>
    </div>
  )
}
