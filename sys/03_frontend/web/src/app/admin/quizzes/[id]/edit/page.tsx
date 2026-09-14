'use client'

import { useEffect, useState, type FormEvent } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'
import { TextField, TextAreaField, NumberField, SubmitButton } from '@/components/admin/FormField'
import DeleteButton from '@/components/admin/DeleteButton'
import type { AdminQuiz, AdminQuizQuestion, QuizMode } from '@/lib/types'

interface ChoiceDraft {
  choice_text: string
  is_correct: boolean
}

const emptyChoices: ChoiceDraft[] = [
  { choice_text: '', is_correct: false },
  { choice_text: '', is_correct: false },
]

export default function EditQuizPage() {
  const { id } = useParams<{ id: string }>()
  const [quiz, setQuiz] = useState<AdminQuiz | null>(null)
  const [questions, setQuestions] = useState<AdminQuizQuestion[]>([])
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<QuizMode>('practice')
  const [passingScore, setPassingScore] = useState(70)
  const [published, setPublished] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    void api.get<AdminQuiz>(`/api/admin/quizzes/${id}`).then((data) => {
      setQuiz(data)
      setQuestions(data.questions ?? [])
      setSlug(data.slug)
      setTitle(data.title)
      setDescription(data.description)
      setMode(data.mode)
      setPassingScore(data.passing_score ?? 70)
      setPublished(data.published)
    })
  }, [id])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const updated = await api.put<AdminQuiz>(`/api/admin/quizzes/${id}`, {
        slug,
        title,
        description,
        mode,
        passing_score: mode === 'exam' ? passingScore : null,
        published,
      })
      setQuiz((prev) => (prev ? { ...prev, ...updated, questions: prev.questions } : updated))
    } catch {
      setError('保存に失敗しました。スラッグが重複していないか確認してください。')
    } finally {
      setSaving(false)
    }
  }

  if (!quiz) {
    return <p className="text-gray-400">読み込み中…</p>
  }

  return (
    <div>
      <Link href="/admin/quizzes" className="text-sm font-medium text-line-green hover:underline">
        ← クイズ・検定一覧へ戻る
      </Link>
      <div className="mt-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{quiz.title}</h1>
        <DeleteButton path={`/api/admin/quizzes/${id}`} confirmMessage="このクイズ・検定を削除しますか？設問も削除されます。" redirectTo="/admin/quizzes" />
      </div>

      <form onSubmit={(e) => void handleSubmit(e)} className="mt-6 max-w-xl space-y-4">
        <TextField label="スラッグ" name="slug" value={slug} onChange={setSlug} required />
        <TextField label="タイトル" name="title" value={title} onChange={setTitle} required />
        <TextAreaField label="説明" name="description" value={description} onChange={setDescription} rows={3} />

        <label className="block">
          <span className="text-sm font-medium text-gray-700">種別</span>
          <select
            value={mode}
            onChange={(e) => setMode(e.target.value as QuizMode)}
            className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
          >
            <option value="practice">クイズ（1問ずつ即時採点）</option>
            <option value="exam">検定（一括提出・最終スコア）</option>
          </select>
        </label>

        {mode === 'exam' && <NumberField label="合格点（0〜100、任意）" name="passing_score" value={passingScore} onChange={setPassingScore} />}

        <label className="flex items-center gap-2">
          <input type="checkbox" checked={published} onChange={(e) => setPublished(e.target.checked)} className="accent-line-green" />
          <span className="text-sm font-medium text-gray-700">公開する</span>
        </label>

        {error && <p className="text-sm text-red-600">{error}</p>}
        <SubmitButton saving={saving}>保存する</SubmitButton>
      </form>

      <div className="mt-10 border-t border-gray-100 pt-6">
        <h2 className="text-lg font-bold text-gray-900">設問</h2>

        <div className="mt-4 space-y-4">
          {questions.map((question) => (
            <QuestionItem
              key={question.id}
              question={question}
              onUpdated={(updated) => setQuestions((prev) => prev.map((q) => (q.id === updated.id ? updated : q)))}
              onDeleted={() => setQuestions((prev) => prev.filter((q) => q.id !== question.id))}
            />
          ))}
        </div>

        <div className="mt-4">
          <QuestionForm
            quizId={quiz.id}
            submitLabel="設問を追加する"
            nextSortOrder={questions.length + 1}
            onSubmit={async (payload) => {
              const created = await api.post<AdminQuizQuestion>(`/api/admin/quizzes/${quiz.id}/questions`, payload)
              setQuestions((prev) => [...prev, created])
            }}
          />
        </div>
      </div>
    </div>
  )
}

