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
