import type { ArticleCategory } from './types'

// Split out from components/learn/ArticleListView.tsx (which transitively
// imports the `server-only`-marked serverApi.ts) so Client Components — the
// admin article pages (dev-plan-11-frontend-admin) — can use this label map
// without pulling a server-only module into the client bundle.
export const CATEGORY_LABELS: Record<ArticleCategory, string> = {
  'line-operation': 'LINE運用・設定',
  ai: '生成AI活用事例',
}
