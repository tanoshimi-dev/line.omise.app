import Link from 'next/link'
import type { Metadata } from 'next'
import { serverApi } from '@/lib/serverApi'
import { excerpt } from '@/lib/markdown'
import type { Article, ArticleCategory } from '@/lib/types'

export const CATEGORY_LABELS: Record<ArticleCategory, string> = {
  'line-operation': 'LINE運用・設定',
  ai: '生成AI活用事例',
}

interface ArticlesResponse {
  articles: Article[]
}

export async function articleListMetadata(category: ArticleCategory): Promise<Metadata> {
  const label = CATEGORY_LABELS[category]
  return {
    title: label,
    description: `${label}に関する記事一覧`,
  }
}

// Shared between /learn/line-operation and /learn/ai (dev-plan-09-frontend-learn
// 9.3 — same shape, category is the only difference).
//
// Tag filter uses the ?tag= query param per README's decided approach. The
// tag menu is derived from the *unfiltered* category list so every tag stays
// visible even while one is selected — fetching only the filtered list would
// otherwise hide every other tag once you picked one.
export default async function ArticleListView({ category, tag }: { category: ArticleCategory; tag?: string }) {
  const full = await serverApi.get<ArticlesResponse>(`/api/articles?category=${category}`)
  const filtered = tag
    ? await serverApi.get<ArticlesResponse>(`/api/articles?category=${category}&tag=${encodeURIComponent(tag)}`)
    : full

  const allTags = new Map<string, string>()
  for (const article of full.articles) {
    for (const t of article.tags) {
      allTags.set(t.slug, t.name)
    }
  }

  const basePath = `/learn/${category}`

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">{CATEGORY_LABELS[category]}</h1>

      {allTags.size > 0 && (
        <div className="mt-6 flex flex-wrap gap-2">
          <Link
            href={basePath}
            className={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
              !tag ? 'bg-line-green text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            すべて
          </Link>
          {[...allTags].map(([slug, name]) => (
            <Link
              key={slug}
              href={`${basePath}?tag=${slug}`}
              className={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
                tag === slug ? 'bg-line-green text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {name}
            </Link>
          ))}
        </div>
      )}

      {filtered.articles.length === 0 ? (
        <p className="mt-12 text-gray-500">記事はまだありません。</p>
      ) : (
        <ul className="mt-10 grid grid-cols-1 gap-6 sm:grid-cols-2">
          {filtered.articles.map((article) => (
            <li key={article.id}>
              <Link
                href={`${basePath}/${article.slug}`}
                className="block h-full rounded-2xl border border-gray-100 bg-white p-6 shadow-sm transition-shadow hover:shadow-md"
              >
                <h2 className="text-lg font-bold text-gray-900">{article.title}</h2>
                <p className="mt-2 text-sm text-gray-600">{excerpt(article.body)}</p>
                {article.tags.length > 0 && (
                  <div className="mt-4 flex flex-wrap gap-2">
                    {article.tags.map((t) => (
                      <span key={t.id} className="rounded-full bg-line-green-light px-2.5 py-0.5 text-xs font-medium text-line-green">
                        {t.name}
                      </span>
                    ))}
                  </div>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
