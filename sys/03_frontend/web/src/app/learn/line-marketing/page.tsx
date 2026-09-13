import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { serverApi } from '@/lib/serverApi'
import CourseLessonList from '@/components/learn/CourseLessonList'
import type { Course } from '@/lib/types'

const COURSE_SLUG = 'line-marketing'

async function getCourse(): Promise<Course> {
  const course = await serverApi.getOrNull<Course>(`/api/courses/${COURSE_SLUG}`)
  if (!course) {
    notFound()
  }
  return course
}

export async function generateMetadata(): Promise<Metadata> {
  const course = await getCourse()
  return {
    title: course.title,
    description: course.description,
  }
}

export default async function CourseTopPage() {
  const course = await getCourse()
  const lessons = course.lessons ?? []

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:py-24">
      <h1 className="text-3xl font-bold text-gray-900">{course.title}</h1>
      <p className="mt-4 text-gray-600">{course.description}</p>

      <div className="mt-10">
        <CourseLessonList courseSlug={COURSE_SLUG} lessons={lessons.map((l) => ({ slug: l.slug, title: l.title }))} />
      </div>
    </div>
  )
}
