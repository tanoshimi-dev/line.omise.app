'use client'

// Mode-selection step before running a quiz (dev-plan-quiz-mode-selection):
// the reader picks single mode (question-by-question, immediate feedback)
// or exam mode (answer everything, then a final score) fresh for every
// attempt — no admin-set default to steer the choice.

import { useState } from 'react'
import QuizQuestionCard from './QuizQuestionCard'
import ExamRunner from './ExamRunner'
import type { Quiz } from '@/lib/types'

type QuizMode = 'single' | 'exam'

export default function QuizModeSelect({ quiz }: { quiz: Quiz }) {
  const [mode, setMode] = useState<QuizMode | null>(null)

  if (mode === 'single') {
    return <QuizQuestionCard quiz={quiz} />
  }
  if (mode === 'exam') {
    return <ExamRunner quiz={quiz} />
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <button
        onClick={() => setMode('single')}
        className="rounded-2xl border border-gray-100 bg-white p-6 text-left shadow-sm transition-shadow hover:shadow-md"
      >
        <h2 className="font-bold text-gray-900">単発モード</h2>
        <p className="mt-2 text-sm text-gray-600">1問ずつ解答して、その場で正誤と解説を確認できます。</p>
      </button>
      <button
        onClick={() => setMode('exam')}
        className="rounded-2xl border border-gray-100 bg-white p-6 text-left shadow-sm transition-shadow hover:shadow-md"
      >
        <h2 className="font-bold text-gray-900">検定モード</h2>
        <p className="mt-2 text-sm text-gray-600">すべての設問に解答してから、まとめて最終スコアを確認できます。</p>
        {quiz.passing_score != null && <p className="mt-2 text-xs text-gray-400">合格点: {quiz.passing_score}%</p>}
      </button>
    </div>
  )
}
