'use client'

import { useState, type FormEvent } from 'react'
import { TextField, TextAreaField, StatusSelect, SubmitButton } from './FormField'
import { demoApps } from '@/data/demoApps'

export interface UsecaseFormValues {
  slug: string
  client_name: string
  title: string
  body: string
  status: string
  thumbnail_url: string
  related_demo_app: string
}

const defaultValues: UsecaseFormValues = {
  slug: '',
  client_name: '',
  title: '',
  body: '',
  status: 'draft',
  thumbnail_url: '',
  related_demo_app: '',
}

// Shared by /admin/usecases/new and /admin/usecases/[id]/edit
// (dev-plan-11-frontend-admin — "共通化できる部分は共通コンポーネント化").
export default function UsecaseForm({
  initialValues = defaultValues,
  submitLabel,
  onSubmit,
}: {
  initialValues?: UsecaseFormValues
  submitLabel: string
  onSubmit: (values: UsecaseFormValues) => Promise<void>
}) {
  const [values, setValues] = useState(initialValues)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const set = <K extends keyof UsecaseFormValues>(key: K) => (value: UsecaseFormValues[K]) => setValues((prev) => ({ ...prev, [key]: value }))

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      await onSubmit(values)
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="max-w-2xl space-y-4">
      <TextField label="スラッグ" name="slug" value={values.slug} onChange={set('slug')} required placeholder="sample-salon" />
      <TextField label="店舗名" name="client_name" value={values.client_name} onChange={set('client_name')} required />
      <TextField label="タイトル" name="title" value={values.title} onChange={set('title')} required />
      <TextAreaField label="本文（Markdown）" name="body" value={values.body} onChange={set('body')} rows={10} />
      <TextField label="サムネイル画像URL" name="thumbnail_url" value={values.thumbnail_url} onChange={set('thumbnail_url')} placeholder="/images/salon-reservation.png" />

      <label className="block">
        <span className="text-sm font-medium text-gray-700">関連するミニアプリ</span>
        <select
          value={values.related_demo_app}
          onChange={(e) => set('related_demo_app')(e.target.value)}
          className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
        >
          <option value="">なし</option>
          {demoApps.map((app) => (
            <option key={app.id} value={app.id}>
              {app.name}
            </option>
          ))}
        </select>
      </label>

      <StatusSelect value={values.status} onChange={set('status')} />
      {error && <p className="text-sm text-red-600">{error}</p>}
      <SubmitButton saving={saving}>{submitLabel}</SubmitButton>
    </form>
  )
}
