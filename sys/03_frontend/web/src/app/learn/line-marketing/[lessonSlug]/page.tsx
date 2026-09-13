import Link from 'next/link'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { serverApi } from '@/lib/serverApi'
import { excerpt } from '@/lib/markdown'
import Markdown from '@/components/learn/Markdown'
import LessonInteractive from '@/components/learn/LessonInteractive'
import type { Course, Lesson } from '@/lib/types'

const COURSE_SLUG = 'line-marketing'

interface PageProps {
  params: Promise<{ lessonSlug: string }>
}

async function getCourseAndLesson(lessonSlug: string): Promise<{ course: Course; lesson: Lesson; index: number }> {
  const course = await serverApi.getOrNull<Course>(`/api/courses/${COURSE_SLUG}`)
  const lessons = course?.lessons ?? []
  const index = lessons.findIndex((l) => l.slug === lessonSlug)
  if (!course || index === -1) {
    notFound()
  }
  return { course, lesson: lessons[index], index }
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { lessonSlug } = await params
  const { lesson } = await getCourseAndLesson(lessonSlug)
  return {
    title: lesson.title,
    description: excerpt(lesson.body, 140),
  }
}

export default async function LessonPage({ params }: PageProps) {
  const { lessonSlug } = await params
  const { course, lesson, index } = await getCourseAndLesson(lessonSlug)
  const lessons = course.lessons ?? []
  const prev = index > 0 ? lessons[index - 1] : null
  const next = index < lessons.length - 1 ? lessons[index + 1] : null

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <Link href={`/learn/${COURSE_SLUG}`} className="text-sm font-medium text-line-green hover:underline">
        ← {course.title}へ戻る
      </Link>
      <h1 className="mt-4 text-3xl font-bold text-gray-900">{lesson.title}</h1>

      <div className="mt-8">
        <Markdown>{lesson.body}</Markdown>
      </div>

      <LessonInteractive courseSlug={COURSE_SLUG} lessonId={lesson.id} />

      <nav className="mt-10 flex items-center justify-between border-t border-gray-100 pt-6 text-sm font-medium">
        {prev ? (
          <Link href={`/learn/${COURSE_SLUG}/${prev.slug}`} className="text-line-green hover:underline">
            ← {prev.title}
          </Link>
        ) : (
          <span />
        )}
        {next ? (
          <Link href={`/learn/${COURSE_SLUG}/${next.slug}`} className="text-line-green hover:underline">
            {next.title} →
          </Link>
        ) : (
          <span />
        )}
      </nav>
    </div>
  )
}
