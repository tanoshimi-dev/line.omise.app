import Link from 'next/link'
import { redirect } from 'next/navigation'
import type { Metadata } from 'next'
import { getCurrentUserServer } from '@/lib/serverAuth'

export const metadata: Metadata = {
  title: '管理画面',
  robots: { index: false, follow: false },
}

const navLinks = [
  { href: '/admin', label: 'ダッシュボード' },
  { href: '/admin/courses', label: '講座・レッスン' },
  { href: '/admin/articles', label: '記事' },
  { href: '/admin/usecases', label: '導入事例' },
]

// Gates every /admin/* route (dev-plan-11-frontend-admin 11.1). Not logged
// in and Reader both land on "/" — every page already has LINE/Google login
// buttons in the header, so a dedicated "please log in" page isn't needed;
// see the result doc for why both cases share one destination.
export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUserServer()
  if (!user || user.role !== 'admin') {
    redirect('/')
  }

  return (
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
  )
}
