'use client'

import { useState, type FormEvent } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, NumberField, StatusSelect, SubmitButton } from '@/components/admin/FormField'
import type { Lesson } from '@/lib/types'

export default function NewLessonPage() {
  const { id: courseId } = useParams<{ id: string }>()
  const router = useRouter()
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [sortOrder, setSortOrder] = useState(0)
  const [status, setStatus] = useState('draft')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const lesson = await api.post<Lesson>(`/api/admin/courses/${courseId}/lessons`, {
        slug,
        title,
        body,
        sort_order: sortOrder,
        status,
      })
      router.push(`/admin/lessons/${lesson.id}/edit`)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Link href={`/admin/courses/${courseId}/edit`} className="text-sm font-medium text-line-green hover:underline">
        ← 講座へ戻る
      </Link>
      <h1 className="mt-4 text-2xl font-bold text-gray-900">新規レッスン</h1>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-2xl space-y-4">
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required placeholder="intro" />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="本文（Markdown）" name="body" value={body} onChange={setBody} rows={14} />
        <NumberField label="表示順" name="sort_order" value={sortOrder} onChange={setSortOrder} />
        <StatusSelect value={status} onChange={setStatus} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>作成する</SubmitButton>
      </form>
    </div>
  )
}