function QuestionItem({
  question,
  onUpdated,
  onDeleted,
}: {
  question: AdminQuizQuestion
  onUpdated: (question: AdminQuizQuestion) => void
  onDeleted: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    if (!window.confirm('この設問を削除しますか？')) return
    setDeleting(true)
    try {
      await api.delete(`/api/admin/quiz-questions/${question.id}`)
      onDeleted()
    } finally {
      setDeleting(false)
    }
  }

  if (editing) {
    return (
      <QuestionForm
        quizId={question.quiz_id}
        submitLabel="更新する"
        nextSortOrder={question.sort_order}
        initialQuestionText={question.question_text}
        initialAllowMultiple={question.allow_multiple}
        initialExplanation={question.explanation}
        initialReferenceUrl={question.reference_url}
        initialSortOrder={question.sort_order}
        initialChoices={question.choices.map((c) => ({ choice_text: c.choice_text, is_correct: c.is_correct }))}
        onCancel={() => setEditing(false)}
        onSubmit={async (payload) => {
          const updated = await api.put<AdminQuizQuestion>(`/api/admin/quiz-questions/${question.id}`, payload)
          onUpdated(updated)
          setEditing(false)
        }}
      />
    )
  }

  return (
    <div className="rounded-xl border border-gray-100 p-4">
      <div className="flex items-start justify-between">
        <div>
          <p className="font-medium text-gray-900">{question.question_text}</p>
          <p className="mt-1 text-xs font-semibold text-gray-400">{question.allow_multiple ? '複数回答' : '単一回答'}</p>
        </div>
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
      <p className="mt-2 text-sm text-gray-500">{question.explanation}</p>
      {question.reference_url && (
        <a href={question.reference_url} target="_blank" rel="noreferrer" className="mt-1 inline-block text-xs text-line-green hover:underline">
          参考リンク
        </a>
      )}
    </div>
  )
}

interface QuestionPayload {
  question_text: string
  allow_multiple: boolean
  explanation: string
  reference_url: string
  sort_order: number
  choices: { choice_text: string; is_correct: boolean; sort_order: number }[]
}

function QuestionForm({
  quizId,
  submitLabel,
  nextSortOrder,
  initialQuestionText = '',
  initialAllowMultiple = false,
  initialExplanation = '',
  initialReferenceUrl = '',
  initialSortOrder,
  initialChoices = emptyChoices,
  onSubmit,
  onCancel,
}: {
  quizId: string
  submitLabel: string
  nextSortOrder: number
  initialQuestionText?: string
  initialAllowMultiple?: boolean
  initialExplanation?: string
  initialReferenceUrl?: string
  initialSortOrder?: number
  initialChoices?: ChoiceDraft[]
  onSubmit: (payload: QuestionPayload) => Promise<void>
  onCancel?: () => void
}) {
  const [questionText, setQuestionText] = useState(initialQuestionText)
  const [allowMultiple, setAllowMultiple] = useState(initialAllowMultiple)
  const [explanation, setExplanation] = useState(initialExplanation)
  const [referenceUrl, setReferenceUrl] = useState(initialReferenceUrl)
  const [sortOrder, setSortOrder] = useState(initialSortOrder ?? nextSortOrder)
  const [choices, setChoices] = useState<ChoiceDraft[]>(initialChoices)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const updateChoice = (index: number, patch: Partial<ChoiceDraft>) => {
    setChoices((prev) => prev.map((c, i) => (i === index ? { ...c, ...patch } : c)))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const correctCount = choices.filter((c) => c.is_correct).length
    if (choices.length < 2) {
      setError('選択肢は2つ以上必要です。')
      return
    }
    if (correctCount === 0) {
      setError('正解の選択肢を1つ以上チェックしてください。')
      return
    }
    if (!allowMultiple && correctCount > 1) {
      setError('複数回答を許可しない場合、正解は1つだけにしてください。')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await onSubmit({
        question_text: questionText,
        allow_multiple: allowMultiple,
        explanation,
        reference_url: referenceUrl,
        sort_order: sortOrder,
        choices: choices.map((c, i) => ({ choice_text: c.choice_text, is_correct: c.is_correct, sort_order: i + 1 })),
      })
      if (!initialSortOrder) {
        setQuestionText('')
        setExplanation('')
        setReferenceUrl('')
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

      <label className="flex items-center gap-2">
        <input type="checkbox" checked={allowMultiple} onChange={(e) => setAllowMultiple(e.target.checked)} className="accent-line-green" />
        <span className="text-sm font-medium text-gray-700">複数回答を許可する</span>
      </label>

      <TextAreaField label="解説" name="explanation" value={explanation} onChange={setExplanation} required rows={3} />
      <TextField label="参考リンク（任意）" name="reference_url" value={referenceUrl} onChange={setReferenceUrl} placeholder="https://..." />
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
