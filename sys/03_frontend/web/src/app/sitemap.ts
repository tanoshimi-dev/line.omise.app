import type { MetadataRoute } from 'next'
import { serverApi } from '@/lib/serverApi'
import type { Article, Usecase } from '@/lib/types'

// Replaces the static public/sitemap.xml from dev-plan-08-frontend-lp — now
// includes published /learn/ and /usecase/ content dynamically
// (dev-plan-09-frontend-learn 9.5, dev-plan-10-frontend-usecase 10.3).
// Fetching from the backend makes this route dynamic (rendered per request,
// not baked in at build time), so new content appears without a rebuild;
// each fetch is best-effort so a backend hiccup degrades to the static
// entries rather than failing the whole sitemap.

const BASE_URL = 'https://line.omise.app'

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const entries: MetadataRoute.Sitemap = [
    { url: BASE_URL, changeFrequency: 'monthly', priority: 1 },
    { url: `${BASE_URL}/learn`, changeFrequency: 'weekly', priority: 0.8 },
    { url: `${BASE_URL}/learn/line-operation`, changeFrequency: 'weekly', priority: 0.7 },
    { url: `${BASE_URL}/learn/ai`, changeFrequency: 'weekly', priority: 0.7 },
    { url: `${BASE_URL}/learn/line-yahoo-certification`, changeFrequency: 'weekly', priority: 0.7 },
    { url: `${BASE_URL}/usecase`, changeFrequency: 'weekly', priority: 0.7 },
  ]

  for (const category of ['line-operation', 'ai'] as const) {
    try {
      const { articles } = await serverApi.get<{ articles: Article[] }>(`/api/articles?category=${category}`)
      for (const article of articles) {
        entries.push({
          url: `${BASE_URL}/learn/${category}/${article.slug}`,
          lastModified: article.published_at ?? undefined,
          changeFrequency: 'monthly',
          priority: 0.6,
        })
      }
    } catch {
      // best-effort
    }
  }

  try {
    const { usecases } = await serverApi.get<{ usecases: Usecase[] }>('/api/usecases')
    for (const usecase of usecases) {
      entries.push({
        url: `${BASE_URL}/usecase/${usecase.slug}`,
        lastModified: usecase.published_at ?? undefined,
        changeFrequency: 'monthly',
        priority: 0.6,
      })
    }
  } catch {
    // best-effort
  }

  return entries
}
