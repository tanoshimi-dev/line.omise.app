'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'
import DeleteButton from '@/components/admin/DeleteButton'
import { CATEGORY_LABELS } from '@/lib/articleCategories'
import type { Article } from '@/lib/types'

export default function AdminArticlesPage() {
  const [articles, setArticles] = useState<Article[] | null>(null)

  useEffect(() => {
    void api.get<{ articles: Article[] }>('/api/admin/articles').then((data) => setArticles(data.articles))
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">記事</h1>
        <Link href="/admin/articles/new" className="rounded-full bg-line-green px-4 py-2 text-sm font-semibold text-white hover:bg-line-green-dark">
          + 新規記事
        </Link>
      </div>

      {articles === null ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : articles.length === 0 ? (
        <p className="mt-8 text-gray-500">記事はまだありません。</p>
      ) : (
        <table className="mt-6 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">タイトル</th>
              <th className="py-2 font-medium">カテゴリ</th>
              <th className="py-2 font-medium">状態</th>
              <th className="py-2 font-medium" />
            </tr>
          </thead>
          <tbody>
            {articles.map((article) => (
              <tr key={article.id} className="border-b border-gray-100">
                <td className="py-3">
                  <Link href={`/admin/articles/${article.id}/edit`} className="font-medium text-line-green hover:underline">
                    {article.title}
                  </Link>
                </td>
                <td className="py-3 text-gray-500">{CATEGORY_LABELS[article.category]}</td>
                <td className="py-3">
                  <span className={article.status === 'published' ? 'text-line-green' : 'text-gray-400'}>
                    {article.status === 'published' ? '公開' : '下書き'}
                  </span>
                </td>
                <td className="py-3 text-right">
                  <DeleteButton
                    path={`/api/admin/articles/${article.id}`}
                    confirmMessage={`「${article.title}」を削除しますか？`}
                    onDeleted={() => setArticles((prev) => prev?.filter((a) => a.id !== article.id) ?? null)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
