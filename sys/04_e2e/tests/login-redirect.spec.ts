import { test, expect } from '@playwright/test'

// dev-plan-12-test-phase1 12.3 decision: we don't drive a real LINE/Google
// OAuth handshake in E2E (see tests/helpers/db.ts for why) — this just
// proves the login buttons on the homepage take you to the right place.
// The authorization-code exchange itself is covered by dev-plan-04-auth's
// manual verification (real accounts) and by the Go handler tests.

test('LINE login button redirects to LINE\'s authorization endpoint', async ({ page }) => {
  await page.goto('/')
  const [response] = await Promise.all([
    page.waitForResponse((res) => res.url().includes('access.line.me'), { timeout: 10000 }).catch(() => null),
    page.getByRole('link', { name: 'LINEでログイン' }).click(),
  ])
  // Either we observed the redirect hit LINE's server, or (offline/blocked)
  // at minimum the browser navigated away to the expected host.
  if (!response) {
    await expect(page).toHaveURL(/access\.line\.me/)
  }
})

test('Google login button redirects to Google\'s authorization endpoint', async ({ page }) => {
  await page.goto('/')
  const [response] = await Promise.all([
    page.waitForResponse((res) => res.url().includes('accounts.google.com'), { timeout: 10000 }).catch(() => null),
    page.getByRole('link', { name: 'Googleでログイン' }).click(),
  ])
  if (!response) {
    await expect(page).toHaveURL(/accounts\.google\.com/)
  }
})
