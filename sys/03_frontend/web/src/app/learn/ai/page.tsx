import type { Metadata } from 'next'
import ArticleListView, { articleListMetadata } from '@/components/learn/ArticleListView'

interface PageProps {
  searchParams: Promise<{ tag?: string }>
}

export async function generateMetadata(): Promise<Metadata> {
  return articleListMetadata('ai')
}

export default async function AiListPage({ searchParams }: PageProps) {
  const { tag } = await searchParams
  return <ArticleListView category="ai" tag={tag} />
}
