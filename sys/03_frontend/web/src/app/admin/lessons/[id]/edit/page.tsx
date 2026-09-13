'use client'

import { useEffect, useState, type FormEvent } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, NumberField, StatusSelect, SubmitButton } from '@/components/admin/FormField'
import DeleteButton from '@/components/admin/DeleteButton'
import type { Lesson } from '@/lib/types'

export default function EditLessonPage() {
  const { id } = useParams<{ id: string }>()
  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [sortOrder, setSortOrder] = useState(0)
  const [status, setStatus] = useState('draft')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void api.get<Lesson>(`/api/admin/lessons/${id}`).then((data) => {
      setLesson(data)
      setSlug(data.slug)
      setTitle(data.title)
      setBody(data.body)
      setSortOrder(data.sort_order)
      setStatus(data.status)
    })
  }, [id])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const updated = await api.put<Lesson>(`/api/admin/lessons/${id}`, {
        slug,
        title,
        body,
        sort_order: sortOrder,
        status,
      })
      setLesson(updated)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  if (!lesson) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  return (
    <div>
      <Link href={`/admin/courses/${lesson.course_id}/edit`} className="text-sm font-medium text-line-green hover:underline">
        ← 講座へ戻る
      </Link>
      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{lesson.title}</h1>
        <div className="flex items-center gap-4">
          <Link href={`/admin/lessons/${id}/exam`} className="text-sm font-medium text-line-green hover:underline">
            試験を管理
          </Link>
          <DeleteButton
            path={`/api/admin/lessons/${id}`}
            confirmMessage="このレッスンを削除しますか？"
            redirectTo={`/admin/courses/${lesson.course_id}/edit`}
          />
        </div>
      </div>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-2xl space-y-4">
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="本文（Markdown）" name="body" value={body} onChange={setBody} rows={14} />
        <NumberField label="表示順" name="sort_order" value={sortOrder} onChange={setSortOrder} />
        <StatusSelect value={status} onChange={setStatus} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>保存する</SubmitButton>
      </form>
    </div>
  )
}
