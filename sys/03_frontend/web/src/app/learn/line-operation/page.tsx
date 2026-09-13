import type { Metadata } from 'next'
import ArticleListView, { articleListMetadata } from '@/components/learn/ArticleListView'

interface PageProps {
  searchParams: Promise<{ tag?: string }>
}

export async function generateMetadata(): Promise<Metadata> {
  return articleListMetadata('line-operation')
}

export default async function LineOperationListPage({ searchParams }: PageProps) {
  const { tag } = await searchParams
  return <ArticleListView category="line-operation" tag={tag} />
}
