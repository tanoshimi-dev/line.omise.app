import Link from 'next/link'
import { notFound } from 'next/navigation'
import type { Metadata } from 'next'
import { serverApi } from '@/lib/serverApi'
import { excerpt } from '@/lib/markdown'
import Markdown from './Markdown'
import { CATEGORY_LABELS } from '@/lib/articleCategories'
import type { Article, ArticleCategory } from '@/lib/types'

// Fetched separately from the render so generateMetadata and the page body
// can each call this — Next.js dedupes identical fetch() calls within one
// request, so this isn't a double network round trip.
export async function getArticleOrNotFound(category: ArticleCategory, slug: string): Promise<Article> {
  const article = await serverApi.getOrNull<Article>(`/api/articles/${category}/${slug}`)
  if (!article) {
    notFound()
  }
  return article
}

export async function articleDetailMetadata(category: ArticleCategory, slug: string): Promise<Metadata> {
  const article = await serverApi.getOrNull<Article>(`/api/articles/${category}/${slug}`)
  if (!article) {
    return {}
  }
  return {
    title: article.title,
    description: excerpt(article.body, 140),
  }
}

// Shared between /learn/line-operation/[slug] and /learn/ai/[slug]
// (dev-plan-09-frontend-learn 9.3).
export function ArticleDetailView({ category, article }: { category: ArticleCategory; article: Article }) {
  return (
    <article className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <Link href={`/learn/${category}`} className="text-sm font-medium text-line-green hover:underline">
        ← {CATEGORY_LABELS[category]}一覧へ戻る
      </Link>
      <h1 className="mt-4 text-3xl font-bold text-gray-900">{article.title}</h1>
      {article.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-2">
          {article.tags.map((t) => (
            <span key={t.id} className="rounded-full bg-line-green-light px-2.5 py-0.5 text-xs font-medium text-line-green">
              {t.name}
            </span>
          ))}
        </div>
      )}
      <div className="mt-8">
        <Markdown>{article.body}</Markdown>
      </div>
    </article>
  )
}
