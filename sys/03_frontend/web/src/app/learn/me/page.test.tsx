import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import MyProgressPage from './page'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { MyQuizzesProgress } from '@/lib/types'

// dev-plan-2-9-line-yahoo-certification: マイページからコース進捗セクションを
// 削除した後の、ログイン状態切り替えとクイズ・検定（LINEヤフー認定資格）
// セクション表示を検証するテスト。QuizProgressSection 自体にログイン判定は
// ない（dev-plan-2-5 2-5.4 の設計）ため、このページ側でその分岐を検証する。

vi.mock('@/lib/auth', () => ({
  useAuth: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  api: { get: vi.fn(), post: vi.fn() },
  loginUrl: (provider: string) => `/auth/${provider}/login`,
}))

const mockedUseAuth = vi.mocked(useAuth)
const mockedApi = vi.mocked(api)

const baseUser = { id: '1', provider: 'google' as const, email: 'a@example.com', display_name: 'A', avatar_url: '', role: 'reader' as const }

describe('MyProgressPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows only a login prompt, and no quiz/exam section, when logged out', async () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn(), deleteAccount: vi.fn() })

    render(<MyProgressPage />)

    expect(await screen.findByText(/ログインすると/)).toBeInTheDocument()
    expect(screen.queryByText('LINEヤフー認定資格')).not.toBeInTheDocument()
    expect(mockedApi.get).not.toHaveBeenCalled()
  })

  it('shows the quiz/exam section when logged in', async () => {
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn(), deleteAccount: vi.fn() })
    mockedApi.get.mockImplementation((path: unknown) => {
      if (path === '/api/me/quizzes/progress') {
        return Promise.resolve({ quizzes: [] } satisfies MyQuizzesProgress)
      }
      return Promise.reject(new Error(`unexpected path ${String(path)}`))
    })

    render(<MyProgressPage />)

    await waitFor(() => expect(screen.getByText('LINEヤフー認定資格')).toBeInTheDocument())
    expect(screen.getByText('クイズ・検定はまだありません。')).toBeInTheDocument()
  })
})
