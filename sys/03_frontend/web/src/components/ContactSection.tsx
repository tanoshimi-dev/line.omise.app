import { siteContent } from '@/data/siteContent'

export default function ContactSection() {
  return (
    <section id="contact" className="py-20 sm:py-28 bg-gradient-to-br from-line-green to-line-green-dark relative overflow-hidden">
      {/* Decorative shapes */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-20 -right-20 h-80 w-80 rounded-full bg-white/10 blur-3xl" />
        <div className="absolute bottom-0 left-1/4 h-64 w-64 rounded-full bg-white/5 blur-3xl" />
      </div>

      <div className="relative mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 text-center">
        {/* Icon */}
        <div className="mx-auto mb-8 flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
          <svg className="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>

        <h2 className="text-3xl sm:text-4xl font-bold text-white whitespace-pre-line">
          {siteContent.contact.headline}
        </h2>
        <p className="mt-6 text-lg text-white/80 whitespace-pre-line">
          {siteContent.contact.subtitle}
        </p>

        <a
          href="#profile"
          className="mt-10 inline-flex items-center gap-2 rounded-full bg-white px-10 py-4 text-lg font-bold text-line-green shadow-xl hover:shadow-2xl hover:scale-105 transition-all duration-300"
        >
          <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          {siteContent.contact.cta}
        </a>
      </div>
    </section>
  )
}
