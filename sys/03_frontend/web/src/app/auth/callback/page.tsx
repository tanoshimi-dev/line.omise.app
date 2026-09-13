'use client'

// Landing page the backend redirects to after a LINE/Google login completes
// (dev-plan-04-auth issues the session cookie, then redirects here). Confirms
// the session via /auth/me, then bounces to the homepage.

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'

export default function AuthCallbackPage() {
  const router = useRouter()
  const { refresh } = useAuth()

  useEffect(() => {
    void refresh().then(() => router.replace('/'))
  }, [refresh, router])

  return (
    <div className="mx-auto max-w-md px-4 py-24 text-center text-gray-600">
      ログイン処理中です…
    </div>
  )
}
