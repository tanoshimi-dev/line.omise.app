import { test, expect } from '@playwright/test'
import { seedUser, seedQuiz, withDB } from './helpers/db'
import { loginAs } from './helpers/auth'

test('an unauthenticated visitor can pick single mode and answer a quiz', async ({ page }) => {
  const quiz = await seedQuiz()
  try {
    await page.goto(`/learn/quiz/${quiz.slug}`)
    await page.getByRole('button', { name: '単発モード' }).click()
    await page.getByLabel('4').check()
    await page.getByRole('button', { name: '解答する' }).click()

    await expect(page.getByText('正解です！')).toBeVisible()
    await expect(page.getByText('Because 2+2=4.')).toBeVisible()
    await expect(page.getByText(/ログインすると/)).not.toBeVisible()
  } finally {
    await quiz.cleanup()
  }
})

test('a logged-in reader can pick exam mode, submit a quiz, and see the final score', async ({ page, context, baseURL }) => {
  const quiz = await seedQuiz({ passingScore: 60 })
  const reader = await seedUser('reader')
  try {
    await loginAs(context, baseURL!, reader)

    await page.goto(`/learn/quiz/${quiz.slug}`)
    await page.getByRole('button', { name: '検定モード' }).click()
    await page.getByLabel('4').check()
    await page.getByRole('button', { name: '提出する' }).click()

    await expect(page.getByText('1 / 1問正解')).toBeVisible()
    await expect(page.getByText('合格（合格点: 60%）')).toBeVisible()
    await expect(page.getByText('Because 2+2=4.')).toBeVisible()
  } finally {
    await quiz.cleanup()
    await reader.cleanup()
  }
})

test('mypage shows single-mode answer history and exam attempt history after logging in', async ({ page, context, baseURL }) => {
  const singleModeQuiz = await seedQuiz()
  const examModeQuiz = await seedQuiz({ passingScore: 60 })
  const reader = await seedUser('reader')
  try {
    await loginAs(context, baseURL!, reader)

    await page.goto(`/learn/quiz/${singleModeQuiz.slug}`)
    await page.getByRole('button', { name: '単発モード' }).click()
    await page.getByLabel('4').check()
    await page.getByRole('button', { name: '解答する' }).click()
    await expect(page.getByText('正解です！')).toBeVisible()

    await page.goto(`/learn/quiz/${examModeQuiz.slug}`)
    await page.getByRole('button', { name: '検定モード' }).click()
    await page.getByLabel('4').check()
    await page.getByRole('button', { name: '提出する' }).click()
    await expect(page.getByText('1 / 1問正解')).toBeVisible()

    await page.goto('/learn/me')
    await expect(page.getByRole('heading', { name: 'クイズ・検定' })).toBeVisible()

    // Single-mode summary + expandable history.
    await expect(page.getByText(singleModeQuiz.title)).toBeVisible()
    await expect(page.getByText('単発モード解答済み: 1 / 1問（正答率 100%）')).toBeVisible()
    await page.getByRole('button', { name: '解答履歴を見る' }).click()
    await expect(page.getByText('2+2?')).toBeVisible()

    // Exam-mode summary + expandable attempt list + attempt detail modal.
    await expect(page.getByText(examModeQuiz.title)).toBeVisible()
    await expect(page.getByText(/受験回数: 1回/)).toBeVisible()
    await expect(page.getByText('直近の結果: 合格')).toBeVisible()
    await page.getByRole('button', { name: '受験履歴を見る' }).click()
    // Scoped to a <button> role: the single-mode card above also contains the
    // substring "1 / 1問" in a plain <p>, which getByText's default
    // substring match would otherwise hit instead of the attempt row.
    await page.getByRole('button', { name: /1 \/ 1問/ }).click()
    await expect(page.getByText('受験詳細')).toBeVisible()
    await expect(page.getByText('Because 2+2=4.')).toBeVisible()
    await page.getByRole('button', { name: '閉じる', exact: true }).click()
    await expect(page.getByText('受験詳細')).not.toBeVisible()
  } finally {
    await singleModeQuiz.cleanup()
    await examModeQuiz.cleanup()
    await reader.cleanup()
  }
})

test('admin can create a quiz with a question, publish it, and it becomes answerable publicly', async ({ page, context, baseURL }) => {
  const admin = await seedUser('admin')
  const slug = `e2e-admin-quiz-${Date.now()}`
  let quizId: string | null = null
  try {
    await loginAs(context, baseURL!, admin)

    await page.goto('/admin/quizzes/new')
    await page.getByLabel('スラッグ').fill(slug)
    await page.getByLabel('タイトル').fill('E2E Admin Quiz')
    await page.getByLabel('公開する').check()
    await page.getByRole('button', { name: '作成する' }).click()

    await expect(page).toHaveURL(/\/admin\/quizzes\/\d+\/edit$/)
    quizId = page.url().match(/\/admin\/quizzes\/(\d+)\/edit$/)?.[1] ?? null

    await page.getByLabel('設問文').fill('2+2?')
    await page.getByLabel('解説').fill('Because 2+2=4.')

    // The choice rows have no accessible label (dev-plan-2-2-admin-api's
    // admin UI leaves them as bare inputs), so they're located positionally
    // within the add-question form: two type="text" inputs precede them
    // (設問文, 参考リンク), and the "複数回答を許可する" checkbox precedes them.
    const questionForm = page.locator('form').nth(1)
    const choiceTextInputs = questionForm.locator('input[type="text"]')
    await choiceTextInputs.nth(2).fill('4')
    await choiceTextInputs.nth(3).fill('5')
    const choiceCheckboxes = questionForm.locator('input[type="checkbox"]')
    await choiceCheckboxes.nth(1).check() // the "4" choice is correct

    await page.getByRole('button', { name: '設問を追加する' }).click()
    await expect(page.getByText('単一回答')).toBeVisible()

    await page.goto(`/learn/quiz/${slug}`)
    await page.getByRole('button', { name: '単発モード' }).click()
    await page.getByLabel('4').check()
    await page.getByRole('button', { name: '解答する' }).click()
    await expect(page.getByText('正解です！')).toBeVisible()
  } finally {
    if (quizId) {
      await withDB((client) => client.query('DELETE FROM quizzes WHERE id = $1', [quizId]))
    }
    await admin.cleanup()
  }
})
