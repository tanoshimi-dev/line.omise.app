'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import UsecaseForm, { type UsecaseFormValues } from '@/components/admin/UsecaseForm'
import DeleteButton from '@/components/admin/DeleteButton'
import type { Usecase } from '@/lib/types'

export default function EditUsecasePage() {
  const { id } = useParams<{ id: string }>()
  const [usecase, setUsecase] = useState<Usecase | null>(null)

  useEffect(() => {
    void api.get<Usecase>(`/api/admin/usecases/${id}`).then(setUsecase)
  }, [id])

  if (!usecase) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  const initialValues: UsecaseFormValues = {
    slug: usecase.slug,
    client_name: usecase.client_name,
    title: usecase.title,
    body: usecase.body,
    status: usecase.status,
    thumbnail_url: usecase.thumbnail_url,
    related_demo_app: usecase.related_demo_app,
  }

  return (
    <div>
      <Link href="/admin/usecases" className="text-sm font-medium text-line-green hover:underline">
        ← 導入事例一覧へ戻る
      </Link>
      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{usecase.title}</h1>
        <DeleteButton path={`/api/admin/usecases/${id}`} confirmMessage="この事例を削除しますか？" redirectTo="/admin/usecases" />
      </div>

      <div className="mt-6">
        <UsecaseForm
          initialValues={initialValues}
          submitLabel="保存する"
          onSubmit={async (values) => {
            const updated = await api.put<Usecase>(`/api/admin/usecases/${id}`, values)
            setUsecase(updated)
          }}
        />
      </div>
    </div>
  )
}
