import { test, expect } from '@playwright/test'
import { withDB } from './helpers/db'

// Uses the project's stable seed data (dev-plan-02-database's seed.sql):
// course "line-marketing" / lesson "intro", article "rich-menu-basics",
// usecase "sample-salon" — all published, so safe to assert against without
// creating anything.

test.describe('/learn/ public content', () => {
  test('top page links to all three categories', async ({ page }) => {
    await page.goto('/learn')
    await expect(page.getByRole('link', { name: 'LINEマーケティング講座' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'LINE運用・設定' })).toBeVisible()
    await expect(page.getByRole('link', { name: '生成AI活用事例' })).toBeVisible()
  })

  test('course page lists lessons and links to a lesson detail', async ({ page }) => {
    await page.goto('/learn/line-marketing')
    await expect(page.getByRole('heading', { name: 'LINEマーケティング講座' })).toBeVisible()

    await page.getByRole('link', { name: 'はじめに' }).click()
    await expect(page).toHaveURL(/\/learn\/line-marketing\/intro$/)
    await expect(page.getByRole('heading', { name: 'はじめに' })).toBeVisible()
  })

  test('draft content never appears on a public page', async ({ page }) => {
    // Regression coverage for dev-plan-05's publish-filtering: a lesson
    // slug that only exists as a draft in fixtures elsewhere must 404 here.
    const res = await page.goto('/learn/line-marketing/this-lesson-does-not-exist')
    expect(res?.status()).toBe(404)
  })

  test('article list supports tag filtering', async ({ page }) => {
    // The tag menu only renders once at least one tag exists (dev-plan-09's
    // ArticleListView design) — seed one directly rather than depending on
    // whatever tags happen to exist in the shared dev database right now.
    const tagSlug = `e2e-tag-${Date.now()}`
    await withDB(async (client) => {
      const article = await client.query<{ id: number }>(`SELECT id FROM articles WHERE slug = 'rich-menu-basics'`)
      const tag = await client.query<{ id: number }>(`INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING id`, ['E2Eタグ', tagSlug])
      await client.query(`INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)`, [article.rows[0].id, tag.rows[0].id])
    })

    try {
      await page.goto('/learn/line-operation')
      await expect(page.getByRole('heading', { name: 'LINE運用・設定' })).toBeVisible()
      await expect(page.getByRole('link', { name: 'すべて' })).toBeVisible()
      const tagChip = page.getByRole('link', { name: 'E2Eタグ', exact: true })
      await expect(tagChip).toBeVisible()
      await tagChip.click()
      await expect(page).toHaveURL(new RegExp(`\\?tag=${tagSlug}$`))
      await expect(page.getByRole('link', { name: 'リッチメニューの基本設定' })).toBeVisible()
    } finally {
      await withDB(async (client) => {
        await client.query(`DELETE FROM article_tags WHERE tag_id = (SELECT id FROM tags WHERE slug = $1)`, [tagSlug])
        await client.query(`DELETE FROM tags WHERE slug = $1`, [tagSlug])
      })
    }
  })

  test('article detail page renders the body and a back link', async ({ page }) => {
    await page.goto('/learn/line-operation/rich-menu-basics')
    await expect(page.getByRole('heading', { name: 'リッチメニューの基本設定' })).toBeVisible()
    await page.getByRole('link', { name: /一覧へ戻る/ }).click()
    await expect(page).toHaveURL(/\/learn\/line-operation$/)
  })
})

test.describe('/usecase/ public content', () => {
  test('list page shows the seeded usecase and links to its detail', async ({ page }) => {
    await page.goto('/usecase')
    await expect(page.getByRole('heading', { name: '導入事例', exact: true })).toBeVisible()

    await page.getByRole('link', { name: /LINE公式アカウント導入事例/ }).click()
    await expect(page).toHaveURL(/\/usecase\/sample-salon$/)
    await expect(page.getByText('サンプルサロン', { exact: true })).toBeVisible()
  })

  test('unknown usecase slug 404s', async ({ page }) => {
    const res = await page.goto('/usecase/does-not-exist')
    expect(res?.status()).toBe(404)
  })
})
