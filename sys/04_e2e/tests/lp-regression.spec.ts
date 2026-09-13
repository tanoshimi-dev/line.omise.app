import { test, expect } from '@playwright/test'

// dev-plan-12-test-phase1 12.4: confirm dev-plan-08's LP migration and its
// SEO metadata are still intact.

test('homepage renders the LP sections with correct SEO metadata', async ({ page }) => {
  await page.goto('/')

  await expect(page).toHaveTitle('はんなりdev | LINE ミニアプリ開発 - 会員管理・予約・ECアプリ')
  await expect(page.locator('meta[name="description"]')).toHaveAttribute('content', /.+/)
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://line.omise.app')
  await expect(page.locator('meta[property="og:title"]')).toHaveAttribute('content', /.+/)
  await expect(page.locator('meta[property="og:type"]')).toHaveAttribute('content', 'website')
  await expect(page.locator('script[type="application/ld+json"]')).toHaveCount(1)

  // Section landmarks from dev-plan-08.
  await expect(page.locator('#demos')).toBeVisible()
  await expect(page.locator('#contact')).toBeVisible()
  await expect(page.locator('#profile')).toBeVisible()

  // Demo app external links (dev-plan-08 8.1).
  await expect(page.getByRole('link', { name: /デモサイトを見る/ }).first()).toHaveAttribute('href', /^https:\/\//)
})

test('robots.txt and sitemap.xml are served', async ({ page, request }) => {
  const robots = await request.get('/robots.txt')
  expect(robots.ok()).toBe(true)

  const sitemap = await request.get('/sitemap.xml')
  expect(sitemap.ok()).toBe(true)
  expect(await sitemap.text()).toContain('<urlset')
})
