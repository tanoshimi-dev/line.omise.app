import Link from 'next/link'
import type { Metadata } from 'next'
import { serverApi } from '@/lib/serverApi'
import type { Usecase } from '@/lib/types'

export const metadata: Metadata = {
  title: '導入事例',
  description: 'LINEミニアプリを導入いただいた店舗様のインタビュー事例です。',
}

interface UsecasesResponse {
  usecases: Usecase[]
}

export default async function UsecaseListPage() {
  const { usecases } = await serverApi.get<UsecasesResponse>('/api/usecases')

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">導入事例</h1>
      <p className="mt-4 text-gray-600">LINEミニアプリをご導入いただいた店舗様の事例をご紹介します。</p>

      {usecases.length === 0 ? (
        <p className="mt-12 text-gray-500">事例はまだありません。</p>
      ) : (
        <ul className="mt-10 grid grid-cols-1 gap-6 sm:grid-cols-2">
          {usecases.map((usecase) => (
            <li key={usecase.id}>
              <Link
                href={`/usecase/${usecase.slug}`}
                className="block h-full overflow-hidden rounded-2xl border border-gray-100 bg-white shadow-sm transition-shadow hover:shadow-md"
              >
                {usecase.thumbnail_url && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={usecase.thumbnail_url} alt={usecase.client_name} className="h-40 w-full object-cover" />
                )}
                <div className="p-6">
                  <p className="text-sm font-medium text-line-green">{usecase.client_name}</p>
                  <h2 className="mt-1 text-lg font-bold text-gray-900">{usecase.title}</h2>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
