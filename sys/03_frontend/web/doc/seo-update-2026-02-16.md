# SEO改善: 「はんなりdev」検索1位を目指す

## 背景

- `line.omise.app` のランディングページが Google に全くインデックスされていない
- キーワード「はんなりdev」で検索しても結果に表示されない
- 競合はほぼゼロのため、技術的SEO対策を行えば1位は十分達成可能

## 実施済み (2026-02-16)

### 1. `index.html` — メタタグ・OGP・構造化データ追加

- `<title>` に「はんなりdev」を追加
- `<meta name="description">` に「はんなりdev」を追加
- `<meta name="keywords">` 追加（はんなりdev, LINE ミニアプリ, etc.）
- `<link rel="canonical">` 追加 → `https://line.omise.app/`
- OGP (Open Graph) タグ追加
  - `og:title`, `og:description`, `og:url`, `og:image`, `og:type`, `og:site_name`, `og:locale`
- Twitter Card タグ追加（summary_large_image）
- JSON-LD 構造化データ追加
  - `Organization` — はんなりdev の基本情報
  - `WebSite` — サイト名・URL
  - `LocalBusiness` — 地域ビジネス情報（京都府）

### 2. `public/robots.txt` — 新規作成

```
User-agent: *
Allow: /
Sitemap: https://line.omise.app/sitemap.xml
```

### 3. `public/sitemap.xml` — 新規作成

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://line.omise.app/</loc>
    <lastmod>2026-02-16</lastmod>
    <changefreq>monthly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>
```

### 4. `src/data/siteContent.ts` — キーワード最適化

- `hero.headline`: 「はんなりdev」を先頭に追加
- `hero.subtitle`: 「はんなりdevは、...」に変更
- h1タグに「はんなりdev」が自動反映（HeroSection.tsx が siteContent を参照）

## TODO（手動作業）

### 必須

- [ ] **OG画像を作成**: `public/images/og-image.png` (1200x630px 推奨)
  - SNS共有時のサムネイルになる
  - 「はんなりdev」のロゴ・サービス概要を含めたデザイン
- [ ] **Google Search Console** に `line.omise.app` を登録
  - https://search.google.com/search-console
  - DNS TXT レコードまたは HTML ファイルで所有権を確認
- [ ] **サイトマップを送信**: Search Console の「サイトマップ」から `https://line.omise.app/sitemap.xml` を送信
- [ ] **URL のインデックス登録をリクエスト**: Search Console の「URL検査」から `https://line.omise.app/` を送信

### 推奨

- [ ] **被リンクの作成**: SNS やブログなどから `line.omise.app` へのリンクを作成
  - X (Twitter) プロフィールにURL掲載
  - Qiita / Zenn の技術記事からリンク
  - GitHub リポジトリの README にリンク
- [ ] **Google ビジネスプロフィール** に登録（LocalBusiness 構造化データとの整合性向上）
- [ ] **ページ速度の監視**: PageSpeed Insights で定期的にチェック
  - https://pagespeed.web.dev/

## 検証方法

1. `npm run build` でビルドしてエラーがないことを確認
2. `dist/index.html` に全メタタグが正しく出力されていることを確認
3. `dist/robots.txt` と `dist/sitemap.xml` が存在することを確認
4. [Rich Results Test](https://search.google.com/test/rich-results) で構造化データを検証
5. [OGP Debugger](https://developers.facebook.com/tools/debug/) で OGP タグを検証

## 効果測定（目安）

| 期間 | 期待される状態 |
|------|---------------|
| 1週間後 | Google にインデックスされる |
| 2-4週間後 | 「はんなりdev」で検索結果に表示 |
| 1-2ヶ月後 | 「はんなりdev」で検索1位 |
