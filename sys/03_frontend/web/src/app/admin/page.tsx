import Link from 'next/link'

const sections = [
  { href: '/admin/courses', title: '講座・レッスン', description: '講座の作成・レッスンの追加・試験の管理' },
  { href: '/admin/articles', title: '記事', description: 'LINE運用・設定 / 生成AI活用事例の記事管理' },
  { href: '/admin/usecases', title: '導入事例', description: '導入店舗インタビューの管理' },
]

export default function AdminDashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">ダッシュボード</h1>
      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        {sections.map((section) => (
          <Link
            key={section.href}
            href={section.href}
            className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
          >
            <h2 className="font-bold text-gray-900">{section.title}</h2>
            <p className="mt-1 text-sm text-gray-500">{section.description}</p>
          </Link>
        ))}
      </div>
    </div>
  )
}
