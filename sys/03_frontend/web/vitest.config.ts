import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'node:path'

const rootDir = import.meta.dirname

// dev-plan-12-test-phase1 12.2. Kept separate from next.config.ts —
// Next.js has no first-party Vitest integration, and this project doesn't
// use Next's own test runner, so a standalone Vite config (mirroring the
// `@/*` alias from tsconfig.json) is the standard setup.
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(rootDir, './src'),
    },
  },
})
