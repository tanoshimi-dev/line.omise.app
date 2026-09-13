import type { Metadata } from 'next'
import { Noto_Sans_JP } from 'next/font/google'
import Script from 'next/script'
import { AuthProvider } from '@/lib/auth'
import Header from '@/components/Header'
import Footer from '@/components/Footer'
import './globals.css'

// Self-hosted via next/font instead of the old index.html's Google Fonts
// <link> tags (dev-plan-08-frontend-lp 8.2) — same font, no extra external
// request/layout shift. globals.css's --font-sans references this variable.
const notoSansJP = Noto_Sans_JP({
  weight: ['400', '500', '600', '700', '800'],
  subsets: ['latin'],
  variable: '--font-noto-sans-jp',
  display: 'swap',
})

// GA4 tag carried over as-is from the old index.html (dev-plan-08-frontend-lp
// 8.2) — same measurement ID, now site-wide via next/script instead of a
// page's own <script> tags.
const GA_MEASUREMENT_ID = 'G-L6C568RP5H'

export const metadata: Metadata = {
  metadataBase: new URL('https://line.omise.app'),
  title: {
    default: 'line.omise.app',
    template: '%s | line.omise.app',
  },
  description: 'はんなりdev — LINE Login / Google OAuth 対応の学習コンテンツプラットフォーム',
  icons: {
    icon: '/favicon.svg',
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja" className={notoSansJP.variable}>
      <body className="min-h-screen">
        <AuthProvider>
          <Header />
          <main className="pt-16">{children}</main>
          <Footer />
        </AuthProvider>
        <Script src={`https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`} strategy="afterInteractive" />
        <Script id="ga4-init" strategy="afterInteractive">
          {`window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', '${GA_MEASUREMENT_ID}');`}
        </Script>
      </body>
    </html>
  )
}
