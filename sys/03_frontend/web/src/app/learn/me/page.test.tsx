import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import MyProgressPage from './page'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { MyProgress, MyQuizzesProgress } from '@/lib/types'

// dev-plan-2-6-test 2-6.2: マイページの履歴・進捗コンポーネント（未ログイン時の表示切り替えを含む）.
// QuizProgressSection itself has no login check (dev-plan-2-5-frontend-mypage
// 2-5.4) — the gate lives here, in the page that mounts it — so this test
// exercises the page rather than the section in isolation.

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
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn() })

    render(<MyProgressPage />)

    expect(await screen.findByText(/ログインすると/)).toBeInTheDocument()
    expect(screen.queryByText('クイズ・検定')).not.toBeInTheDocument()
    expect(mockedApi.get).not.toHaveBeenCalled()
  })

  it('shows the course progress and the quiz/exam section when logged in', async () => {
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn() })
    mockedApi.get.mockImplementation((path: unknown) => {
      if (path === '/api/me/progress') {
        return Promise.resolve({ courses: [] } satisfies MyProgress)
      }
      if (path === '/api/me/quizzes/progress') {
        return Promise.resolve({ quizzes: [] } satisfies MyQuizzesProgress)
      }
      return Promise.reject(new Error(`unexpected path ${String(path)}`))
    })

    render(<MyProgressPage />)

    await waitFor(() => expect(screen.getByText('クイズ・検定')).toBeInTheDocument())
    expect(screen.getByText('受講中の講座はまだありません。')).toBeInTheDocument()
    expect(screen.getByText('クイズ・検定はまだありません。')).toBeInTheDocument()
  })
})
