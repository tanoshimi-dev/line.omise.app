import { test, expect } from '@playwright/test'
import { seedUser } from './helpers/db'
import { loginAs } from './helpers/auth'

// Usecases are used here: /learn/'s top page is fixed category cards
// (README's declared structure), not a dynamic list, so creating learn
// content wouldn't be visible from any public list page. The /usecase/
// list, by contrast, reflects any published usecase dynamically — exactly
// what's needed to prove "admin creates → publishes → shows up publicly"
// end to end.

test('admin can create a draft, publish it, see it publicly, then delete it', async ({ page, context, baseURL }) => {
  const admin = await seedUser('admin')
  const uniqueSlug = `e2e-usecase-${Date.now()}`
  const title = `E2E Test Usecase ${Date.now()}`

  try {
    await loginAs(context, baseURL!, admin)

    await page.goto('/admin/usecases/new')
    await page.getByLabel('スラッグ *').fill(uniqueSlug)
    await page.getByLabel('店舗名 *').fill('E2E Client')
    await page.getByLabel('タイトル *').fill(title)
    await page.getByRole('button', { name: '作成する' }).click()

    await expect(page).toHaveURL(new RegExp(`/admin/usecases/\\d+/edit$`))

    // Draft: must not appear on the public list yet.
    await page.goto('/usecase')
    await expect(page.getByText(title)).not.toBeVisible()

    // Publish it.
    await page.goBack()
    await page.getByLabel('公開状態').selectOption('published')
    await page.getByRole('button', { name: '保存する' }).click()
    await expect(page.getByRole('button', { name: '保存する' })).toBeVisible() // form re-rendered after save

    await page.goto('/usecase')
    await expect(page.getByText(title)).toBeVisible()

    // Delete it via the admin UI, then confirm it's gone from the public list.
    await page.goto('/admin/usecases')
    page.once('dialog', (dialog) => dialog.accept())
    await page.getByRole('row', { name: new RegExp(title) }).getByRole('button', { name: '削除' }).click()
    await expect(page.getByText(title)).not.toBeVisible()

    await page.goto('/usecase')
    await expect(page.getByText(title)).not.toBeVisible()
  } finally {
    await admin.cleanup()
  }
})
