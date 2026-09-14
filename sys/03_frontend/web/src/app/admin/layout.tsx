import Link from 'next/link'
import type { Metadata } from 'next'
import { AdminGuard } from '@/components/admin/AdminGuard'

export const metadata: Metadata = {
  title: '管理画面',
  robots: { index: false, follow: false },
}

const navLinks = [
  { href: '/admin', label: 'ダッシュボード' },
  { href: '/admin/courses', label: '講座・レッスン' },
  { href: '/admin/articles', label: '記事' },
  { href: '/admin/usecases', label: '導入事例' },
  { href: '/admin/quizzes', label: 'クイズ・検定' },
]

// The session cookie is scoped to the separate API origin, so the Next.js
// server cannot inspect it while rendering line.omise.app. Wait for the
// browser-side AuthProvider (which calls api-line.omise.app with credentials)
// before gating the page. Every admin API remains protected by RequireAdmin.
export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminGuard>
      <div className="mx-auto max-w-6xl px-4 py-12 sm:py-16">
        <div className="flex flex-col gap-8 sm:flex-row">
          <nav className="shrink-0 sm:w-48">
            <p className="mb-3 text-xs font-semibold uppercase text-gray-400">管理画面</p>
            <ul className="space-y-1">
              {navLinks.map((link) => (
                <li key={link.href}>
                  <Link href={link.href} className="block rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100">
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>
          <div className="min-w-0 flex-1">{children}</div>
        </div>
      </div>
    </AdminGuard>
  )
}
