'use client'

// Mypage "クイズ／検定" section (dev-plan-2-5-frontend-mypage 2-5.3): summarizes
// every published quiz/exam via GET /api/me/quizzes/progress, expanding into
// per-question history (practice) or attempt history (exam) on demand.
// Rendered only while the caller is logged in — see src/app/learn/me/page.tsx
// for the gating (2-5.4).

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import type { MyQuizzesProgress, QuizExamProgress, QuizPracticeProgress, QuizProgressSummary } from '@/lib/types'
import QuizHistoryPanel from './QuizHistoryPanel'
import ExamAttemptsPanel from './ExamAttemptsPanel'

export default function QuizProgressSection() {
  const [quizzes, setQuizzes] = useState<QuizProgressSummary[] | null>(null)

  useEffect(() => {
    let cancelled = false
    api
      .get<MyQuizzesProgress>('/api/me/quizzes/progress')
      .then((data) => {
        if (!cancelled) setQuizzes(data.quizzes)
      })
      .catch(() => {
        if (!cancelled) setQuizzes([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="mt-10 border-t border-gray-100 pt-8">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">クイズ・検定</h2>
        <Link href="/learn/quiz" className="text-sm font-medium text-line-green hover:underline">
          一覧を見る
        </Link>
      </div>

      {quizzes === null ? (
        <p className="mt-4 text-gray-400">読み込み中…</p>
      ) : quizzes.length === 0 ? (
        <p className="mt-4 text-gray-500">クイズ・検定はまだありません。</p>
      ) : (
        <div className="mt-4 space-y-4">
          {quizzes.map((quiz) =>
            quiz.mode === 'practice' ? <PracticeProgressCard key={quiz.quiz_id} quiz={quiz} /> : <ExamProgressCard key={quiz.quiz_id} quiz={quiz} />,
          )}
        </div>
      )}
    </div>
  )
}

function PracticeProgressCard({ quiz }: { quiz: QuizPracticeProgress }) {
  const [expanded, setExpanded] = useState(false)
  const accuracy = quiz.answered_count === 0 ? null : Math.round((quiz.correct_count / quiz.answered_count) * 100)

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <Link href={`/learn/quiz/${quiz.slug}`} className="font-bold text-gray-900 hover:text-line-green">
            {quiz.title}
          </Link>
          <p className="mt-1 text-sm text-gray-500">
            解答済み: {quiz.answered_count} / {quiz.total_questions}問
            {accuracy !== null && `（正答率 ${accuracy}%）`}
          </p>
        </div>
        {quiz.answered_count > 0 && (
          <button onClick={() => setExpanded((e) => !e)} className="shrink-0 text-sm font-medium text-line-green hover:underline">
            {expanded ? '履歴を閉じる' : '解答履歴を見る'}
          </button>
        )}
      </div>
      {expanded && <QuizHistoryPanel slug={quiz.slug} />}
    </div>
  )
}

function ExamProgressCard({ quiz }: { quiz: QuizExamProgress }) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <Link href={`/learn/quiz/${quiz.slug}`} className="font-bold text-gray-900 hover:text-line-green">
            {quiz.title}
          </Link>
          <p className="mt-1 text-sm text-gray-500">
            受験回数: {quiz.attempt_count}回
            {quiz.best_score !== null && `・最高 ${quiz.best_score}/${quiz.total_questions}問`}
            {quiz.latest_score !== null && `・直近 ${quiz.latest_score}/${quiz.total_questions}問`}
          </p>
          {quiz.latest_passed !== null && (
            <p className={`mt-1 text-sm font-semibold ${quiz.latest_passed ? 'text-line-green' : 'text-red-600'}`}>
              直近の結果: {quiz.latest_passed ? '合格' : '不合格'}
            </p>
          )}
        </div>
        {quiz.attempt_count > 0 && (
          <button onClick={() => setExpanded((e) => !e)} className="shrink-0 text-sm font-medium text-line-green hover:underline">
            {expanded ? '受験履歴を閉じる' : '受験履歴を見る'}
          </button>
        )}
      </div>
      {expanded && <ExamAttemptsPanel slug={quiz.slug} />}
    </div>
  )
}
