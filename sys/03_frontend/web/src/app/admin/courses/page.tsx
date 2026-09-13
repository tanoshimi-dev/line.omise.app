'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import type { Course } from '@/lib/types'

export default function AdminCoursesPage() {
  const [courses, setCourses] = useState<Course[] | null>(null)

  useEffect(() => {
    void api.get<{ courses: Course[] }>('/api/admin/courses').then((data) => setCourses(data.courses))
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">講座</h1>
        <Link href="/admin/courses/new" className="rounded-full bg-line-green px-4 py-2 text-sm font-semibold text-white hover:bg-line-green-dark">
          + 新規講座
        </Link>
      </div>

      {courses === null ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : courses.length === 0 ? (
        <p className="mt-8 text-gray-500">講座はまだありません。</p>
      ) : (
        <table className="mt-6 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">タイトル</th>
              <th className="py-2 font-medium">スラッグ</th>
              <th className="py-2 font-medium">状態</th>
              <th className="py-2 font-medium" />
            </tr>
          </thead>
          <tbody>
            {courses.map((course) => (
              <tr key={course.id} className="border-b border-gray-100">
                <td className="py-3">
                  <Link href={`/admin/courses/${course.id}/edit`} className="font-medium text-line-green hover:underline">
                    {course.title}
                  </Link>
                </td>
                <td className="py-3 text-gray-500">{course.slug}</td>
                <td className="py-3">
                  <span className={course.status === 'published' ? 'text-line-green' : 'text-gray-400'}>
                    {course.status === 'published' ? '公開' : '下書き'}
                  </span>
                </td>
                <td className="py-3 text-right">
                  <DeleteButton
                    path={`/api/admin/courses/${course.id}`}
                    confirmMessage={`「${course.title}」を削除しますか？`}
                    onDeleted={() => setCourses((prev) => prev?.filter((c) => c.id !== course.id) ?? null)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
