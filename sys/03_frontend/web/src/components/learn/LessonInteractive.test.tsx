import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LessonInteractive from './LessonInteractive'
import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import type { CourseProgress, Exam, SubmitExamResponse } from '@/lib/types'

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

describe('LessonInteractive', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows a login prompt instead of any progress UI when logged out', () => {
    mockedUseAuth.mockReturnValue({ user: null, loading: false, refresh: vi.fn(), logout: vi.fn() })

    render(<LessonInteractive courseSlug="line-marketing" lessonId="1" />)

    expect(screen.getByText(/ログインすると/)).toBeInTheDocument()
    expect(screen.queryByText('レッスンを完了にする')).not.toBeInTheDocument()
  })

  it('marks the lesson complete when the button is clicked', async () => {
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn() })
    mockedApi.get.mockResolvedValueOnce({
      course: { id: '1', slug: 'line-marketing', title: 'Course' },
      completed_count: 0,
      total_count: 1,
      lessons: [{ id: '1', slug: 'intro', title: 'Intro', completed: false, completed_at: null, exam: null }],
    } satisfies CourseProgress)
    mockedApi.post.mockResolvedValueOnce(undefined)

    render(<LessonInteractive courseSlug="line-marketing" lessonId="1" />)

    const button = await screen.findByText('レッスンを完了にする')
    await userEvent.click(button)

    await waitFor(() => expect(screen.getByText('このレッスンは完了済みです')).toBeInTheDocument())
    expect(mockedApi.post).toHaveBeenCalledWith('/api/lessons/1/complete')
  })

  it('shows the already-completed state without a clickable button', async () => {
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn() })
    mockedApi.get.mockResolvedValueOnce({
      course: { id: '1', slug: 'line-marketing', title: 'Course' },
      completed_count: 1,
      total_count: 1,
      lessons: [{ id: '1', slug: 'intro', title: 'Intro', completed: true, completed_at: '2026-01-01T00:00:00Z', exam: null }],
    } satisfies CourseProgress)

    render(<LessonInteractive courseSlug="line-marketing" lessonId="1" />)

    await waitFor(() => expect(screen.getByText('このレッスンは完了済みです')).toBeInTheDocument())
    expect(screen.queryByText('レッスンを完了にする')).not.toBeInTheDocument()
  })

  it('lets the user take the exam and shows the graded result', async () => {
    mockedUseAuth.mockReturnValue({ user: baseUser, loading: false, refresh: vi.fn(), logout: vi.fn() })
    mockedApi.get
      .mockResolvedValueOnce({
        course: { id: '1', slug: 'line-marketing', title: 'Course' },
        completed_count: 0,
        total_count: 1,
        lessons: [
          {
            id: '1',
            slug: 'intro',
            title: 'Intro',
            completed: false,
            completed_at: null,
            exam: { id: '10', title: 'Quiz', passing_score: 70, latest_result: null },
          },
        ],
      } satisfies CourseProgress)
      .mockResolvedValueOnce({
        id: '10',
        lesson_id: '1',
        title: 'Quiz',
        passing_score: 70,
        questions: [
          {
            id: '100',
            question_text: '2+2?',
            sort_order: 1,
            choices: [
              { id: '1000', choice_text: '4', sort_order: 1 },
              { id: '1001', choice_text: '5', sort_order: 2 },
            ],
          },
        ],
      } satisfies Exam)
    mockedApi.post.mockResolvedValueOnce({
      exam_id: '10',
      score: 100,
      passed: true,
      submitted_at: '2026-01-01T00:00:00Z',
    } satisfies SubmitExamResponse)

    render(<LessonInteractive courseSlug="line-marketing" lessonId="1" />)

    const startButton = await screen.findByText('試験を受ける')
    await userEvent.click(startButton)

    const choice = await screen.findByLabelText('4')
    await userEvent.click(choice)
    await userEvent.click(screen.getByText('解答を送信する'))

    await waitFor(() =>
      expect(screen.getByText((_, el) => el?.textContent === '前回のスコア: 100点（合格）')).toBeInTheDocument(),
    )
    expect(mockedApi.post).toHaveBeenCalledWith('/api/lessons/1/exam/submit', {
      answers: [{ question_id: 100, choice_id: 1000 }],
    })
  })
})
