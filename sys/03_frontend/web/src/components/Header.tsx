'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { siteContent } from '@/data/siteContent'
import { useAuth } from '@/lib/auth'
import { loginUrl } from '@/lib/api'

// Nav order decided in dev-plan-07-frontend-base (task 7.3), resolving the
// open question left in README.md: learning content is ordered
// マーケティング講座 → 運用・設定 → AI活用事例 (flagship course first, then the
// practical how-to category, then the smaller AI use-case category), followed
// by 導入事例.
const learnLinks = [
  { href: '/learn/line-marketing', label: 'マーケティング講座' },
  { href: '/learn/line-operation', label: '運用・設定' },
  { href: '/learn/ai', label: 'AI活用事例' },
]

export default function Header() {
  const [scrolled, setScrolled] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)
  const { user, loading, logout } = useAuth()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 50)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  const topLevelLinks = [
    { href: '/learn', label: '学習コンテンツ' },
    { href: '/usecase', label: '導入事例' },
    { href: '/#contact', label: siteContent.nav.contact },
  ]

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled ? 'bg-white/95 backdrop-blur-md shadow-sm' : 'bg-transparent'
      }`}
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <Link href="/" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-line-green">
              <svg className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <span className="text-lg font-bold text-gray-900">{siteContent.brand}</span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-8">
            {topLevelLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="text-sm font-medium text-gray-600 hover:text-line-green transition-colors"
              >
                {link.label}
              </Link>
            ))}
            <AuthControls loading={loading} user={user} logout={logout} />
          </nav>

          {/* Mobile hamburger */}
          <button className="md:hidden p-2" onClick={() => setMenuOpen(!menuOpen)} aria-label="メニューを開く">
            <svg className="h-6 w-6 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              {menuOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>
      </div>

      {/* Mobile menu */}
      {menuOpen && (
        <div className="md:hidden bg-white/95 backdrop-blur-md border-t border-gray-100">
          <div className="px-4 py-4 space-y-3">
            <p className="text-xs font-semibold uppercase text-gray-400">学習コンテンツ</p>
            {learnLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="block pl-2 text-sm font-medium text-gray-600 hover:text-line-green"
                onClick={() => setMenuOpen(false)}
              >
                {link.label}
              </Link>
            ))}
            {topLevelLinks.slice(1).map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="block text-sm font-medium text-gray-600 hover:text-line-green"
                onClick={() => setMenuOpen(false)}
              >
                {link.label}
              </Link>
            ))}
            <div className="pt-2">
              <AuthControls loading={loading} user={user} logout={logout} />
            </div>
          </div>
        </div>
      )}
    </header>
  )
}

function AuthControls({
  loading,
  user,
  logout,
}: {
  loading: boolean
  user: ReturnType<typeof useAuth>['user']
  logout: () => Promise<void>
}) {
  if (loading) {
    return null
  }

  if (user) {
    return (
      <div className="flex items-center gap-3">
        <span className="text-sm text-gray-600">{user.display_name}</span>
        <button
          onClick={() => void logout()}
          className="text-sm font-medium text-gray-500 hover:text-line-green transition-colors"
        >
          ログアウト
        </button>
      </div>
    )
  }

  return (
    <div className="flex items-center gap-3">
      <a
        href={loginUrl('line')}
        className="rounded-full bg-line-green px-4 py-1.5 text-sm font-medium text-white hover:bg-line-green-dark transition-colors"
      >
        LINEでログイン
      </a>
      <a
        href={loginUrl('google')}
        className="rounded-full border border-gray-300 px-4 py-1.5 text-sm font-medium text-gray-700 hover:border-gray-400 transition-colors"
      >
        Googleでログイン
      </a>
    </div>
  )
}
