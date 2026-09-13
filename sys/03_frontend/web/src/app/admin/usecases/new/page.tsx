'use client'

import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import UsecaseForm, { type UsecaseFormValues } from '@/components/admin/UsecaseForm'
import type { Usecase } from '@/lib/types'

export default function NewUsecasePage() {
  const router = useRouter()

  return (
    <div>
      <Link href="/admin/usecases" className="text-sm font-medium text-line-green hover:underline">
        ← 導入事例一覧へ戻る
      </Link>
      <h1 className="mt-4 text-2xl font-bold text-gray-900">新規導入事例</h1>

      <div className="mt-6">
        <UsecaseForm
          submitLabel="作成する"
          onSubmit={async (values) => {
            const usecase = await api.post<Usecase>('/api/admin/usecases', values)
            router.push(`/admin/usecases/${usecase.id}/edit`)
          }}
        />
      </div>
    </div>
  )
}
