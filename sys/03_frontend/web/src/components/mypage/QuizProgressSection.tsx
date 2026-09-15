'use client'

// Mypage "クイズ／検定" section (dev-plan-2-5-frontend-mypage 2-5.3): summarizes
// every published quiz via GET /api/me/quizzes/progress, expanding into
// per-question history and/or attempt history on demand. A single quiz may
// have both, since the reader picks the mode per attempt
// (dev-plan-quiz-mode-selection). Rendered only while the caller is logged
// in — see src/app/learn/me/page.tsx for the gating (2-5.4).

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import type { MyQuizzesProgress, QuizProgressSummary } from '@/lib/types'
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
          {quizzes.map((quiz) => (
            <QuizProgressCard key={quiz.quiz_id} quiz={quiz} />
          ))}
        </div>
      )}
    </div>
  )
}

function QuizProgressCard({ quiz: initialQuiz }: { quiz: QuizProgressSummary }) {
  const [quiz, setQuiz] = useState(initialQuiz)
  const [historyExpanded, setHistoryExpanded] = useState(false)
  const [attemptsExpanded, setAttemptsExpanded] = useState(false)
  const accuracy = quiz.answered_count === 0 ? null : Math.round((quiz.correct_count / quiz.answered_count) * 100)

  const refreshProgress = () => {
    void api.get<QuizProgressSummary>(`/api/me/quizzes/${quiz.slug}/progress`).then(setQuiz)
  }

  return (
    <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
      <Link href={`/learn/quiz/${quiz.slug}`} className="font-bold text-gray-900 hover:text-line-green">
        {quiz.title}
      </Link>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-4">
        <p className="text-sm text-gray-500">
          単発モード解答済み: {quiz.answered_count} / {quiz.total_questions}問
          {accuracy !== null && `（正答率 ${accuracy}%）`}
        </p>
        {(quiz.answered_count > 0 || historyExpanded) && (
          <button onClick={() => setHistoryExpanded((e) => !e)} className="shrink-0 text-sm font-medium text-line-green hover:underline">
            {historyExpanded ? '解答履歴を閉じる' : '解答履歴を見る'}
          </button>
        )}
      </div>
      {historyExpanded && <QuizHistoryPanel slug={quiz.slug} onChanged={refreshProgress} />}

      <div className="mt-3 flex flex-wrap items-center justify-between gap-4 border-t border-gray-100 pt-3">
        <div>
          <p className="text-sm text-gray-500">
            検定モード受験回数: {quiz.attempt_count}回
            {quiz.best_score !== null && `・最高 ${quiz.best_score}/${quiz.total_questions}問`}
            {quiz.latest_score !== null && `・直近 ${quiz.latest_score}/${quiz.total_questions}問`}
          </p>
          {quiz.latest_passed !== null && (
            <p className={`mt-1 text-sm font-semibold ${quiz.latest_passed ? 'text-line-green' : 'text-red-600'}`}>
              直近の結果: {quiz.latest_passed ? '合格' : '不合格'}
            </p>
          )}
        </div>
        {(quiz.attempt_count > 0 || attemptsExpanded) && (
          <button onClick={() => setAttemptsExpanded((e) => !e)} className="shrink-0 text-sm font-medium text-line-green hover:underline">
            {attemptsExpanded ? '受験履歴を閉じる' : '受験履歴を見る'}
          </button>
        )}
      </div>
      {attemptsExpanded && <ExamAttemptsPanel slug={quiz.slug} onChanged={refreshProgress} />}
    </div>
  )
}
