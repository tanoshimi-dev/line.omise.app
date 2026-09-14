'use client'

// Exam-mode attempt history for one quiz, lazily fetched when a
// QuizProgressSection card is expanded (dev-plan-2-5-frontend-mypage 2-5.2).
// Clicking a row opens AttemptDetailModal for the per-question review.

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import type { QuizAttemptsList, QuizAttemptSummary } from '@/lib/types'
import AttemptDetailModal from './AttemptDetailModal'

export default function ExamAttemptsPanel({ slug }: { slug: string }) {
  const [attempts, setAttempts] = useState<QuizAttemptSummary[] | null>(null)
  const [selectedAttemptId, setSelectedAttemptId] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    api
      .get<QuizAttemptsList>(`/api/me/quizzes/${slug}/attempts`)
      .then((data) => {
        if (!cancelled) setAttempts(data.attempts)
      })
      .catch(() => {
        if (!cancelled) setAttempts([])
      })
    return () => {
      cancelled = true
    }
  }, [slug])

  if (attempts === null) {
    return <p className="mt-4 text-sm text-gray-400">読み込み中…</p>
  }

  if (attempts.length === 0) {
    return <p className="mt-4 text-sm text-gray-500">受験履歴がありません。</p>
  }

  return (
    <>
      <ul className="mt-4 space-y-1 border-t border-gray-100 pt-4">
        {attempts.map((attempt) => (
          <li key={attempt.id}>
            <button
              onClick={() => setSelectedAttemptId(attempt.id)}
              className="flex w-full items-center justify-between gap-4 rounded-lg px-2 py-1.5 text-left text-sm hover:bg-gray-50"
            >
              <span className="text-gray-500">{new Date(attempt.submitted_at).toLocaleString('ja-JP')}</span>
              <span className="font-medium text-gray-900">
                {attempt.score} / {attempt.total_questions}問
              </span>
              {attempt.passed !== null && (
                <span className={`font-semibold ${attempt.passed ? 'text-line-green' : 'text-red-600'}`}>{attempt.passed ? '合格' : '不合格'}</span>
              )}
            </button>
          </li>
        ))}
      </ul>
      {selectedAttemptId && <AttemptDetailModal slug={slug} attemptId={selectedAttemptId} onClose={() => setSelectedAttemptId(null)} />}
    </>
  )
}
