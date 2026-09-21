'use client'

// My-page (dev-plan-09-frontend-learn 9.4). Entirely user-specific — no
// public/SEO-relevant content — so this is a plain Client Component page
// rather than a Server Component shell.
//
// The course-progress section this page used to show was removed with the
// マーケティング講座 feature (dev-plan-2-9-line-yahoo-certification). The
// quiz/exam-progress section (QuizProgressSection) was kept — it's the same
// underlying feature, just rebranded under the LINEヤフー認定資格 nav
// entry — so it's still rendered here.

import { useAuth } from '@/lib/auth'
import LoginPrompt from '@/components/learn/LoginPrompt'
import QuizProgressSection from '@/components/mypage/QuizProgressSection'
import DeleteAccountSection from '@/components/mypage/DeleteAccountSection'

export default function MyProgressPage() {
  const { user, loading: authLoading } = useAuth()

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">マイページ</h1>

      {authLoading ? (
        <p className="mt-8 text-gray-400">読み込み中…</p>
      ) : !user ? (
        <div className="mt-8">
          <LoginPrompt message="ログインすると、学習の進捗を確認できます。" />
        </div>
      ) : (
        <>
          <QuizProgressSection />
          <DeleteAccountSection />
        </>
      )}
    </div>
  )
}
