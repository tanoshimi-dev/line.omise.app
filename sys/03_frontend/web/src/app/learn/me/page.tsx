'use client'

// My-page progress summary (dev-plan-09-frontend-learn 9.4). Entirely
// user-specific — no public/SEO-relevant content — so this is a plain
// Client Component page rather than a Server Component shell.

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import LoginPrompt from '@/components/learn/LoginPrompt'
import QuizProgressSection from '@/components/mypage/QuizProgressSection'
import type { MyProgress } from '@/lib/types'

export default function MyProgressPage() {
  const { user, loading: authLoading } = useAuth()
  const [progress, setProgress] = useState<MyProgress | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!user) {
      setLoading(false)
      return
    }
    let cancelled = false
    api
      .get<MyProgress>('/api/me/progress')
      .then((data) => {
        if (!cancelled) setProgress(data)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [user])

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">マイページ</h1>

      {authLoading || loading ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : !user ? (
        <div className="mt-8">
          <LoginPrompt message="ログインすると、学習の進捗を確認できます。" />
        </div>
      ) : !progress || progress.courses.length === 0 ? (
        <p className="mt-8 text-gray-500">受講中の講座はまだありません。</p>
      ) : (
        <ul className="mt-8 space-y-4">
          {progress.courses.map((course) => {
            const percent = course.total_count === 0 ? 0 : Math.round((course.completed_count / course.total_count) * 100)
            return (
              <li key={course.course_id}>
                <Link
                  href={`/learn/${course.slug}`}
                  className="block rounded-2xl border border-gray-100 bg-white p-6 shadow-sm transition-shadow hover:shadow-md"
                >
                  <div className="flex items-center justify-between">
                    <h2 className="font-bold text-gray-900">{course.title}</h2>
                    <span className="text-sm text-gray-500">
                      {course.completed_count} / {course.total_count}
                    </span>
                  </div>
                  <div className="mt-3 h-2 overflow-hidden rounded-full bg-gray-100">
                    <div className="h-full rounded-full bg-line-green" style={{ width: `${percent}%` }} />
                  </div>
                </Link>
              </li>
            )
          })}
        </ul>
      )}

      {!authLoading && user && <QuizProgressSection />}
    </div>
  )
}
