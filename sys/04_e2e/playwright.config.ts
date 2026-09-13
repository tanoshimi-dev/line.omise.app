import { defineConfig, devices } from '@playwright/test'
import 'dotenv/config'

// dev-plan-12-test-phase1 12.3. Assumes the full stack is already running
// (docker compose up in sys/01_infra, or `next dev` + `go run ./cmd/server`
// locally) — this suite intentionally has no webServer block, since it needs
// both line-web AND line-api (plus Postgres) up together, which is exactly
// what docker-compose already orchestrates; duplicating that here would be
// redundant and slower.
export default defineConfig({
  testDir: './tests',
  fullyParallel: false, // tests share one dev database — avoid cross-test races
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: process.env.BASE_URL ?? 'http://localhost:3000',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
