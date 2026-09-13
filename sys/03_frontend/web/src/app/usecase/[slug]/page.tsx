import Link from 'next/link'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { serverApi } from '@/lib/serverApi'
import { excerpt } from '@/lib/markdown'
import Markdown from '@/components/learn/Markdown'
import { demoApps } from '@/data/demoApps'
import type { Usecase } from '@/lib/types'

interface PageProps {
  params: Promise<{ slug: string }>
}

async function getUsecaseOrNotFound(slug: string): Promise<Usecase> {
  const usecase = await serverApi.getOrNull<Usecase>(`/api/usecases/${slug}`)
  if (!usecase) {
    notFound()
  }
  return usecase
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params
  const usecase = await getUsecaseOrNotFound(slug)
  return {
    title: `${usecase.client_name} 導入事例`,
    description: excerpt(usecase.body, 140),
  }
}

export default async function UsecaseDetailPage({ params }: PageProps) {
  const { slug } = await params
  const usecase = await getUsecaseOrNotFound(slug)
  // dev-plan-08-frontend-lp gave each demo app a live externalUrl — reuse it
  // here rather than duplicating the URL (dev-plan-10 10.2).
  const relatedApp = usecase.related_demo_app ? demoApps.find((app) => app.id === usecase.related_demo_app) : undefined

  return (
    <article className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <Link href="/usecase" className="text-sm font-medium text-line-green hover:underline">
        ← 導入事例一覧へ戻る
      </Link>
      <p className="mt-4 text-sm font-medium text-line-green">{usecase.client_name}</p>
      <h1 className="mt-1 text-3xl font-bold text-gray-900">{usecase.title}</h1>

      {usecase.thumbnail_url && (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={usecase.thumbnail_url} alt={usecase.client_name} className="mt-6 w-full rounded-2xl object-cover" />
      )}

      <div className="mt-8">
        <Markdown>{usecase.body}</Markdown>
      </div>

      {relatedApp && (
        <a
          href={relatedApp.externalUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-10 flex items-center justify-center gap-2 rounded-full border-2 border-line-green px-6 py-3 text-sm font-semibold text-line-green transition-colors hover:bg-line-green hover:text-white"
        >
          {relatedApp.name}のデモサイトを見る
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
          </svg>
        </a>
      )}
    </article>
  )
}
