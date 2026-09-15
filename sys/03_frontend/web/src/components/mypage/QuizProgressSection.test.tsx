import { describe, expect, it, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import QuizProgressSection from './QuizProgressSection'
import { api } from '@/lib/api'
import type { MyQuizzesProgress, QuizPracticeHistory, QuizAttemptsList, QuizAttemptDetail } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  api: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}))

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), refresh: vi.fn() }),
}))

const mockedApi = vi.mocked(api)

describe('QuizProgressSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  it('shows an empty message when there are no published quizzes', async () => {
    mockedApi.get.mockResolvedValueOnce({ quizzes: [] } satisfies MyQuizzesProgress)

    render(<QuizProgressSection />)

    await waitFor(() => expect(screen.getByText('クイズ・検定はまだありません。')).toBeInTheDocument())
  })

  it('shows single-mode progress and expands to the answer history', async () => {
    mockedApi.get.mockResolvedValueOnce({
      quizzes: [
        {
          quiz_id: '1',
          slug: 'sample-quiz',
          title: 'Sample Quiz',
          total_questions: 2,
          answered_count: 1,
          correct_count: 1,
          attempt_count: 0,
          passing_score: null,
          best_score: null,
          latest_score: null,
          latest_passed: null,
        },
      ],
    } satisfies MyQuizzesProgress)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '1',
      history: [
        {
          id: '200',
          question_id: '100',
          question_text: '2+2?',
          selected_choice_ids: ['1000'],
          is_correct: true,
          answered_at: '2026-01-01T00:00:00Z',
        },
      ],
    } satisfies QuizPracticeHistory)

    render(<QuizProgressSection />)

    await screen.findByText('Sample Quiz')
    expect(screen.getByText('単発モード解答済み: 1 / 2問（正答率 100%）')).toBeInTheDocument()

    await userEvent.click(screen.getByText('解答履歴を見る'))
    await waitFor(() => expect(screen.getByText('2+2?')).toBeInTheDocument())
    expect(mockedApi.get).toHaveBeenCalledWith('/api/me/quizzes/sample-quiz/history')
  })

  it('shows exam-mode progress, attempt list, and attempt detail on click', async () => {
    mockedApi.get.mockResolvedValueOnce({
      quizzes: [
        {
          quiz_id: '2',
          slug: 'basic-exam',
          title: '基礎LINE検定',
          total_questions: 2,
          answered_count: 0,
          correct_count: 0,
          attempt_count: 1,
          passing_score: 60,
          best_score: 2,
          latest_score: 2,
          latest_passed: true,
        },
      ],
    } satisfies MyQuizzesProgress)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '2',
      attempts: [
        { id: '9', score: 2, total_questions: 2, passed: true, started_at: '2026-01-01T00:00:00Z', submitted_at: '2026-01-01T00:01:00Z' },
      ],
    } satisfies QuizAttemptsList)
    mockedApi.get.mockResolvedValueOnce({
      id: '9',
      score: 2,
      total_questions: 2,
      passed: true,
      started_at: '2026-01-01T00:00:00Z',
      submitted_at: '2026-01-01T00:01:00Z',
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
      ],
    } satisfies QuizAttemptDetail)

    render(<QuizProgressSection />)

    await screen.findByText('基礎LINE検定')
    expect(screen.getByText(/受験回数: 1回/)).toBeInTheDocument()
    expect(screen.getByText('直近の結果: 合格')).toBeInTheDocument()

    await userEvent.click(screen.getByText('受験履歴を見る'))
    const attemptRow = await screen.findByText('2 / 2問')
    await userEvent.click(attemptRow)

    await waitFor(() => expect(screen.getByText('Because 2+2=4.')).toBeInTheDocument())
    expect(mockedApi.get).toHaveBeenCalledWith('/api/me/quizzes/basic-exam/attempts/9')

    await userEvent.click(screen.getByText('閉じる'))
    expect(screen.queryByText('Because 2+2=4.')).not.toBeInTheDocument()
  })

  it('shows both single-mode and exam-mode sections for the same quiz', async () => {
    mockedApi.get.mockResolvedValueOnce({
      quizzes: [
        {
          quiz_id: '3',
          slug: 'both-modes-quiz',
          title: 'Both Modes Quiz',
          total_questions: 2,
          answered_count: 1,
          correct_count: 1,
          attempt_count: 1,
          passing_score: 60,
          best_score: 2,
          latest_score: 2,
          latest_passed: true,
        },
      ],
    } satisfies MyQuizzesProgress)

    render(<QuizProgressSection />)

    await screen.findByText('Both Modes Quiz')
    expect(screen.getByText('解答履歴を見る')).toBeInTheDocument()
    expect(screen.getByText('受験履歴を見る')).toBeInTheDocument()
  })

  it('deletes a single history entry and refreshes the progress summary', async () => {
    mockedApi.get.mockResolvedValueOnce({
      quizzes: [
        {
          quiz_id: '1',
          slug: 'sample-quiz',
          title: 'Sample Quiz',
          total_questions: 2,
          answered_count: 1,
          correct_count: 1,
          attempt_count: 0,
          passing_score: null,
          best_score: null,
          latest_score: null,
          latest_passed: null,
        },
      ],
    } satisfies MyQuizzesProgress)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '1',
      history: [
        {
          id: '200',
          question_id: '100',
          question_text: '2+2?',
          selected_choice_ids: ['1000'],
          is_correct: true,
          answered_at: '2026-01-01T00:00:00Z',
        },
      ],
    } satisfies QuizPracticeHistory)
    mockedApi.delete.mockResolvedValueOnce(undefined)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '1',
      slug: 'sample-quiz',
      title: 'Sample Quiz',
      total_questions: 2,
      answered_count: 0,
      correct_count: 0,
      attempt_count: 0,
      passing_score: null,
      best_score: null,
      latest_score: null,
      latest_passed: null,
    })

    render(<QuizProgressSection />)

    await screen.findByText('Sample Quiz')
    await userEvent.click(screen.getByText('解答履歴を見る'))
    await screen.findByText('2+2?')

    const deleteButtons = screen.getAllByText('削除')
    await userEvent.click(deleteButtons[deleteButtons.length - 1])

    expect(mockedApi.delete).toHaveBeenCalledWith('/api/me/quizzes/sample-quiz/history/200')
    await waitFor(() => expect(screen.getByText('解答履歴がありません。')).toBeInTheDocument())
    await waitFor(() => expect(mockedApi.get).toHaveBeenCalledWith('/api/me/quizzes/sample-quiz/progress'))
    await waitFor(() => expect(screen.getByText('単発モード解答済み: 0 / 2問')).toBeInTheDocument())
  })

  it('bulk-deletes all exam attempts and refreshes the progress summary', async () => {
    mockedApi.get.mockResolvedValueOnce({
      quizzes: [
        {
          quiz_id: '2',
          slug: 'basic-exam',
          title: '基礎LINE検定',
          total_questions: 2,
          answered_count: 0,
          correct_count: 0,
          attempt_count: 1,
          passing_score: 60,
          best_score: 2,
          latest_score: 2,
          latest_passed: true,
        },
      ],
    } satisfies MyQuizzesProgress)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '2',
      attempts: [
        { id: '9', score: 2, total_questions: 2, passed: true, started_at: '2026-01-01T00:00:00Z', submitted_at: '2026-01-01T00:01:00Z' },
      ],
    } satisfies QuizAttemptsList)
    mockedApi.delete.mockResolvedValueOnce(undefined)
    mockedApi.get.mockResolvedValueOnce({
      quiz_id: '2',
      slug: 'basic-exam',
      title: '基礎LINE検定',
      total_questions: 2,
      answered_count: 0,
      correct_count: 0,
      attempt_count: 0,
      passing_score: 60,
      best_score: null,
      latest_score: null,
      latest_passed: null,
    })

    render(<QuizProgressSection />)

    await screen.findByText('基礎LINE検定')
    await userEvent.click(screen.getByText('受験履歴を見る'))
    await screen.findByText('2 / 2問')

    const deleteButtons = screen.getAllByText('削除')
    await userEvent.click(deleteButtons[0])

    expect(window.confirm).toHaveBeenCalledWith('このクイズの受験履歴をすべて削除しますか？')
    expect(mockedApi.delete).toHaveBeenCalledWith('/api/me/quizzes/basic-exam/attempts')
    await waitFor(() => expect(screen.getByText('受験履歴がありません。')).toBeInTheDocument())
    await waitFor(() => expect(mockedApi.get).toHaveBeenCalledWith('/api/me/quizzes/basic-exam/progress'))
    await waitFor(() => expect(screen.getByText(/受験回数: 0回/)).toBeInTheDocument())
  })
})
