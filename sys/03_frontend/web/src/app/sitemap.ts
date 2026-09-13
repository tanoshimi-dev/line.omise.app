import type { MetadataRoute } from 'next'
import { serverApi } from '@/lib/serverApi'
import type { Article, Course } from '@/lib/types'

// Replaces the static public/sitemap.xml from dev-plan-08-frontend-lp — now
// includes published /learn/ content dynamically (dev-plan-09-frontend-learn
// 9.5). Fetching from the backend makes this route dynamic (rendered per
// request, not baked in at build time), so new content appears without a
// rebuild; each fetch is best-effort so a backend hiccup degrades to the
// static entries rather than failing the whole sitemap.

const BASE_URL = 'https://line.omise.app'

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const entries: MetadataRoute.Sitemap = [
    { url: BASE_URL, changeFrequency: 'monthly', priority: 1 },
    { url: `${BASE_URL}/learn`, changeFrequency: 'weekly', priority: 0.8 },
    { url: `${BASE_URL}/learn/line-marketing`, changeFrequency: 'weekly', priority: 0.7 },
    { url: `${BASE_URL}/learn/line-operation`, changeFrequency: 'weekly', priority: 0.7 },
    { url: `${BASE_URL}/learn/ai`, changeFrequency: 'weekly', priority: 0.7 },
  ]

  try {
    const course = await serverApi.getOrNull<Course>('/api/courses/line-marketing')
    for (const lesson of course?.lessons ?? []) {
      entries.push({
        url: `${BASE_URL}/learn/line-marketing/${lesson.slug}`,
        changeFrequency: 'monthly',
        priority: 0.6,
      })
    }
  } catch {
    // best-effort
  }

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

  return entries
}
