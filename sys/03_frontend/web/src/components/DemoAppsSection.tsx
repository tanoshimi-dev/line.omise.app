import { siteContent } from '@/data/siteContent'
import { demoApps } from '@/data/demoApps'
import DemoAppCard from './DemoAppCard'

export default function DemoAppsSection() {
  return (
    <section id="demos" className="py-20 sm:py-28 bg-gray-50">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Section header */}
        <div className="text-center max-w-2xl mx-auto mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900">
            {siteContent.demos.sectionTitle}
          </h2>
          <p className="mt-4 text-lg text-gray-600">
            {siteContent.demos.sectionSubtitle}
          </p>
        </div>

        {/* Cards grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {demoApps.map((app, index) => (
            <DemoAppCard key={app.id} app={app} index={index} />
          ))}
        </div>
      </div>
    </section>
  )
}
