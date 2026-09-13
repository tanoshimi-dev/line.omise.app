import type { Metadata } from 'next'
import { ArticleDetailView, articleDetailMetadata, getArticleOrNotFound } from '@/components/learn/ArticleDetailView'

interface PageProps {
  params: Promise<{ slug: string }>
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params
  return articleDetailMetadata('ai', slug)
}

export default async function AiArticlePage({ params }: PageProps) {
  const { slug } = await params
  const article = await getArticleOrNotFound('ai', slug)
  return <ArticleDetailView category="ai" article={article} />
}
