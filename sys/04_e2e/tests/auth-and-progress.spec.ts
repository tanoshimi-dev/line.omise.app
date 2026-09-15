import { test, expect } from '@playwright/test'
import { seedUser } from './helpers/db'
import { loginAs } from './helpers/auth'

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
