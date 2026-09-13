import type { Metadata } from 'next'
import { AuthProvider } from '@/lib/auth'
import Header from '@/components/Header'
import Footer from '@/components/Footer'
import './globals.css'

export const metadata: Metadata = {
  title: 'line.omise.app',
  description: 'はんなりdev — LINE Login / Google OAuth 対応の学習コンテンツプラットフォーム',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body className="min-h-screen">
        <AuthProvider>
          <Header />
          <main className="pt-16">{children}</main>
          <Footer />
        </AuthProvider>
      </body>
    </html>
  )
}
