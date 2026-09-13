'use client'

import { useEffect, useState, type FormEvent } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api, ApiError } from '@/lib/api'
import { TextField, NumberField, SubmitButton } from '@/components/admin/FormField'
import type { AdminExam, AdminExamQuestion } from '@/lib/types'

interface ChoiceDraft {
  choice_text: string
  is_correct: boolean
}

const emptyChoices: ChoiceDraft[] = [
  { choice_text: '', is_correct: false },
  { choice_text: '', is_correct: false },
]

export default function LessonExamPage() {
  const { id: lessonId } = useParams<{ id: string }>()
  const [loading, setLoading] = useState(true)
  const [exam, setExam] = useState<AdminExam | null>(null)
  const [questions, setQuestions] = useState<AdminExamQuestion[]>([])

  useEffect(() => {
    api
      .get<AdminExam>(`/api/admin/lessons/${lessonId}/exam`)
      .then((data) => {
        setExam(data)
        setQuestions(data.questions ?? [])
      })
      .catch((err) => {
        if (!(err instanceof ApiError && err.status === 404)) {
          console.error(err)
        }
      })
      .finally(() => setLoading(false))
  }, [lessonId])

  if (loading) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  return (
    <div>
      <Link href={`/admin/lessons/${lessonId}/edit`} className="text-sm font-medium text-line-green hover:underline">
        ← レッスンへ戻る
      </Link>
      <h1 className="mt-4 text-2xl font-bold text-gray-900">試験の管理</h1>

      {!exam ? (
        <CreateExamForm lessonId={lessonId} onCreated={(created) => setExam(created)} />
      ) : (
        <div className="mt-6 space-y-6">
          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4">
            <p className="font-bold text-gray-900">{exam.title}</p>
            <p className="text-sm text-gray-500">合格点: {exam.passing_score}点</p>
          </div>

          <div className="space-y-4">
            {questions.map((question) => (
              <QuestionItem
                key={question.id}
                question={question}
                onUpdated={(updated) => setQuestions((prev) => prev.map((q) => (q.id === updated.id ? updated : q)))}
                onDeleted={() => setQuestions((prev) => prev.filter((q) => q.id !== question.id))}
              />
            ))}
          </div>

          <QuestionForm
            examId={exam.id}
            submitLabel="設問を追加する"
            nextSortOrder={questions.length + 1}
            onSubmit={async (payload) => {
              const created = await api.post<AdminExamQuestion>(`/api/admin/exams/${exam.id}/questions`, payload)
              setQuestions((prev) => [...prev, created])
            }}
          />
        </div>
      )}
    </div>
  )
}

function CreateExamForm({ lessonId, onCreated }: { lessonId: string; onCreated: (exam: AdminExam) => void }) {
  const [title, setTitle] = useState('')
  const [passingScore, setPassingScore] = useState(70)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const exam = await api.post<AdminExam>(`/api/admin/lessons/${lessonId}/exam`, { title, passing_score: passingScore })
      onCreated({ ...exam, questions: [] })
    } catch {
      setError('作成に失敗しました。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mt-6">
      <p className="text-gray-500">このレッスンにはまだ試験がありません。</p>
      <form onSubmit={(e) => void handleSubmit(e)} className="mt-4 max-w-md space-y-4">
        <TextField label="試験タイトル" name="title" value={title} onChange={setTitle} required />
        <NumberField label="合格点（0〜100）" name="passing_score" value={passingScore} onChange={setPassingScore} />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>試験を作成する</SubmitButton>
      </form>
    </div>
  )
}

