'use client'

// Exam-mode quiz runner (dev-plan-2-4-frontend-quiz-ui 2-4.3): shows one
// question at a time (dev-plan-quiz-question-pagination), buffering answers
// client-side with no feedback until the whole thing is submitted via
// POST /api/quizzes/:slug/submit (dev-plan-2-3 2-3.3), then renders the
// final score/pass-fail plus a per-question review. Works whether the
// caller is logged in or not — only the login prompt on the result screen
// differs. Any quiz can be run this way — the reader picks single mode vs
// exam mode per attempt (dev-plan-quiz-mode-selection).

import { useState } from 'react'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import LoginPrompt from '@/components/learn/LoginPrompt'
import type { Quiz, QuizSubmitResult } from '@/lib/types'

export default function ExamRunner({ quiz }: { quiz: Quiz }) {
  const { user } = useAuth()
  const [index, setIndex] = useState(0)
  const [answers, setAnswers] = useState<Record<string, string[]>>({})
  const [result, setResult] = useState<QuizSubmitResult | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [startedAt] = useState(() => new Date().toISOString())

  const questions = quiz.questions
  const answeredCount = questions.filter((q) => (answers[q.id]?.length ?? 0) > 0).length

  const toggleChoice = (questionId: string, choiceId: string, allowMultiple: boolean) => {
    setAnswers((prev) => {
      const current = prev[questionId] ?? []
      if (allowMultiple) {
        return {
          ...prev,
          [questionId]: current.includes(choiceId) ? current.filter((id) => id !== choiceId) : [...current, choiceId],
        }
      }
      return { ...prev, [questionId]: [choiceId] }
    })
  }

  const handleSubmit = async () => {
    const unanswered = questions.length - answeredCount
    if (unanswered > 0 && !window.confirm(`未回答の設問が${unanswered}問あります。このまま提出しますか？`)) {
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const payload = {
        started_at: startedAt,
        answers: questions.map((q) => ({
          question_id: Number(q.id),
          choice_ids: (answers[q.id] ?? []).map(Number),
        })),
      }
      const res = await api.post<QuizSubmitResult>(`/api/quizzes/${quiz.slug}/submit`, payload)
      setResult(res)
    } catch {
      setError('提出に失敗しました。')
    } finally {
      setSubmitting(false)
    }
  }

  if (result) {
    return <ExamResult quiz={quiz} result={result} loggedIn={!!user} />
  }

  if (questions.length === 0) {
    return <p className="text-gray-500">このクイズにはまだ設問がありません。</p>
  }

  const question = questions[index]
  const isLast = index === questions.length - 1

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <p className="text-sm text-gray-400">
        {index + 1} / {questions.length}問
      </p>
      <fieldset className="mt-2">
        <legend className="font-medium text-gray-900">{question.question_text}</legend>
        <div className="mt-3 space-y-2">
          {question.choices.map((choice) => (
            <label key={choice.id} className="flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700">
              <input
                type={question.allow_multiple ? 'checkbox' : 'radio'}
                name={`question-${question.id}`}
                checked={(answers[question.id] ?? []).includes(choice.id)}
                onChange={() => toggleChoice(question.id, choice.id, question.allow_multiple)}
                className="accent-line-green"
              />
              {choice.choice_text}
            </label>
          ))}
        </div>
      </fieldset>

      {error && <p className="mt-3 text-sm text-red-600">{error}</p>}

      <div className="mt-4 flex items-center gap-3">
        {index > 0 && (
          <button
            type="button"
            onClick={() => setIndex((i) => i - 1)}
            className="rounded-full border-2 border-line-green px-6 py-2.5 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white"
          >
            前の問題へ
          </button>
        )}
        {isLast ? (
          <button
            onClick={() => void handleSubmit()}
            disabled={submitting}
            className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
          >
            {submitting ? '採点中…' : '提出する'}
          </button>
        ) : (
          <button
            type="button"
            onClick={() => setIndex((i) => i + 1)}
            className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark"
          >
            次の問題へ
          </button>
        )}
      </div>
    </div>
  )
}

function ExamResult({ quiz, result, loggedIn }: { quiz: Quiz; result: QuizSubmitResult; loggedIn: boolean }) {
  return (
    <div className="space-y-6">
      <div className="rounded-2xl border border-gray-100 bg-white p-8 text-center shadow-sm">
        <p className="text-2xl font-bold text-gray-900">
          {result.score} / {result.total_questions}問正解
        </p>
        {result.passed !== null && (
          <p className={`mt-2 font-semibold ${result.passed ? 'text-line-green' : 'text-red-600'}`}>
            {result.passed ? '合格' : '不合格'}
            {quiz.passing_score != null && `（合格点: ${quiz.passing_score}%）`}
          </p>
        )}
        {!loggedIn && (
          <div className="mt-6">
            <LoginPrompt message="ログインすると受験履歴・進捗が保存されます。" />
          </div>
        )}
      </div>

      <div className="space-y-4">
        {result.questions.map((q, qi) => (
          <div
            key={q.question_id}
            className={`rounded-xl border p-4 ${q.is_correct ? 'border-line-green/30 bg-green-50' : 'border-red-200 bg-red-50'}`}
          >
            <p className="font-medium text-gray-900">
              {qi + 1}. {q.question_text}
            </p>
            <p className={`mt-1 text-sm font-semibold ${q.is_correct ? 'text-line-green' : 'text-red-600'}`}>
              {q.is_correct ? '正解' : '不正解'}
            </p>
            <p className="mt-2 whitespace-pre-line text-sm text-gray-600">{q.explanation}</p>
            {q.reference_url && (
              <a href={q.reference_url} target="_blank" rel="noreferrer" className="mt-1 inline-block text-xs text-line-green hover:underline">
                参考リンク
              </a>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
