import { test, expect } from '@playwright/test'
import { seedUser } from './helpers/db'
import { loginAs } from './helpers/auth'

test.describe('Reader progress (dev-plan-06/09)', () => {
  test('completing a lesson updates the course progress badge', async ({ page, context, baseURL }) => {
    const reader = await seedUser('reader')
    try {
      await loginAs(context, baseURL!, reader)

      await page.goto('/learn/line-marketing/intro')
      await page.getByRole('button', { name: 'レッスンを完了にする' }).click()
      await expect(page.getByText('このレッスンは完了済みです')).toBeVisible()

      await page.goto('/learn/line-marketing')
      await expect(page.getByText('進捗: 1 / 1 レッスン完了')).toBeVisible()
      await expect(page.getByText('完了').first()).toBeVisible()
    } finally {
      await reader.cleanup()
    }
  })

  test('my-page progress reflects the completed lesson', async ({ page, context, baseURL }) => {
    const reader = await seedUser('reader')
    try {
      await loginAs(context, baseURL!, reader)
      await page.goto('/learn/line-marketing/intro')
      await page.getByRole('button', { name: 'レッスンを完了にする' }).click()
      await expect(page.getByText('このレッスンは完了済みです')).toBeVisible()

      await page.goto('/learn/me')
      await expect(page.getByText('LINEマーケティング講座')).toBeVisible()
      await expect(page.getByText('1 / 1')).toBeVisible()
    } finally {
      await reader.cleanup()
    }
  })
})

test.describe('Access control (dev-plan-11 11.1)', () => {
  test('unauthenticated visitors are redirected away from /admin', async ({ page }) => {
    await page.goto('/admin')
    await expect(page).toHaveURL(/\/$/)
  })

  test('a Reader is redirected away from /admin', async ({ page, context, baseURL }) => {
    const reader = await seedUser('reader')
    try {
      await loginAs(context, baseURL!, reader)
      await page.goto('/admin')
      await expect(page).toHaveURL(/\/$/)
    } finally {
      await reader.cleanup()
    }
  })

  test('an Admin can reach /admin', async ({ page, context, baseURL }) => {
    const admin = await seedUser('admin')
    try {
      await loginAs(context, baseURL!, admin)
      await page.goto('/admin')
      await expect(page.getByRole('heading', { name: 'ダッシュボード' })).toBeVisible()
    } finally {
      await admin.cleanup()
    }
  })
})
