'use client'

import { useState, type FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, NumberField, SubmitButton } from '@/components/admin/FormField'
import type { AdminQuiz } from '@/lib/types'

export default function NewQuizPage() {
  const router = useRouter()
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [hasPassingScore, setHasPassingScore] = useState(false)
  const [passingScore, setPassingScore] = useState(70)
  const [published, setPublished] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const quiz = await api.post<AdminQuiz>('/api/admin/quizzes', {
        slug,
        title,
        description,
        passing_score: hasPassingScore ? passingScore : null,
        published,
      })
      router.push(`/admin/quizzes/${quiz.id}/edit`)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Link href="/admin/quizzes" className="text-sm font-medium text-line-green hover:underline">
        ← クイズ・検定一覧へ戻る
      </Link>
      <h1 className="mt-4 text-2xl font-bold text-gray-900">新規クイズ・検定</h1>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-xl space-y-4">
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required placeholder="basic-line-exam" />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required placeholder="基礎LINE検定" />
        <TextAreaField label="説明" name="description" value={description} onChange={setDescription} rows={3} />

        <label className="flex items-center gap-2">
          <input type="checkbox" checked={hasPassingScore} onChange={(e) => setHasPassingScore(e.target.checked)} className="accent-line-green" />
          <span className="text-sm font-medium text-gray-700">合格点を設定する（検定モードでの合否判定に使われます）</span>
        </label>
        {hasPassingScore && <NumberField label="合格点（0〜100）" name="passing_score" value={passingScore} onChange={setPassingScore} />}

        <label className="flex items-center gap-2">
          <input type="checkbox" checked={published} onChange={(e) => setPublished(e.target.checked)} className="accent-line-green" />
          <span className="text-sm font-medium text-gray-700">公開する</span>
        </label>

        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>作成する</SubmitButton>
      </form>
    </div>
  )
}
