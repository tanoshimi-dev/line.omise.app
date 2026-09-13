'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'

// Shared confirm+delete button for the admin list/edit pages
// (dev-plan-11-frontend-admin — "共通化できる部分は共通コンポーネント化").
//
// Pass `redirectTo` when used on an edit page (navigates away after
// deleting); pass `onDeleted` when used in a list page instead, since these
// pages fetch their data client-side on mount — router.refresh() only
// re-runs Server Components, so it wouldn't re-trigger that fetch.
export default function DeleteButton({
  path,
  confirmMessage,
  redirectTo,
  onDeleted,
}: {
  path: string
  confirmMessage: string
  redirectTo?: string
  onDeleted?: () => void
}) {
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)

  const handleClick = async () => {
    if (!window.confirm(confirmMessage)) {
      return
    }
    setDeleting(true)
    try {
      await api.delete(path)
      if (redirectTo) {
        router.push(redirectTo)
      }
      onDeleted?.()
    } finally {
      setDeleting(false)
    }
  }

  return (
    <button
      onClick={() => void handleClick()}
      disabled={deleting}
      className="text-sm font-medium text-red-600 hover:text-red-700 disabled:opacity-50"
    >
      {deleting ? '削除中…' : '削除'}
    </button>
  )
}
