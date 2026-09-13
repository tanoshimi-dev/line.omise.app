import { useRef, useState } from 'react'
import type { DemoApp } from '@/data/demoApps'

interface Props {
  app: DemoApp
  index: number
}

export default function DemoAppCard({ app, index }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [isHovered, setIsHovered] = useState(false)

  const handleMouseEnter = () => {
    setIsHovered(true)
    videoRef.current?.play()
  }

  const handleMouseLeave = () => {
    setIsHovered(false)
    if (videoRef.current) {
      videoRef.current.pause()
      videoRef.current.currentTime = 0
    }
  }

  return (
    <div className="group relative rounded-2xl bg-white border border-gray-100 shadow-sm hover:shadow-xl transition-all duration-300 overflow-hidden">
      {/* Accent top bar */}
      <div className="h-1" style={{ backgroundColor: app.accentColor }} />

      <div className="p-6 sm:p-8">
        {/* Header */}
        <div className="flex items-start gap-4 mb-6">
          <div
            className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl text-white text-xl font-bold"
            style={{ backgroundColor: app.accentColor }}
          >
            {index + 1}
          </div>
          <div>
            <h3 className="text-xl font-bold text-gray-900">{app.name}</h3>
            <p className="mt-1 text-sm text-gray-500">{app.description}</p>
          </div>
        </div>

        {/* Screenshot / Movie */}
        <div
          className="relative rounded-xl overflow-hidden mb-6 cursor-pointer"
          style={{ backgroundColor: app.accentColorLight }}
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          {app.screenshot ? (
            <>
              <img
                src={app.screenshot}
                alt={`${app.name}のスクリーンショット`}
                className={`w-full h-48 object-cover transition-opacity duration-300 ${isHovered && app.movie ? 'opacity-0' : 'opacity-100'}`}
              />
              {app.movie && (
                <video
                  ref={videoRef}
                  src={app.movie}
                  muted
                  loop
                  playsInline
                  className={`absolute inset-0 w-full h-48 object-cover transition-opacity duration-300 ${isHovered ? 'opacity-100' : 'opacity-0'}`}
                />
              )}
              {/* Color overlay when not hovered */}
              <div
                className="absolute inset-0 transition-opacity duration-300 pointer-events-none"
                style={{ backgroundColor: app.accentColor, opacity: isHovered ? 0 : 0.15 }}
              />
            </>
          ) : (
            <div className="flex h-48 items-center justify-center">
              <div className="text-center">
                <svg
                  className="mx-auto h-12 w-12 opacity-30"
                  style={{ color: app.accentColor }}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={1.5}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.41a2.25 2.25 0 013.182 0l2.909 2.91m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z"
                  />
                </svg>
                <p className="mt-2 text-xs opacity-40" style={{ color: app.accentColor }}>
                  スクリーンショット準備中
                </p>
              </div>
            </div>
          )}
        </div>

        {/* Features */}
        <ul className="space-y-2 mb-6">
          {app.features.map((feature) => (
            <li key={feature} className="flex items-center gap-2 text-sm text-gray-700">
              <svg
                className="h-4 w-4 shrink-0"
                style={{ color: app.accentColor }}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
              {feature}
            </li>
          ))}
        </ul>

        {/* Tech badges */}
        <div className="flex flex-wrap gap-2">
          {app.techBadges.map((badge) => (
            <span
              key={badge}
              className="rounded-full px-3 py-1 text-xs font-medium"
              style={{
                backgroundColor: app.accentColorLight,
                color: app.accentColor,
              }}
            >
              {badge}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}
