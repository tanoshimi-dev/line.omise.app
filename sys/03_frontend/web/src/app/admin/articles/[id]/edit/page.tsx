'use client'

import { useEffect, useState, type FormEvent } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, StatusSelect, SubmitButton } from '@/components/admin/FormField'
import DeleteButton from '@/components/admin/DeleteButton'
import { CATEGORY_LABELS } from '@/lib/articleCategories'
import type { Article, ArticleCategory, Tag } from '@/lib/types'

export default function EditArticlePage() {
  const { id } = useParams<{ id: string }>()
  const [article, setArticle] = useState<Article | null>(null)
  const [category, setCategory] = useState<ArticleCategory>('line-operation')
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [status, setStatus] = useState('draft')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void api.get<Article>(`/api/admin/articles/${id}`).then((data) => {
      setArticle(data)
      setCategory(data.category)
      setSlug(data.slug)
      setTitle(data.title)
      setBody(data.body)
      setStatus(data.status)
    })
  }, [id])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const updated = await api.put<Article>(`/api/admin/articles/${id}`, { category, slug, title, body, status })
      setArticle(updated)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  if (!article) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  return (
    <div>
      <Link href="/admin/articles" className="text-sm font-medium text-line-green hover:underline">
        ← 記事一覧へ戻る
      </Link>
      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{article.title}</h1>
        <DeleteButton path={`/api/admin/articles/${id}`} confirmMessage="この記事を削除しますか？" redirectTo="/admin/articles" />
      </div>

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
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="本文（Markdown）" name="body" value={body} onChange={setBody} rows={14} />
        <StatusSelect value={status} onChange={setStatus} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>保存する</SubmitButton>
      </form>

      <div className="mt-10 border-t border-gray-100 pt-6">
        <h2 className="text-lg font-bold text-gray-900">タグ</h2>
        <TagManager articleId={id} tags={article.tags} onAdded={(tag) => setArticle((prev) => (prev ? { ...prev, tags: [...prev.tags, tag] } : prev))} />
      </div>
    </div>
  )
}

function TagManager({ articleId, tags, onAdded }: { articleId: string; tags: Tag[]; onAdded: (tag: Tag) => void }) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const tag = await api.post<Tag>(`/api/admin/articles/${articleId}/tags`, { name, slug })
      onAdded(tag)
      setName('')
      setSlug('')
    } catch {
      setError('追加に失敗しました。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mt-4">
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {tags.map((tag) => (
            <span key={tag.id} className="rounded-full bg-line-green-light px-3 py-1 text-sm font-medium text-line-green">
              {tag.name}
            </span>
          ))}
        </div>
      )}
      {/* Tag removal isn't supported — dev-plan-05's API only ever added a
          create+attach endpoint, never a detach one, so this UI matches
          what the backend can actually do. */}
      <form onSubmit={(e) => void handleSubmit(e)} className="mt-4 flex items-end gap-3">
        <TextField label="タグ名" name="tag_name" value={name} onChange={setName} required placeholder="リッチメニュー" />
        <TextField label="タグスラッグ" name="tag_slug" value={slug} onChange={setSlug} required placeholder="rich-menu" />
        <SubmitButton saving={saving}>追加</SubmitButton>
      </form>
      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
    </div>
  )
}
