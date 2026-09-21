import Link from 'next/link'
import type { Metadata } from 'next'
import { serverApi } from '@/lib/serverApi'
import type { QuizListItem } from '@/lib/types'

export const metadata: Metadata = {
  title: 'LINEヤフー認定資格',
  description: 'LINEヤフー認定資格の取得に向けたクイズ・検定形式の学習コンテンツです。',
}

export default async function QuizListPage() {
  const data = await serverApi.get<{ quizzes: QuizListItem[] }>('/api/quizzes')

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">LINEヤフー認定資格</h1>
      <p className="mt-4 text-gray-600">
        LINEヤフー認定資格の取得に向けて、知識を確認できます。<br />
        ログインした状態で実施すると、回答結果が記録され、後から振り返ることができます。
      </p>

      {data.quizzes.length === 0 ? (
        <p className="mt-8 text-gray-500">クイズ・検定はまだありません。</p>
      ) : (
        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2">
          {data.quizzes.map((quiz) => (
            <Link
              key={quiz.id}
              href={`/learn/line-yahoo-certification/${quiz.slug}`}
              className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm transition-shadow hover:shadow-md"
            >
              <h3 className="font-bold text-gray-900">{quiz.title}</h3>
              {quiz.description && <p className="mt-2 text-sm text-gray-600">{quiz.description}</p>}
              {quiz.passing_score != null && <p className="mt-3 text-xs text-gray-400">検定モードの合格点: {quiz.passing_score}%</p>}
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
