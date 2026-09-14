'use client'

// Per-question review of one past exam-mode submission
// (dev-plan-2-5-frontend-mypage 2-5.2), fetched from
// GET /api/me/quizzes/:slug/attempts/:attemptId.

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import type { QuizAttemptDetail } from '@/lib/types'

export default function AttemptDetailModal({ slug, attemptId, onClose }: { slug: string; attemptId: string; onClose: () => void }) {
  const [detail, setDetail] = useState<QuizAttemptDetail | null>(null)

  useEffect(() => {
    let cancelled = false
    api.get<QuizAttemptDetail>(`/api/me/quizzes/${slug}/attempts/${attemptId}`).then((data) => {
      if (!cancelled) setDetail(data)
    })
    return () => {
      cancelled = true
    }
  }, [slug, attemptId])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-gray-900">受験詳細</h3>
          <button onClick={onClose} className="text-sm text-gray-400 hover:text-gray-600">
            閉じる
          </button>
        </div>

        {!detail ? (
          <p className="mt-4 text-sm text-gray-400">読み込み中…</p>
        ) : (
          <>
            <p className="mt-2 text-sm text-gray-600">
              {detail.score} / {detail.total_questions}問正解
              {detail.passed !== null && (
                <span className={`ml-2 font-semibold ${detail.passed ? 'text-line-green' : 'text-red-600'}`}>
                  （{detail.passed ? '合格' : '不合格'}）
                </span>
              )}
            </p>
            <div className="mt-4 space-y-3">
              {detail.questions.map((q, qi) => (
                <div
                  key={q.question_id}
                  className={`rounded-xl border p-3 text-sm ${q.is_correct ? 'border-line-green/30 bg-green-50' : 'border-red-200 bg-red-50'}`}
                >
                  <p className="font-medium text-gray-900">
                    {qi + 1}. {q.question_text}
                  </p>
                  <p className={`mt-1 font-semibold ${q.is_correct ? 'text-line-green' : 'text-red-600'}`}>{q.is_correct ? '正解' : '不正解'}</p>
                  <p className="mt-1 text-gray-600">{q.explanation}</p>
                  {q.reference_url && (
                    <a href={q.reference_url} target="_blank" rel="noreferrer" className="mt-1 inline-block text-xs text-line-green hover:underline">
                      参考リンク
                    </a>
                  )}
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
