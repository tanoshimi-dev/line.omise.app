'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import type { AdminQuiz } from '@/lib/types'

export default function AdminQuizzesPage() {
  const [quizzes, setQuizzes] = useState<AdminQuiz[] | null>(null)

  useEffect(() => {
    void api.get<{ quizzes: AdminQuiz[] }>('/api/admin/quizzes').then((data) => setQuizzes(data.quizzes))
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">LINEヤフー認定資格</h1>
        <Link href="/admin/quizzes/new" className="rounded-full bg-line-green px-4 py-2 text-sm font-semibold text-white hover:bg-line-green-dark">
          + 新規作成
        </Link>
      </div>

      {quizzes === null ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : quizzes.length === 0 ? (
        <p className="mt-8 text-gray-500">クイズ・検定はまだありません。</p>
      ) : (
        <table className="mt-6 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">タイトル</th>
              <th className="py-2 font-medium">スラッグ</th>
              <th className="py-2 font-medium">合格点</th>
              <th className="py-2 font-medium">状態</th>
              <th className="py-2 font-medium" />
            </tr>
          </thead>
          <tbody>
            {quizzes.map((quiz) => (
              <tr key={quiz.id} className="border-b border-gray-100">
                <td className="py-3">
                  <Link href={`/admin/quizzes/${quiz.id}/edit`} className="font-medium text-line-green hover:underline">
                    {quiz.title}
                  </Link>
                </td>
                <td className="py-3 text-gray-500">{quiz.slug}</td>
                <td className="py-3 text-gray-500">{quiz.passing_score != null ? `${quiz.passing_score}%` : '—'}</td>
                <td className="py-3">
                  <span className={quiz.published ? 'text-line-green' : 'text-gray-400'}>{quiz.published ? '公開' : '下書き'}</span>
                </td>
                <td className="py-3 text-right">
                  <DeleteButton
                    path={`/api/admin/quizzes/${quiz.id}`}
                    confirmMessage={`「${quiz.title}」を削除しますか？設問も削除されます。`}
                    onDeleted={() => setQuizzes((prev) => prev?.filter((q) => q.id !== quiz.id) ?? null)}
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