function QuestionItem({
  question,
  onUpdated,
  onDeleted,
}: {
  question: AdminExamQuestion
  onUpdated: (question: AdminExamQuestion) => void
  onDeleted: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    if (!window.confirm('この設問を削除しますか？')) return
    setDeleting(true)
    try {
      await api.delete(`/api/admin/questions/${question.id}`)
      onDeleted()
    } finally {
      setDeleting(false)
    }
  }

  if (editing) {
    return (
      <QuestionForm
        examId={question.exam_id}
        submitLabel="更新する"
        nextSortOrder={question.sort_order}
        initialQuestionText={question.question_text}
        initialSortOrder={question.sort_order}
        initialChoices={question.choices.map((c) => ({ choice_text: c.choice_text, is_correct: c.is_correct }))}
        onCancel={() => setEditing(false)}
        onSubmit={async (payload) => {
          const updated = await api.put<AdminExamQuestion>(`/api/admin/questions/${question.id}`, payload)
          onUpdated(updated)
          setEditing(false)
        }}
      />
    )
  }

  return (
    <div className="rounded-xl border border-gray-100 p-4">
      <div className="flex items-start justify-between">
        <p className="font-medium text-gray-900">{question.question_text}</p>
        <div className="flex shrink-0 gap-3 pl-4">
          <button onClick={() => setEditing(true)} className="text-sm font-medium text-line-green hover:underline">
            編集
          </button>
          <button onClick={() => void handleDelete()} disabled={deleting} className="text-sm font-medium text-red-600 hover:text-red-700 disabled:opacity-50">
            {deleting ? '削除中…' : '削除'}
          </button>
        </div>
      </div>
      <ul className="mt-2 space-y-1">
        {question.choices.map((choice) => (
          <li key={choice.id} className={`text-sm ${choice.is_correct ? 'font-semibold text-line-green' : 'text-gray-600'}`}>
            {choice.is_correct && '✓ '}
            {choice.choice_text}
          </li>
        ))}
      </ul>
    </div>
  )
}

interface QuestionPayload {
  question_text: string
  sort_order: number
  choices: { choice_text: string; is_correct: boolean; sort_order: number }[]
}

function QuestionForm({
  examId,
  submitLabel,
  nextSortOrder,
  initialQuestionText = '',
  initialSortOrder,
  initialChoices = emptyChoices,
  onSubmit,
  onCancel,
}: {
  examId: string
  submitLabel: string
  nextSortOrder: number
  initialQuestionText?: string
  initialSortOrder?: number
  initialChoices?: ChoiceDraft[]
  onSubmit: (payload: QuestionPayload) => Promise<void>
  onCancel?: () => void
}) {
  const [questionText, setQuestionText] = useState(initialQuestionText)
  const [sortOrder, setSortOrder] = useState(initialSortOrder ?? nextSortOrder)
  const [choices, setChoices] = useState<ChoiceDraft[]>(initialChoices)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const updateChoice = (index: number, patch: Partial<ChoiceDraft>) => {
    setChoices((prev) => prev.map((c, i) => (i === index ? { ...c, ...patch } : c)))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!choices.some((c) => c.is_correct)) {
      setError('正解の選択肢を1つ以上チェックしてください。')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await onSubmit({
        question_text: questionText,
        sort_order: sortOrder,
        choices: choices.map((c, i) => ({ choice_text: c.choice_text, is_correct: c.is_correct, sort_order: i + 1 })),
      })
      if (!initialSortOrder) {
        setQuestionText('')
        setChoices(emptyChoices)
      }
    } catch {
      setError('保存に失敗しました。')
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="space-y-3 rounded-xl border border-dashed border-gray-300 p-4">
      <TextField label="設問文" name="question_text" value={questionText} onChange={setQuestionText} required />
      <NumberField label="表示順" name="sort_order" value={sortOrder} onChange={setSortOrder} />

      <div className="space-y-2">
        <span className="text-sm font-medium text-gray-700">選択肢（正解にチェック）</span>
        {choices.map((choice, i) => (
          <div key={i} className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={choice.is_correct}
              onChange={(e) => updateChoice(i, { is_correct: e.target.checked })}
              className="accent-line-green"
            />
            <input
              type="text"
              value={choice.choice_text}
              onChange={(e) => updateChoice(i, { choice_text: e.target.value })}
              required
              className="flex-1 rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
            />
            {choices.length > 2 && (
              <button type="button" onClick={() => setChoices((prev) => prev.filter((_, idx) => idx !== i))} className="text-sm text-gray-400 hover:text-red-600">
                削除
              </button>
            )}
          </div>
        ))}
        <button
          type="button"
          onClick={() => setChoices((prev) => [...prev, { choice_text: '', is_correct: false }])}
          className="text-sm font-medium text-line-green hover:underline"
        >
          + 選択肢を追加
        </button>
      </div>

      {error && <p className="text-sm text-red-600">{error}</p>}
      <div className="flex items-center gap-3">
        <SubmitButton saving={saving}>{submitLabel}</SubmitButton>
        {onCancel && (
          <button type="button" onClick={onCancel} className="text-sm text-gray-500 hover:underline">
            キャンセル
          </button>
        )}
      </div>
    </form>
  )
}
