# Step 08 — 既存 LP（React+Vite）の Next.js への移行

**フェーズ:** Phase 1 — Frontend
**依存:** Step 07（Next.js 基盤）

---

## ゴール

現行の React + Vite ランディングページ（`Header`, `HeroSection`, `DemoAppsSection`,
`FeaturesSection`, `WhyUsSection`, `ContactSection`, `ProfileSection`, `Footer`）を
Next.js に移植する。3ミニアプリは別ドメインで稼働中のため、紹介セクションは
「外部リンク付きの実績紹介」に位置づけを変更する。

---

## タスク

### 8.1 コンポーネント移植

- [ ] `src/components/*` を Next.js プロジェクトへ移植（`'use client'` が必要な箇所を判別）
- [ ] `src/data/siteContent.ts`, `src/data/demoApps.ts` をそのまま移植（データ駆動の構造を維持）
- [ ] `DemoAppsSection` / `DemoAppCard`: 各ミニアプリの遷移先を「別ドメインで稼働中の実アプリ URL」への
  外部リンクに変更（`demoApps.ts` に `externalUrl` フィールドを追加）

### 8.2 SEO メタデータ移行

- [ ] `index.html` に直書きされていたメタタグ（title, description, OGP, Twitter Card, JSON-LD,
  GA4）を Next.js の `metadata` API（`app/layout.tsx` / `app/page.tsx` の `export const metadata`）
  に移行
- [ ] `public/robots.txt`, `public/sitemap.xml` を Next.js の `public/` にそのまま配置
  （`/learn/`, `/usecase/` の動的ページも `sitemap.xml` に含める方式を検討 — 静的ファイル vs
  `app/sitemap.ts` での動的生成）
- [ ] `doc/seo-update-2026-02-16.md` の検証チェックリストを再実施

### 8.3 アセット移行

- [ ] `assets/`（gif, image, movie）, `public/images/*` を新プロジェクトに移植
- [ ] Next.js `<Image>` コンポーネントへの置き換えを検討（最適化のメリットと動画/GIFの扱いを確認）

### 8.4 お問い合わせセクション

- [ ] `ContactSection` の LINE 公式アカウント誘導（`siteContent.lineOaUrl`）は現状維持

---

## 成果物

- Next.js 上で動作する LP（既存デザイン・コンテンツを踏襲）
- 更新済み `demoApps.ts`（外部リンク対応）
- 移行済み SEO メタデータ

## 完了条件

- `https://line.omise.app/` で現行と同等の LP が表示される
- 各ミニアプリカードから対応する外部ドメインへ遷移できる
- `npm run build` の `dist/`（Next.js の場合はビルド出力）で主要メタタグが確認できる
