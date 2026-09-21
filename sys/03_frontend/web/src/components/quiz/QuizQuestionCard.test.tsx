import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import QuizQuestionCard from './QuizQuestionCard'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { Quiz, QuizAnswerResult } from '@/lib/types'

vi.mock('@/lib/auth', () => ({
  useAuth: vi.fn(),
}))

vi.mock('@/lib/api', () => ({
  api: { get: vi.fn(), post: vi.fn() },
  loginUrl: (provider: string) => `/auth/${provider}/login`,
  ApiError: class ApiError extends Error {
    status: number
    constructor(message: string, status: number) {
      super(message)
      this.status = status
    }
  },
}))

const mockedUseAuth = vi.mocked(useAuth)
const mockedApi = vi.mocked(api)

const baseUser = { id: '1', provider: 'google' as const, email: 'a@example.com', display_name: 'A', avatar_url: '', role: 'reader' as const }

const oneQuestionQuiz: Quiz = {
  id: '1',
  slug: 'sample-quiz',
  title: 'Sample Quiz',
  description: '',
  passing_score: null,
  questions: [
    {
      id: '100',
      question_text: '2+2?',
      allow_multiple: false,
      sort_order: 1,
      choices: [
        { id: '1000', choice_text: '4', sort_order: 1 },
        { id: '1001', choice_text: '5', sort_order: 2 },
      ],
    },
  ],
}

describe('QuizQuestionCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn(), deleteAccount: vi.fn() })
  })

  it('answers a question, shows the result and explanation, then finishes', async () => {
    mockedApi.post.mockResolvedValueOnce({
      question_id: '100',
      is_correct: true,
      selected_choice_ids: ['1000'],
      correct_choice_ids: ['1000'],
      explanation: 'Because 2+2=4.',
      reference_url: '',
    } satisfies QuizAnswerResult)

    render(<QuizQuestionCard quiz={oneQuestionQuiz} />)

    const choice = screen.getByLabelText('4')
    await userEvent.click(choice)
    await userEvent.click(screen.getByText('解答する'))

    await waitFor(() => expect(screen.getByText('正解です！')).toBeInTheDocument())
    expect(screen.getByText('Because 2+2=4.')).toBeInTheDocument()
    expect(mockedApi.post).toHaveBeenCalledWith('/api/quiz-questions/100/answer', { choice_ids: [1000] })

    await userEvent.click(screen.getByText('結果を見る'))

    await waitFor(() => expect(screen.getByText('1問中 1問正解しました')).toBeInTheDocument())
    expect(screen.queryByText(/ログインすると/)).not.toBeInTheDocument()
  })

  it('shows a login prompt on the finish screen when logged out', async () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn(), deleteAccount: vi.fn() })
    mockedApi.post.mockResolvedValueOnce({
      question_id: '100',
      is_correct: false,
      selected_choice_ids: ['1001'],
      correct_choice_ids: ['1000'],
      explanation: 'Because 2+2=4.',
      reference_url: '',
    } satisfies QuizAnswerResult)

    render(<QuizQuestionCard quiz={oneQuestionQuiz} />)

    await userEvent.click(screen.getByLabelText('5'))
    await userEvent.click(screen.getByText('解答する'))
    await waitFor(() => expect(screen.getByText('不正解です')).toBeInTheDocument())
    await userEvent.click(screen.getByText('結果を見る'))

    await waitFor(() => expect(screen.getByText(/ログインすると/)).toBeInTheDocument())
  })

  it('disables the answer button until a choice is selected', () => {
    render(<QuizQuestionCard quiz={oneQuestionQuiz} />)
    expect(screen.getByText('解答する')).toBeDisabled()
  })

  it('going back to a graded question redisplays its result without re-submitting', async () => {
    const twoQuestionQuiz: Quiz = {
      ...oneQuestionQuiz,
      questions: [
        oneQuestionQuiz.questions[0],
        {
          id: '101',
          question_text: '3+3?',
          allow_multiple: false,
          sort_order: 2,
          choices: [
            { id: '1010', choice_text: '6', sort_order: 1 },
            { id: '1011', choice_text: '7', sort_order: 2 },
          ],
        },
      ],
    }
    mockedApi.post.mockResolvedValueOnce({
      question_id: '100',
      is_correct: true,
      selected_choice_ids: ['1000'],
      correct_choice_ids: ['1000'],
      explanation: 'Because 2+2=4.',
      reference_url: '',
    } satisfies QuizAnswerResult)

    render(<QuizQuestionCard quiz={twoQuestionQuiz} />)

    await userEvent.click(screen.getByLabelText('4'))
    await userEvent.click(screen.getByText('解答する'))
    await waitFor(() => expect(screen.getByText('正解です！')).toBeInTheDocument())
    expect(mockedApi.post).toHaveBeenCalledTimes(1)

    await userEvent.click(screen.getByText('次の問題へ'))
    await waitFor(() => expect(screen.getByText('3+3?')).toBeInTheDocument())

    await userEvent.click(screen.getByText('前の問題へ'))
    await waitFor(() => expect(screen.getByText('2+2?')).toBeInTheDocument())
    expect(screen.getByText('正解です！')).toBeInTheDocument()
    expect(mockedApi.post).toHaveBeenCalledTimes(1)
  })
})
