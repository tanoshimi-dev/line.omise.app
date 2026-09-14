'use client'

// Practice-mode quiz runner (dev-plan-2-4-frontend-quiz-ui 2-4.2): shows one
// question at a time, grades it immediately via POST /api/quiz-questions/:id
// /answer, and reveals the explanation before moving to the next question.
// Works whether the caller is logged in or not (dev-plan-2-3 2-3.2) — only
// the login prompt at the end differs.

import { useState } from 'react'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import LoginPrompt from '@/components/learn/LoginPrompt'
import type { Quiz, QuizQuestion, QuizAnswerResult } from '@/lib/types'

export default function QuizQuestionCard({ quiz }: { quiz: Quiz }) {
  const { user } = useAuth()
  const [index, setIndex] = useState(0)
  const [correctCount, setCorrectCount] = useState(0)
  const [finished, setFinished] = useState(false)
  const questions = quiz.questions

  if (questions.length === 0) {
    return <p className="text-gray-500">このクイズにはまだ設問がありません。</p>
  }

  if (finished) {
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

  return (
    <QuestionStep
      key={question.id}
      question={question}
      questionNumber={index + 1}
      totalQuestions={questions.length}
      onNext={(isCorrect) => {
        setCorrectCount((c) => c + (isCorrect ? 1 : 0))
        if (index + 1 < questions.length) {
          setIndex(index + 1)
        } else {
          setFinished(true)
        }
      }}
    />
  )
}

function QuestionStep({
  question,
  questionNumber,
  totalQuestions,
  onNext,
}: {
  question: QuizQuestion
  questionNumber: number
  totalQuestions: number
  onNext: (isCorrect: boolean) => void
}) {
  const [selected, setSelected] = useState<string[]>([])
  const [result, setResult] = useState<QuizAnswerResult | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const toggle = (choiceId: string) => {
    if (result) return
    if (question.allow_multiple) {
      setSelected((prev) => (prev.includes(choiceId) ? prev.filter((id) => id !== choiceId) : [...prev, choiceId]))
    } else {
      setSelected([choiceId])
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
      setResult(res)
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

      {!result ? (
        <button
          onClick={() => void submit()}
          disabled={selected.length === 0 || submitting}
          className="mt-4 rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
        >
          {submitting ? '採点中…' : '解答する'}
        </button>
      ) : (
        <div className="mt-4 space-y-3">
          <p className={`font-semibold ${result.is_correct ? 'text-line-green' : 'text-red-600'}`}>
            {result.is_correct ? '正解です！' : '不正解です'}
          </p>
          <p className="text-sm text-gray-600">{result.explanation}</p>
          {result.reference_url && (
            <a href={result.reference_url} target="_blank" rel="noreferrer" className="inline-block text-sm text-line-green hover:underline">
              参考リンク
            </a>
          )}
          <div>
            <button
              onClick={() => onNext(result.is_correct)}
              className="rounded-full border-2 border-line-green px-6 py-2.5 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white"
            >
              次の問題へ
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
