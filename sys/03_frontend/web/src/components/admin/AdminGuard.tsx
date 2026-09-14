'use client'

import { useEffect, type ReactNode } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'

export function AdminGuard({ children }: { children: ReactNode }) {
  const router = useRouter()
  const { user, loading } = useAuth()

  useEffect(() => {
    if (!loading && user?.role !== 'admin') {
      router.replace('/')
    }
  }, [loading, router, user])

  if (loading || user?.role !== 'admin') {
    return (
      <div className="mx-auto max-w-md px-4 py-24 text-center text-gray-600">
        ログイン状態を確認しています…
      </div>
    )
  }

  return children
}
