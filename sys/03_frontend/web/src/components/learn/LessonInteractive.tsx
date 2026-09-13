'use client'

// Progress-saving and exam UI for a lesson (dev-plan-09-frontend-learn 9.2).
// A single client component so it can share one fetch of the course's
// progress (dev-plan-06's /api/courses/:slug/progress) to learn both whether
// this lesson is already completed and whether it has an exam + a prior
// result — rather than fetching each separately.

import { useEffect, useState } from 'react'
import { useAuth } from '@/lib/auth'
import { api, ApiError } from '@/lib/api'
import type { CourseProgress, CourseProgressLesson, Exam, ExamResult, SubmitExamResponse } from '@/lib/types'
import LoginPrompt from './LoginPrompt'

interface Props {
  courseSlug: string
  lessonId: string
}

export default function LessonInteractive({ courseSlug, lessonId }: Props) {
  const { user, loading: authLoading } = useAuth()
  const [progress, setProgress] = useState<CourseProgressLesson | null>(null)
  const [loadingProgress, setLoadingProgress] = useState(true)

  useEffect(() => {
    if (!user) {
      setLoadingProgress(false)
      return
    }
    let cancelled = false
    api
      .get<CourseProgress>(`/api/courses/${courseSlug}/progress`)
      .then((data) => {
        if (cancelled) return
        setProgress(data.lessons.find((l) => l.id === lessonId) ?? null)
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelled) setLoadingProgress(false)
      })
    return () => {
      cancelled = true
    }
  }, [user, courseSlug, lessonId])

  if (authLoading) {
    return null
  }

  if (!user) {
    return (
      <div className="mt-10 border-t border-gray-100 pt-8">
        <LoginPrompt message="ログインすると、レッスンの完了記録・試験の受験ができます。" />
      </div>
    )
  }

  if (loadingProgress) {
    return <p className="mt-10 border-t border-gray-100 pt-8 text-sm text-gray-400">読み込み中…</p>
  }

  return (
    <div className="mt-10 space-y-8 border-t border-gray-100 pt-8">
      <CompleteSection lessonId={lessonId} initiallyCompleted={progress?.completed ?? false} />
      {progress?.exam && <ExamSection lessonId={lessonId} examSummary={progress.exam} />}
    </div>
  )
}

function CompleteSection({ lessonId, initiallyCompleted }: { lessonId: string; initiallyCompleted: boolean }) {
  const [completed, setCompleted] = useState(initiallyCompleted)
  const [saving, setSaving] = useState(false)

  if (completed) {
    return (
      <div className="flex items-center gap-2 font-semibold text-line-green">
        <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        このレッスンは完了済みです
      </div>
    )
  }

  const handleClick = async () => {
    setSaving(true)
    try {
      await api.post(`/api/lessons/${lessonId}/complete`)
      setCompleted(true)
    } finally {
      setSaving(false)
    }
  }

  return (
    <button
      onClick={() => void handleClick()}
      disabled={saving}
      className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
    >
      {saving ? '記録中…' : 'レッスンを完了にする'}
    </button>
  )
}

interface ExamSummary {
  id: string
  title: string
  passing_score: number
  latest_result: ExamResult | null
}

function ExamSection({ lessonId, examSummary }: { lessonId: string; examSummary: ExamSummary }) {
  const [exam, setExam] = useState<Exam | null>(null)
  const [taking, setTaking] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [result, setResult] = useState<ExamResult | null>(examSummary.latest_result)

  const startExam = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await api.get<Exam>(`/api/lessons/${lessonId}/exam`)
      setExam(data)
      setAnswers({})
      setTaking(true)
    } catch {
      setError('試験の読み込みに失敗しました。')
    } finally {
      setLoading(false)
    }
  }

  const submit = async () => {
    if (!exam) return
    setLoading(true)
    setError(null)
    try {
      const payload = {
        answers: exam.questions.map((q) => ({
          question_id: Number(q.id),
          choice_id: Number(answers[q.id]),
        })),
      }
      const res = await api.post<SubmitExamResponse>(`/api/lessons/${lessonId}/exam/submit`, payload)
      setResult(res)
      setTaking(false)
    } catch (err) {
      setError(err instanceof ApiError && err.status === 400 ? 'すべての設問に回答してください。' : '採点に失敗しました。')
    } finally {
      setLoading(false)
    }
  }

  const allAnswered = exam ? exam.questions.every((q) => answers[q.id] != null) : false

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <h3 className="text-lg font-bold text-gray-900">{examSummary.title}</h3>
      <p className="mt-1 text-sm text-gray-500">合格点: {examSummary.passing_score}点</p>

      {error && <p className="mt-3 text-sm text-red-600">{error}</p>}

      {!taking && (
        <div className="mt-4">
          {result && (
            <p className={`mb-4 font-semibold ${result.passed ? 'text-line-green' : 'text-gray-700'}`}>
              前回のスコア: {result.score}点（{result.passed ? '合格' : '不合格'}）
            </p>
          )}
          <button
            onClick={() => void startExam()}
            disabled={loading}
            className="rounded-full border-2 border-line-green px-6 py-2.5 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white disabled:opacity-50"
          >
            {loading ? '読み込み中…' : result ? '再受験する' : '試験を受ける'}
          </button>
        </div>
      )}

      {taking && exam && (
        <div className="mt-6 space-y-6">
          {exam.questions.map((q, qi) => (
            <fieldset key={q.id}>
              <legend className="font-medium text-gray-900">
                {qi + 1}. {q.question_text}
              </legend>
              <div className="mt-2 space-y-2">
                {q.choices.map((choice) => (
                  <label key={choice.id} className="flex items-center gap-2 text-sm text-gray-700">
                    <input
                      type="radio"
                      name={`question-${q.id}`}
                      value={choice.id}
                      checked={answers[q.id] === choice.id}
                      onChange={() => setAnswers((prev) => ({ ...prev, [q.id]: choice.id }))}
                      className="accent-line-green"
                    />
                    {choice.choice_text}
                  </label>
                ))}
              </div>
            </fieldset>
          ))}
          <button
            onClick={() => void submit()}
            disabled={!allAnswered || loading}
            className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
          >
            {loading ? '採点中…' : '解答を送信する'}
          </button>
        </div>
      )}
    </div>
  )
}
