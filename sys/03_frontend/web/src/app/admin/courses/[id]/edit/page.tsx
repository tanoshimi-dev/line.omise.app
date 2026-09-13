'use client'

import { useEffect, useState, type FormEvent } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, NumberField, StatusSelect, SubmitButton } from '@/components/admin/FormField'
import DeleteButton from '@/components/admin/DeleteButton'
import type { Course } from '@/lib/types'

export default function EditCoursePage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const [course, setCourse] = useState<Course | null>(null)
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [sortOrder, setSortOrder] = useState(0)
  const [status, setStatus] = useState('draft')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void api.get<Course>(`/api/admin/courses/${id}`).then((data) => {
      setCourse(data)
      setSlug(data.slug)
      setTitle(data.title)
      setDescription(data.description)
      setSortOrder(data.sort_order)
      setStatus(data.status)
    })
  }, [id])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const updated = await api.put<Course>(`/api/admin/courses/${id}`, {
        slug,
        title,
        description,
        sort_order: sortOrder,
        status,
      })
      setCourse((prev) => (prev ? { ...prev, ...updated, lessons: prev.lessons } : updated))
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  if (!course) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  return (
    <div>
      <Link href="/admin/courses" className="text-sm font-medium text-line-green hover:underline">
        ← 講座一覧へ戻る
      </Link>
      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{course.title}</h1>
        <DeleteButton path={`/api/admin/courses/${id}`} confirmMessage="この講座を削除しますか？レッスンも削除されます。" redirectTo="/admin/courses" />
      </div>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-xl space-y-4">
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="説明" name="description" value={description} onChange={setDescription} rows={3} />
        <NumberField label="表示順" name="sort_order" value={sortOrder} onChange={setSortOrder} />
        <StatusSelect value={status} onChange={setStatus} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>保存する</SubmitButton>
      </form>

      <div className="mt-10 border-t border-gray-100 pt-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-bold text-gray-900">レッスン</h2>
          <Link
            href={`/admin/courses/${id}/lessons/new`}
            className="rounded-full bg-line-green px-4 py-2 text-sm font-semibold text-white hover:bg-line-green-dark"
          >
            + 新規レッスン
          </Link>
        </div>

        {(course.lessons ?? []).length === 0 ? (
          <p className="mt-4 text-gray-500">レッスンはまだありません。</p>
        ) : (
          <ul className="mt-4 space-y-2">
            {(course.lessons ?? []).map((lesson) => (
              <li key={lesson.id} className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
                <Link href={`/admin/lessons/${lesson.id}/edit`} className="font-medium text-line-green hover:underline">
                  {lesson.title}
                </Link>
                <span className={`text-xs font-semibold ${lesson.status === 'published' ? 'text-line-green' : 'text-gray-400'}`}>
                  {lesson.status === 'published' ? '公開' : '下書き'}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
