import { siteContent } from '@/data/siteContent'

export default function HeroSection() {
  return (
    <section className="relative min-h-screen flex items-center overflow-hidden bg-gradient-to-br from-white via-white to-line-green-light">
      {/* Decorative background shapes */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 h-96 w-96 rounded-full bg-line-green/5 blur-3xl" />
        <div className="absolute top-1/2 -left-20 h-72 w-72 rounded-full bg-line-green/8 blur-3xl" />
        <div className="absolute bottom-20 right-1/4 h-64 w-64 rounded-full bg-line-green/5 blur-3xl" />
      </div>

      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-24 pb-16">
        <div className="max-w-3xl">
          {/* Badge */}
          <div className="mb-6 inline-flex items-center gap-2 rounded-full bg-line-green/10 px-4 py-1.5">
            <div className="h-2 w-2 rounded-full bg-line-green animate-pulse" />
            <span className="text-sm font-medium text-line-green">
              {siteContent.hero.badge}
            </span>
          </div>

          {/* Headline */}
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold text-gray-900 leading-tight whitespace-pre-line">
            {siteContent.hero.headline}
          </h1>

          {/* Subtitle */}
          <p className="mt-6 text-lg sm:text-xl text-gray-600 leading-relaxed whitespace-pre-line max-w-2xl">
            {siteContent.hero.subtitle}
          </p>

          {/* CTAs */}
          <div className="mt-10 flex flex-wrap gap-4">
            <a
              href="#profile"
              className="inline-flex items-center gap-2 rounded-full bg-line-green px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-line-green/25 hover:bg-line-green-dark hover:shadow-xl hover:shadow-line-green/30 transition-all"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              {siteContent.hero.ctaPrimary}
            </a>
            <a
              href="#demos"
              className="inline-flex items-center gap-2 rounded-full border-2 border-gray-300 bg-white px-8 py-3.5 text-base font-semibold text-gray-700 hover:border-line-green hover:text-line-green transition-all"
            >
              {siteContent.hero.ctaSecondary}
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
              </svg>
            </a>
          </div>

          {/* Stats */}
          <div className="mt-16 flex gap-12">
            <div>
              <div className="text-3xl font-bold text-gray-900">3</div>
              <div className="text-sm text-gray-500">デモアプリ</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-gray-900">100%</div>
              <div className="text-sm text-gray-500">LINE完結</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-gray-900">24h</div>
              <div className="text-sm text-gray-500">即時通知</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
