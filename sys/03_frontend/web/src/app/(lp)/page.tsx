import type { Metadata } from 'next'
import HeroSection from '@/components/HeroSection'
import DemoAppsSection from '@/components/DemoAppsSection'
import FeaturesSection from '@/components/FeaturesSection'
import WhyUsSection from '@/components/WhyUsSection'
import ContactSection from '@/components/ContactSection'
import ProfileSection from '@/components/ProfileSection'

// Migrated from the old index.html <head> (dev-plan-08-frontend-lp 8.2). Kept
// on the page itself, not the root layout, since /learn, /usecase, /admin
// will each define their own metadata in later steps — only the homepage
// needs this full OGP/Twitter/JSON-LD treatment.
const title = 'はんなりdev | LINE ミニアプリ開発 - 会員管理・予約・ECアプリ'
const description =
  'はんなりdev - LINE公式アカウント連携のミニアプリ開発。会員管理・予約システム・ECなど、ビジネスに最適なLINEアプリを開発します。'

export const metadata: Metadata = {
  // `absolute` opts out of the root layout's `%s | line.omise.app` title
  // template — this string already is the full page title.
  title: { absolute: title },
  description,
  keywords: [
    'はんなりdev',
    'LINE ミニアプリ',
    'LINE アプリ開発',
    '会員管理',
    '予約システム',
    'LINE公式アカウント',
    'LINEミニアプリ開発',
  ],
  alternates: {
    canonical: 'https://line.omise.app/',
  },
  openGraph: {
    title: 'はんなりdev | LINE ミニアプリ開発',
    description,
    url: 'https://line.omise.app/',
    siteName: 'はんなりdev',
    images: ['https://line.omise.app/images/og-image.png'],
    locale: 'ja_JP',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'はんなりdev | LINE ミニアプリ開発',
    description,
    images: ['https://line.omise.app/images/og-image.png'],
  },
}

const jsonLd = {
  '@context': 'https://schema.org',
  '@graph': [
    {
      '@type': 'Organization',
      name: 'はんなりdev',
      url: 'https://line.omise.app',
      description,
    },
    {
      '@type': 'WebSite',
      name: 'はんなりdev',
      url: 'https://line.omise.app',
      description: 'LINE ミニアプリ開発 - 会員管理・予約・ECアプリ',
    },
    {
      '@type': 'LocalBusiness',
      name: 'はんなりdev',
      url: 'https://line.omise.app',
      description: 'LINE公式アカウント連携のミニアプリ開発サービス',
      address: {
        '@type': 'PostalAddress',
        addressCountry: 'JP',
        addressRegion: '京都府',
      },
      priceRange: 'お問い合わせ',
    },
  ],
}

export default function HomePage() {
  return (
    <>
      {/* eslint-disable-next-line react/no-danger */}
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <HeroSection />
      <DemoAppsSection />
      <FeaturesSection />
      <WhyUsSection />
      <ContactSection />
      <ProfileSection />
    </>
  )
}
