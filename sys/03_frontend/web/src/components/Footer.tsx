import Link from 'next/link'
import { siteContent } from '@/data/siteContent'

const navLinks = [
  { href: '/learn', label: '学習コンテンツ' },
  { href: '/usecase', label: '導入事例' },
  { href: '/#contact', label: siteContent.nav.contact },
]

export default function Footer() {
  return (
    <footer className="bg-dark-navy py-12">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center gap-8">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-line-green">
              <svg className="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <span className="text-lg font-bold text-white">{siteContent.brand}</span>
          </div>

          <nav className="flex flex-wrap justify-center gap-6">
            {navLinks.map((link) => (
              <Link key={link.href} href={link.href} className="text-sm text-gray-400 hover:text-white transition-colors">
                {link.label}
              </Link>
            ))}
          </nav>

          <div className="w-full max-w-xs border-t border-gray-700" />

          <p className="text-sm text-gray-500">{siteContent.footer.copyright}</p>
        </div>
      </div>
    </footer>
  )
}
