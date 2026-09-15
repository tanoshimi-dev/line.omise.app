// Shapes returned by line-api's /api/* content endpoints (dev-plan-05/06).
// All ids are strings — the backend serializes int64 ids as strings to avoid
// JS number precision issues (same convention as CurrentUser.id in auth.tsx).

export interface Course {
  id: string
  slug: string
  title: string
  description: string
  sort_order: number
  status: 'draft' | 'published'
  created_at: string
  updated_at: string
  lessons?: Lesson[]
}

export interface Lesson {
  id: string
  course_id: string
  slug: string
  title: string
  body: string
  sort_order: number
  status: 'draft' | 'published'
  created_at: string
  updated_at: string
}

export interface Tag {
  id: string
  name: string
  slug: string
}

export type ArticleCategory = 'line-operation' | 'ai'

export interface Article {
  id: string
  category: ArticleCategory
  slug: string
  title: string
  body: string
  status: 'draft' | 'published'
  published_at: string | null
  tags: Tag[]
  created_at: string
  updated_at: string
}

export interface ExamChoice {
  id: string
  choice_text: string
  sort_order: number
}

export interface ExamQuestion {
  id: string
  question_text: string
  sort_order: number
  choices: ExamChoice[]
}

// Reader-facing exam shape — never carries is_correct (dev-plan-06 6.2).
export interface Exam {
  id: string
  lesson_id: string
  title: string
  passing_score: number
  questions: ExamQuestion[]
}

// Admin-facing shapes DO carry is_correct (dev-plan-11-frontend-admin 11.5,
// GET /api/admin/lessons/:id/exam) — kept as separate types from the
// Reader-facing ones above so a stray is_correct can never leak by using
// the wrong type in the wrong place.
export interface AdminExamChoice {
  id: string
  choice_text: string
  is_correct: boolean
  sort_order: number
}

export interface AdminExamQuestion {
  id: string
  exam_id: string
  question_text: string
  sort_order: number
  choices: AdminExamChoice[]
}

export interface AdminExam {
  id: string
  lesson_id: string
  title: string
  passing_score: number
  questions?: AdminExamQuestion[]
}

export interface ExamResult {
  score: number
  passed: boolean
  submitted_at: string
}

export interface SubmitExamResponse extends ExamResult {
  exam_id: string
}

export interface CourseProgressLesson {
  id: string
  slug: string
  title: string
  completed: boolean
  completed_at: string | null
  exam: {
    id: string
    title: string
    passing_score: number
    latest_result: ExamResult | null
  } | null
}

export interface CourseProgress {
  course: { id: string; slug: string; title: string }
  completed_count: number
  total_count: number
  lessons: CourseProgressLesson[]
}

export interface MyProgressCourse {
  course_id: string
  slug: string
  title: string
  completed_count: number
  total_count: number
}

export interface MyProgress {
  courses: MyProgressCourse[]
}

// Quiz content — dev-plan-2-2-admin-api. A reader chooses per-attempt
// whether to answer question-by-question or submit the whole quiz at once
// (dev-plan-quiz-mode-selection); unrelated to the Phase 1 lesson-tied
// Exam/AdminExam types above (see that plan's terminology note).

export interface AdminQuizChoice {
  id: string
  choice_text: string
  is_correct: boolean
  sort_order: number
}

export interface AdminQuizQuestion {
  id: string
  quiz_id: string
  question_text: string
  allow_multiple: boolean
  explanation: string
  reference_url: string
  sort_order: number
  choices: AdminQuizChoice[]
}

export interface AdminQuiz {
  id: string
  slug: string
  title: string
  description: string
  passing_score: number | null
  published: boolean
  questions?: AdminQuizQuestion[]
}

// Reader/public-facing quiz shapes (dev-plan-2-3/2-4) — never carry
// is_correct/explanation/reference_url until after answering/submitting,
// mirroring the Exam/AdminExam split above.
export interface QuizListItem {
  id: string
  slug: string
  title: string
  description: string
  passing_score: number | null
}

export interface QuizChoice {
  id: string
  choice_text: string
  sort_order: number
}

export interface QuizQuestion {
  id: string
  question_text: string
  allow_multiple: boolean
  sort_order: number
  choices: QuizChoice[]
}

export interface Quiz {
  id: string
  slug: string
  title: string
  description: string
  passing_score: number | null
  questions: QuizQuestion[]
}

// Response to POST /api/quiz-questions/:id/answer — reveals the answer for
// this one question only, after it's been answered.
export interface QuizAnswerResult {
  question_id: string
  is_correct: boolean
  selected_choice_ids: string[]
  correct_choice_ids: string[]
  explanation: string
  reference_url: string
}

export interface QuizSubmitQuestionResult {
  question_id: string
  question_text: string
  selected_choice_ids: string[]
  correct_choice_ids: string[]
  is_correct: boolean
  explanation: string
  reference_url: string
}

// Response to POST /api/quizzes/:slug/submit (exam mode). passed is null
// when the quiz has no passing_score configured; attempt_id is null when
// the caller wasn't logged in (nothing was saved).
export interface QuizSubmitResult {
  quiz_id: string
  score: number
  total_questions: number
  passed: boolean | null
  questions: QuizSubmitQuestionResult[]
  attempt_id: string | null
}

// Reader mypage: quiz/exam history and progress (dev-plan-2-5-frontend-mypage).

export interface QuizPracticeHistoryEntry {
  id: string
  question_id: string
  question_text: string
  selected_choice_ids: string[]
  is_correct: boolean
  answered_at: string
}

export interface QuizPracticeHistory {
  quiz_id: string
  history: QuizPracticeHistoryEntry[]
}

export interface QuizAttemptSummary {
  id: string
  score: number
  total_questions: number
  passed: boolean | null
  started_at: string
  submitted_at: string
}

export interface QuizAttemptsList {
  quiz_id: string
  attempts: QuizAttemptSummary[]
}

export interface QuizAttemptQuestionReview {
  question_id: string
  question_text: string
  selected_choice_ids: string[]
  correct_choice_ids: string[]
  is_correct: boolean
  explanation: string
  reference_url: string
}

export interface QuizAttemptDetail extends QuizAttemptSummary {
  questions: QuizAttemptQuestionReview[]
}

// Combines question-by-question and whole-quiz-submission progress for one
// quiz, since a reader may have used either or both (dev-plan-quiz-mode-selection).
export interface QuizProgressSummary {
  quiz_id: string
  slug: string
  title: string
  total_questions: number
  answered_count: number
  correct_count: number
  attempt_count: number
  passing_score: number | null
  best_score: number | null
  latest_score: number | null
  latest_passed: boolean | null
}

export interface MyQuizzesProgress {
  quizzes: QuizProgressSummary[]
}

// related_demo_app matches a DemoApp.id from src/data/demoApps.ts (dev-plan-10
// migration 006) — empty string when no mini-app is associated.
export interface Usecase {
  id: string
  slug: string
  client_name: string
  title: string
  body: string
  status: 'draft' | 'published'
  thumbnail_url: string
  related_demo_app: string
  published_at: string | null
  created_at: string
  updated_at: string
}
