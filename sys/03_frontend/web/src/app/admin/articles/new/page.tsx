'use client'

import { useState, type FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, StatusSelect, SubmitButton } from '@/components/admin/FormField'
import { CATEGORY_LABELS } from '@/lib/articleCategories'
import type { Article, ArticleCategory } from '@/lib/types'

export default function NewArticlePage() {
  const router = useRouter()
  const [category, setCategory] = useState<ArticleCategory>('line-operation')
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [status, setStatus] = useState('draft')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const article = await api.post<Article>('/api/admin/articles', { category, slug, title, body, status })
      router.push(`/admin/articles/${article.id}/edit`)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Link href="/admin/articles" className="text-sm font-medium text-line-green hover:underline">
        ← 記事一覧へ戻る
      </Link>
      <h1 className="mt-4 text-2xl font-bold text-gray-900">新規記事</h1>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-2xl space-y-4">
        <label className="block">
          <span className="text-sm font-medium text-gray-700">カテゴリ</span>
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value as ArticleCategory)}
            className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
          >
            {Object.entries(CATEGORY_LABELS).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </label>
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required placeholder="rich-menu-basics" />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="本文（Markdown）" name="body" value={body} onChange={setBody} rows={14} />
        <StatusSelect value={status} onChange={setStatus} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>作成する</SubmitButton>
      </form>
    </div>
  )
}
