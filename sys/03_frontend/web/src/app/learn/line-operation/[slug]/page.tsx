import type { Metadata } from 'next'
import { ArticleDetailView, articleDetailMetadata, getArticleOrNotFound } from '@/components/learn/ArticleDetailView'

interface PageProps {
  params: Promise<{ slug: string }>
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params
  return articleDetailMetadata('line-operation', slug)
}

export default async function LineOperationArticlePage({ params }: PageProps) {
  const { slug } = await params
  const article = await getArticleOrNotFound('line-operation', slug)
  return <ArticleDetailView category="line-operation" article={article} />
}
