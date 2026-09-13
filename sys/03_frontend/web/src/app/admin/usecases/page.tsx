'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import type { Usecase } from '@/lib/types'

export default function AdminUsecasesPage() {
  const [usecases, setUsecases] = useState<Usecase[] | null>(null)

  useEffect(() => {
    void api.get<{ usecases: Usecase[] }>('/api/admin/usecases').then((data) => setUsecases(data.usecases))
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">導入事例</h1>
        <Link href="/admin/usecases/new" className="rounded-full bg-line-green px-4 py-2 text-sm font-semibold text-white hover:bg-line-green-dark">
          + 新規事例
        </Link>
      </div>

      {usecases === null ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : usecases.length === 0 ? (
        <p className="mt-8 text-gray-500">事例はまだありません。</p>
      ) : (
        <table className="mt-6 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">店舗名</th>
              <th className="py-2 font-medium">タイトル</th>
              <th className="py-2 font-medium">状態</th>
              <th className="py-2 font-medium" />
            </tr>
          </thead>
          <tbody>
            {usecases.map((usecase) => (
              <tr key={usecase.id} className="border-b border-gray-100">
                <td className="py-3 text-gray-700">{usecase.client_name}</td>
                <td className="py-3">
                  <Link href={`/admin/usecases/${usecase.id}/edit`} className="font-medium text-line-green hover:underline">
                    {usecase.title}
                  </Link>
                </td>
                <td className="py-3">
                  <span className={usecase.status === 'published' ? 'text-line-green' : 'text-gray-400'}>
                    {usecase.status === 'published' ? '公開' : '下書き'}
                  </span>
                </td>
                <td className="py-3 text-right">
                  <DeleteButton
                    path={`/api/admin/usecases/${usecase.id}`}
                    confirmMessage={`「${usecase.title}」を削除しますか？`}
                    onDeleted={() => setUsecases((prev) => prev?.filter((u) => u.id !== usecase.id) ?? null)}
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
