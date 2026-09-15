'use client'

// Exam-mode attempt history for one quiz, lazily fetched when a
// QuizProgressSection card is expanded (dev-plan-2-5-frontend-mypage 2-5.2).
// Clicking a row opens AttemptDetailModal for the per-question review. Each
// attempt and the whole list can be deleted by their owner
// (dev-plan-quiz-history-delete) — onChanged lets the parent card refresh
// its progress summary afterward.

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import type { QuizAttemptsList, QuizAttemptSummary } from '@/lib/types'
import AttemptDetailModal from './AttemptDetailModal'

export default function ExamAttemptsPanel({ slug, onChanged }: { slug: string; onChanged?: () => void }) {
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
      <div className="mt-4 border-t border-gray-100 pt-4">
        <div className="flex justify-end">
          <DeleteButton
            path={`/api/me/quizzes/${slug}/attempts`}
            confirmMessage="このクイズの受験履歴をすべて削除しますか？"
            onDeleted={() => {
              setAttempts([])
              onChanged?.()
            }}
          />
        </div>
        <ul className="mt-2 space-y-1">
          {attempts.map((attempt) => (
            <li key={attempt.id} className="flex items-center gap-2">
              <button
                onClick={() => setSelectedAttemptId(attempt.id)}
                className="flex flex-1 items-center justify-between gap-4 rounded-lg px-2 py-1.5 text-left text-sm hover:bg-gray-50"
              >
                <span className="text-gray-500">{new Date(attempt.submitted_at).toLocaleString('ja-JP')}</span>
                <span className="font-medium text-gray-900">
                  {attempt.score} / {attempt.total_questions}問
                </span>
                {attempt.passed !== null && (
                  <span className={`font-semibold ${attempt.passed ? 'text-line-green' : 'text-red-600'}`}>
                    {attempt.passed ? '合格' : '不合格'}
                  </span>
                )}
              </button>
              <DeleteButton
                path={`/api/me/quizzes/${slug}/attempts/${attempt.id}`}
                confirmMessage="この受験履歴を削除しますか？"
                onDeleted={() => {
                  setAttempts((prev) => (prev ? prev.filter((a) => a.id !== attempt.id) : prev))
                  onChanged?.()
                }}
              />
            </li>
          ))}
        </ul>
      </div>
      {selectedAttemptId && <AttemptDetailModal slug={slug} attemptId={selectedAttemptId} onClose={() => setSelectedAttemptId(null)} />}
    </>
  )
}
