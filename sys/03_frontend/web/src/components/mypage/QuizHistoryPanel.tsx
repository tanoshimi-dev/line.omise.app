'use client'

// Practice-mode answer history for one quiz, lazily fetched when a
// QuizProgressSection card is expanded (dev-plan-2-5-frontend-mypage 2-5.1).

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import type { QuizPracticeHistory, QuizPracticeHistoryEntry } from '@/lib/types'

export default function QuizHistoryPanel({ slug }: { slug: string }) {
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
    <ul className="mt-4 space-y-2 border-t border-gray-100 pt-4">
      {history.map((entry, i) => (
        <li key={`${entry.question_id}-${entry.answered_at}-${i}`} className="flex items-start justify-between gap-4 text-sm">
          <div>
            <p className="text-gray-700">{entry.question_text}</p>
            <p className="text-xs text-gray-400">{new Date(entry.answered_at).toLocaleString('ja-JP')}</p>
          </div>
          <span className={`shrink-0 font-semibold ${entry.is_correct ? 'text-line-green' : 'text-red-600'}`}>
            {entry.is_correct ? '正解' : '不正解'}
          </span>
        </li>
      ))}
    </ul>
  )
}
