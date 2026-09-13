import Link from 'next/link'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: '学習コンテンツ',
  description: 'LINEマーケティング講座、LINE運用・設定記事、生成AI活用事例をまとめた学習コンテンツです。',
}

// Order matches Header.tsx's mobile menu, decided in dev-plan-07-frontend-base
// 7.3 (マーケティング講座 → 運用・設定 → AI活用事例) and reconfirmed here per
// dev-plan-09-frontend-learn 9.1 — no change was needed.
const categories = [
  {
    href: '/learn/line-marketing',
    title: 'LINEマーケティング講座',
    description: '基礎から学ぶ、LINEを使ったマーケティング戦略の講座です。',
  },
  {
    href: '/learn/line-operation',
    title: 'LINE運用・設定',
    description: 'リッチメニューや公式アカウントの設定など、実践的な運用ノウハウ記事です。',
  },
  {
    href: '/learn/ai',
    title: '生成AI活用事例',
    description: '生成AIをビジネスに活用する事例を紹介する記事です。',
  },
]

export default function LearnTopPage() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">学習コンテンツ</h1>
      <p className="mt-4 text-gray-600">LINEアプリ運用に役立つ講座・記事をまとめています。</p>

      <div className="mt-12 grid grid-cols-1 gap-6 sm:grid-cols-3">
        {categories.map((category) => (
          <Link
            key={category.href}
            href={category.href}
            className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm transition-shadow hover:shadow-md"
          >
            <h2 className="text-lg font-bold text-gray-900">{category.title}</h2>
            <p className="mt-2 text-sm text-gray-600">{category.description}</p>
          </Link>
        ))}
      </div>
    </div>
  )
}
