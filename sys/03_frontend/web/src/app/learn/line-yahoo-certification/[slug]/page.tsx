import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import Link from 'next/link'
import { serverApi } from '@/lib/serverApi'
import QuizModeSelect from '@/components/quiz/QuizModeSelect'
import type { Quiz } from '@/lib/types'

interface PageProps {
  params: Promise<{ slug: string }>
}

async function getQuiz(slug: string): Promise<Quiz> {
  const quiz = await serverApi.getOrNull<Quiz>(`/api/quizzes/${slug}`)
  if (!quiz) {
    notFound()
  }
  return quiz
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { slug } = await params
  const quiz = await getQuiz(slug)
  return {
    title: quiz.title,
    description: quiz.description || undefined,
  }
}

export default async function QuizPage({ params }: PageProps) {
  const { slug } = await params
  const quiz = await getQuiz(slug)

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <Link href="/learn/line-yahoo-certification" className="text-sm font-medium text-line-green hover:underline">
        ← LINEヤフー認定資格一覧へ戻る
      </Link>
      <h1 className="mt-4 text-3xl font-bold text-gray-900">{quiz.title}</h1>
      {quiz.description && <p className="mt-2 text-gray-600">{quiz.description}</p>}

      <div className="mt-8">
        <QuizModeSelect quiz={quiz} />
      </div>
    </div>
  )
}
