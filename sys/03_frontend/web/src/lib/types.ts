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
