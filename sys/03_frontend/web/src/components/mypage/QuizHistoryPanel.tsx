'use client'

// Single-mode answer history for one quiz, lazily fetched when a
// QuizProgressSection card is expanded (dev-plan-2-5-frontend-mypage 2-5.1).
// Each entry and the whole list can be deleted by their owner
// (dev-plan-quiz-history-delete) — onChanged lets the parent card refresh
// its progress summary afterward.

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import type { QuizPracticeHistory, QuizPracticeHistoryEntry } from '@/lib/types'

export default function QuizHistoryPanel({ slug, onChanged }: { slug: string; onChanged?: () => void }) {
  const [history, setHistory] = useState<QuizPracticeHistoryEntry[] | null>(null)

  useEffect(() => {
    let cancelled = false
    api
      .get<QuizPracticeHistory>(`/api/me/quizzes/${slug}/history`)
      .then((data) => {
        if (!cancelled) setHistory(data.history)
      })
      .catch(() => {
        if (!cancelled) setHistory([])
      })
    return () => {
      cancelled = true
    }
  }, [slug])

  if (history === null) {
    return <p className="mt-4 text-sm text-gray-400">読み込み中…</p>
  }

  if (history.length === 0) {
    return <p className="mt-4 text-sm text-gray-500">解答履歴がありません。</p>
  }

  return (
    <div className="mt-4 border-t border-gray-100 pt-4">
      <div className="flex justify-end">
        <DeleteButton
          path={`/api/me/quizzes/${slug}/history`}
          confirmMessage="このクイズの解答履歴をすべて削除しますか？"
          onDeleted={() => {
            setHistory([])
            onChanged?.()
          }}
        />
      </div>
      <ul className="mt-2 space-y-2">
        {history.map((entry) => (
          <li key={entry.id} className="flex items-start justify-between gap-4 text-sm">
            <div>
              <p className="text-gray-700">{entry.question_text}</p>
              <p className="text-xs text-gray-400">{new Date(entry.answered_at).toLocaleString('ja-JP')}</p>
            </div>
            <div className="flex shrink-0 items-center gap-3">
              <span className={`font-semibold ${entry.is_correct ? 'text-line-green' : 'text-red-600'}`}>
                {entry.is_correct ? '正解' : '不正解'}
              </span>
              <DeleteButton
                path={`/api/me/quizzes/${slug}/history/${entry.id}`}
                confirmMessage="この解答履歴を削除しますか？"
                onDeleted={() => {
                  setHistory((prev) => (prev ? prev.filter((e) => e.id !== entry.id) : prev))
                  onChanged?.()
                }}
              />
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
