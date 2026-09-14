import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ExamRunner from './ExamRunner'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { Quiz, QuizSubmitResult } from '@/lib/types'

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

const examQuiz: Quiz = {
  id: '1',
  slug: 'basic-line-exam',
  title: '基礎LINE検定',
  description: '',
  mode: 'exam',
  passing_score: 60,
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

describe('ExamRunner', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn() })
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('submits all answers at once and shows the final score with a per-question review', async () => {
    mockedApi.post.mockResolvedValueOnce({
      quiz_id: '1',
      score: 2,
      total_questions: 2,
      passed: true,
      attempt_id: '5',
      questions: [
        {
          question_id: '100',
          question_text: '2+2?',
          selected_choice_ids: ['1000'],
          correct_choice_ids: ['1000'],
          is_correct: true,
          explanation: 'Because 2+2=4.',
          reference_url: '',
        },
        {
          question_id: '101',
          question_text: '3+3?',
          selected_choice_ids: ['1010'],
          correct_choice_ids: ['1010'],
          is_correct: true,
          explanation: 'Because 3+3=6.',
          reference_url: '',
        },
      ],
    } satisfies QuizSubmitResult)

    render(<ExamRunner quiz={examQuiz} />)

    await userEvent.click(screen.getByLabelText('4'))
    await userEvent.click(screen.getByLabelText('6'))
    await userEvent.click(screen.getByText('提出する'))

    await waitFor(() => expect(screen.getByText('2 / 2問正解')).toBeInTheDocument())
    expect(screen.getByText('合格（合格点: 60%）')).toBeInTheDocument()
    expect(screen.getByText('Because 2+2=4.')).toBeInTheDocument()
    expect(screen.getByText('Because 3+3=6.')).toBeInTheDocument()
    expect(window.confirm).not.toHaveBeenCalled()

    const [, body] = mockedApi.post.mock.calls[0]
    expect(body).toMatchObject({
      answers: [
        { question_id: 100, choice_ids: [1000] },
        { question_id: 101, choice_ids: [1010] },
      ],
    })
  })

  it('warns before submitting with unanswered questions', async () => {
    mockedApi.post.mockResolvedValueOnce({
      quiz_id: '1',
      score: 1,
      total_questions: 2,
      passed: false,
      attempt_id: '6',
      questions: [],
    } satisfies QuizSubmitResult)

    render(<ExamRunner quiz={examQuiz} />)

    await userEvent.click(screen.getByLabelText('4'))
    await userEvent.click(screen.getByText('提出する'))

    expect(window.confirm).toHaveBeenCalledWith('未回答の設問が1問あります。このまま提出しますか？')
    await waitFor(() => expect(mockedApi.post).toHaveBeenCalled())
  })

  it('does not submit when the user cancels the unanswered-question warning', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)

    render(<ExamRunner quiz={examQuiz} />)
    await userEvent.click(screen.getByText('提出する'))

    expect(window.confirm).toHaveBeenCalled()
    expect(mockedApi.post).not.toHaveBeenCalled()
  })

  it('shows a login prompt on the result screen when logged out', async () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn() })
    mockedApi.post.mockResolvedValueOnce({
      quiz_id: '1',
      score: 0,
      total_questions: 2,
      passed: false,
      attempt_id: null,
      questions: [],
    } satisfies QuizSubmitResult)

    render(<ExamRunner quiz={examQuiz} />)
    await userEvent.click(screen.getByText('提出する'))

    await waitFor(() => expect(screen.getByText(/ログインすると/)).toBeInTheDocument())
  })
})
