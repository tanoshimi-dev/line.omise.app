import type { BrowserContext } from '@playwright/test'
import type { TestUser } from './db'

/** Injects a seeded session cookie into a browser context, simulating a completed login. */
export async function loginAs(context: BrowserContext, baseURL: string, user: TestUser): Promise<void> {
  const url = new URL(baseURL)
  await context.addCookies([
    {
      name: 'line_omise_session',
      value: user.cookie,
      domain: url.hostname,
      path: '/',
      httpOnly: true,
      secure: false,
    },
  ])
}
