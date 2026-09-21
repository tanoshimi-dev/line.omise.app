'use client'

import { useState } from 'react'
import { useAuth } from '@/lib/auth'

const confirmationPhrase = '退会する'

// An intentionally separate danger-zone control: account deletion is
// irreversible and must never be conflated with deleting one quiz history.
export default function DeleteAccountSection() {
  const { deleteAccount } = useAuth()
  const [open, setOpen] = useState(false)
  const [confirmation, setConfirmation] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')

  const close = () => {
    if (deleting) return
    setOpen(false)
    setConfirmation('')
    setError('')
  }

  const handleDelete = async () => {
    if (confirmation !== confirmationPhrase || deleting) return
    setDeleting(true)
    setError('')
    try {
      await deleteAccount()
    } catch {
      setError('退会処理に失敗しました。時間をおいてもう一度お試しください。')
      setDeleting(false)
    }
  }

  return (
    <section className="mt-10 border-t border-red-100 pt-8" aria-labelledby="delete-account-heading">
      <h2 id="delete-account-heading" className="text-xl font-bold text-gray-900">退会</h2>
      <p className="mt-3 text-sm leading-6 text-gray-600">
        アカウントを削除すると、保存済みのクイズ解答・検定受験履歴を含むすべてのデータとログイン状態が削除されます。この操作は元に戻せません。
      </p>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="mt-4 rounded-lg border border-red-600 px-4 py-2 text-sm font-semibold text-red-600 hover:bg-red-50"
      >
        退会する
      </button>

      {open && (
        <div role="dialog" aria-modal="true" aria-labelledby="delete-account-dialog-title" className="mt-4 rounded-xl border border-red-200 bg-red-50 p-5">
          <h3 id="delete-account-dialog-title" className="font-bold text-gray-900">本当に退会しますか？</h3>
          <p className="mt-2 text-sm leading-6 text-gray-700">
            削除後はアカウント、すべてのクイズ解答履歴、検定受験履歴を復元できません。続けるには「{confirmationPhrase}」と入力してください。
          </p>
          <label className="mt-4 block text-sm font-medium text-gray-800" htmlFor="delete-account-confirmation">
            確認語句
          </label>
          <input
            id="delete-account-confirmation"
            value={confirmation}
            onChange={(event) => setConfirmation(event.target.value)}
            className="mt-1 w-full rounded-lg border border-gray-300 bg-white px-3 py-2"
            autoComplete="off"
          />
          {error && <p role="alert" className="mt-3 text-sm text-red-700">{error}</p>}
          <div className="mt-4 flex gap-3">
            <button type="button" onClick={close} disabled={deleting} className="rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-white disabled:opacity-50">
              キャンセル
            </button>
            <button
              type="button"
              onClick={() => void handleDelete()}
              disabled={confirmation !== confirmationPhrase || deleting}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {deleting ? '退会処理中…' : 'アカウントを削除する'}
            </button>
          </div>
        </div>
      )}
    </section>
  )
}
