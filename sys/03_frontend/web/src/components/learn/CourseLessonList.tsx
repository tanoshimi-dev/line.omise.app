'use client'

// Lesson list for a course-top page, with completion badges for logged-in
// users (dev-plan-09-frontend-learn 9.2). Takes the (public, SSR-fetched)
// lesson list as props and only adds a client-side fetch for the per-user
// completion overlay — unauthenticated visitors see the plain list with no
// extra request.

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { CourseProgress } from '@/lib/types'

interface LessonSummary {
  slug: string
  title: string
}

export default function CourseLessonList({ courseSlug, lessons }: { courseSlug: string; lessons: LessonSummary[] }) {
  const { user } = useAuth()
  const [progress, setProgress] = useState<CourseProgress | null>(null)

  useEffect(() => {
    if (!user) return
    let cancelled = false
    api
      .get<CourseProgress>(`/api/courses/${courseSlug}/progress`)
      .then((data) => {
        if (!cancelled) setProgress(data)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [user, courseSlug])

  return (
    <div>
      {progress && (
        <p className="mb-4 text-sm text-gray-500">
          進捗: {progress.completed_count} / {progress.total_count} レッスン完了
        </p>
      )}
      <ol className="space-y-3">
        {lessons.map((lesson, index) => {
          const completed = progress?.lessons.find((l) => l.slug === lesson.slug)?.completed ?? false
          return (
            <li key={lesson.slug}>
              <Link
                href={`/learn/${courseSlug}/${lesson.slug}`}
                className="flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 transition-shadow hover:shadow-md"
              >
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-line-green/10 text-sm font-bold text-line-green">
                  {index + 1}
                </span>
                <span className="flex-1 font-medium text-gray-900">{lesson.title}</span>
                {completed && (
                  <span className="flex items-center gap-1 text-xs font-semibold text-line-green">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
                    </svg>
                    完了
                  </span>
                )}
              </Link>
            </li>
          )
        })}
      </ol>
    </div>
  )
}
