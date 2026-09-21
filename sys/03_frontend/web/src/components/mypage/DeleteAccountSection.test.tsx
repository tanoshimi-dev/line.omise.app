import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import DeleteAccountSection from './DeleteAccountSection'
import { useAuth } from '@/lib/auth'

vi.mock('@/lib/auth', () => ({ useAuth: vi.fn() }))

const mockedUseAuth = vi.mocked(useAuth)
const deleteAccount = vi.fn()

describe('DeleteAccountSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn(), deleteAccount })
  })

  it('requires the exact confirmation phrase before deleting', async () => {
    const user = userEvent.setup()
    render(<DeleteAccountSection />)

    expect(screen.getByText(/すべてのデータ/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '退会する' }))
    const deleteButton = screen.getByRole('button', { name: 'アカウントを削除する' })
    expect(deleteButton).toBeDisabled()

    await user.type(screen.getByLabelText('確認語句'), '退会します')
    expect(deleteButton).toBeDisabled()
    await user.clear(screen.getByLabelText('確認語句'))
    await user.type(screen.getByLabelText('確認語句'), '退会する')
    expect(deleteButton).toBeEnabled()
  })

  it('calls account deletion once after confirmation', async () => {
    const user = userEvent.setup()
    deleteAccount.mockResolvedValue(undefined)
    render(<DeleteAccountSection />)

    await user.click(screen.getByRole('button', { name: '退会する' }))
    await user.type(screen.getByLabelText('確認語句'), '退会する')
    await user.click(screen.getByRole('button', { name: 'アカウントを削除する' }))

    expect(deleteAccount).toHaveBeenCalledTimes(1)
  })

  it('shows an error and allows retry when deletion fails', async () => {
    const user = userEvent.setup()
    deleteAccount.mockRejectedValueOnce(new Error('network'))
    render(<DeleteAccountSection />)

    await user.click(screen.getByRole('button', { name: '退会する' }))
    await user.type(screen.getByLabelText('確認語句'), '退会する')
    await user.click(screen.getByRole('button', { name: 'アカウントを削除する' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('退会処理に失敗しました')
    expect(screen.getByRole('button', { name: 'アカウントを削除する' })).toBeEnabled()
  })
})
