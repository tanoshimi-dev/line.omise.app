# Step 08 実装結果 — 既存 LP（React+Vite）の Next.js への移行

**対応プラン:** [dev-plan-08-frontend-lp.md](../dev-plan-08-frontend-lp.md)

---

## 実施内容

### 8.1 コンポーネント移植

- `src/components/*`・`src/data/*.ts` は dev-plan-07-frontend-base の時点で既に新プロジェクトへ
  温存済みだったため、実質的な作業は「`app/(lp)/page.tsx` に配線する」「`'use client'` の要否を確認する」
  「外部リンク対応」の3点
- `src/app/(lp)/page.tsx` — プレースホルダーを置き換え、`HeroSection` → `DemoAppsSection` →
  `FeaturesSection` → `WhyUsSection` → `ContactSection` → `ProfileSection` の順で描画
  （`Header`/`Footer` は dev-plan-07 で既にルート `layout.tsx` に組み込み済みのため、
  ページ側では描画しない）
- `'use client'` の判別: `DemoAppCard.tsx` が `useState`/`useRef` を使うにもかかわらず
  ディレクティブが付いていなかった（Step07 の温存時点から未指定のまま）ため追加。
  他のセクションコンポーネントは全て hooks/イベントハンドラを持たない純粋な表示コンポーネントで、
  Server Component のままで問題ないことを確認
- `src/data/demoApps.ts` に `externalUrl` フィールドを追加（ユーザーからヒアリングした実URL）:
  - 会員管理: `https://membership.omise.app`
  - サロン予約: `https://salon-reservation.omise.app`
  - スイーツショップ: `https://sweets-shop.omise.app`
- `DemoAppCard.tsx` — カード下部に「デモサイトを見る」リンク（`target="_blank"`）を追加。
  既存のホバーで動画に切り替わる表示パターン（CLAUDE.md 記載の規約）はそのまま維持し、
  新しい要素として外部リンクボタンを追加する形にとどめた
- `src/App.tsx`・`src/index.css` を削除（Vite 時代のエントリポイント。`index.css` の内容は
  Step07 で既に `app/globals.css` に複製済みで完全に重複していた）

### 8.2 SEO メタデータ移行

- 旧 `index.html`（git 履歴の初回コミットから復元）の内容を Next.js Metadata API に移行:
  - `app/(lp)/page.tsx` — title, description, keywords, canonical, OGP, Twitter Card を
    ホームページ専用の `export const metadata` として定義（`/learn` 等は Step09〜11 で
    それぞれ独自の metadata を持つ想定のため、リッチな内容はレイアウトではなくページ側に置いた）
  - JSON-LD 構造化データ（Organization / WebSite / LocalBusiness）は `<script type="application/ld+json">`
    としてページ本体にそのまま出力
  - `app/layout.tsx` — サイト全体の既定値（`metadataBase`, タイトルテンプレート `%s | line.omise.app`,
    favicon）を設定
  - GA4（`gtag.js`, 測定ID `G-L6C568RP5H` は変更なし）を `next/script`（`strategy="afterInteractive"`）で
    ルートレイアウトに移行 — LP限定だった旧実装と異なり、サイト全体で計測されるようになった
  - Google Fonts（Noto Sans JP）を `<link>` タグ方式から **`next/font/google` に切り替え**
    （プラン外の判断・後述）
- `public/robots.txt` / `public/sitemap.xml` は Step07 の時点で既に `public/` に存在しており、
  変更不要（Next.js が自動的にそのまま配信することを確認）
- `/learn`, `/usecase` の動的 sitemap 化は **見送り**（プランは「検討」と明記のみ。
  両ルートの実コンテンツは Step09/10 未着手のため、実際のURL一覧が定まってから
  `app/sitemap.ts` への切り替えを検討する方が手戻りが少ないと判断）
- `doc/seo-update-2026-02-16.md` の検証チェックリストのうち、ローカルで検証可能な項目
  （ビルド成功、メタタグ出力、robots/sitemap配信）を再実施。Search Console 送信・
  Rich Results Test・OGP Debugger 等の外部ツールによる検証は本番ドメインへのデプロイ後
  （Step13）でないと実施できないため未実施（TODO はそのまま `doc/seo-update-2026-02-16.md` に残置）
- **`public/images/og-image.png` は依然未作成**（同ドキュメントの既存 TODO のまま。
  OGP/Twitter Card は画像URLを参照しているが、ファイル自体は本 Step のスコープ外のデザイン作業）

### 8.3 アセット移行

- `assets/`（gif/image/movie）と `public/images/*` は Step07 の時点でプロジェクト内に
  既に存在しており、追加の移植作業は不要だった。`assets/` はどのコンポーネントからも
  参照されていない未使用の元データ（`public/images/*` が実際に使われている最適化済みアセット）
  であることを確認 — 削除はせず現状維持（本 Step のタスクではないため）
- **`next/image` への置き換えは見送り**（検討した上で非採用）: `DemoAppCard` の
  「静止画像→ホバーで動画にクロスフェード」という表示は `next/image` に対応物がなく
  （動画側は結局プレーン `<video>` のまま）、画像側だけ置き換えると実装方式が混在するため、
  CLAUDE.md が明記する既存パターンをそのまま維持する方を優先した

### 8.4 お問い合わせセクション

- `ContactSection` の LINE 公式アカウント誘導は現状維持（プラン記載通り）。
  ただし実装を確認したところ、`ContactSection`/`HeroSection` の CTA はいずれも
  `siteContent.lineOaUrl` を直接使わず `#profile`（QRコードが実際に置かれている
  `ProfileSection`）へスクロールする設計になっていた。`siteContent.lineOaUrl` 自体は
  現状どのコンポーネントからも参照されていない（定義のみ）。これは Step08 が触れる前から
  存在した構成で、プランの「現状維持」という指示にも合致するため変更していない
  （挙動として破綻はしていない — CTA→QRコード表示という導線自体は成立している）

---

## プラン外で追加対応したこと

- **バグ修正（実装中に発見・Next.js 移行で顕在化）**: `ProfileSection.tsx` が
  `<p>` の子要素に `<div>` を2つ含む不正な HTML ネストになっており、Next.js の
  SSR+ハイドレーションで `Console Error: In HTML, <div> cannot be a descendant of <p>. This will
  cause a hydration error.` が発生していた（Vite の SPA では SSR 比較がないため症状が出ていなかった）。
  外側を `<div>`、内側を `<p>` に入れ替えて修正し、ブラウザで確認済み（修正前は dev overlay で
  3件の issue が報告され、うち2件がこのハイドレーションエラー起因だった）
- **Google Fonts を next/font/google に切り替え**: 旧 `index.html` の `<link rel="preconnect">` +
  Google Fonts スタイルシート方式は `app/layout.tsx` に単純移植することもできたが、
  Next.js の標準的な自己ホスト方式（外部リクエスト削減・レイアウトシフト防止）に切り替えた。
  `globals.css` の `--font-sans` と `body { font-family }` を next/font が生成する
  CSS 変数 `--font-noto-sans-jp` 参照に更新
- ビルド時に判明したタイトルの二重サフィックス問題: `layout.tsx` にタイトルテンプレート
  （`%s | line.omise.app`）を設定した結果、`page.tsx` の完成済みタイトル文字列にまで
  テンプレートが適用され「...ECアプリ | line.omise.app」という意図しない表示になっていた。
  `page.tsx` 側で `title: { absolute: title }` を使いテンプレートを無効化して解消

## プラン未実施 / 保留

- `public/images/og-image.png` の作成（デザイン作業、既存 TODO）
- Search Console 登録・サイトマップ送信・Rich Results Test / OGP Debugger（本番デプロイ後、Step13）
- `/learn`, `/usecase` の動的 sitemap 化（Step09/10 でコンテンツ確定後に判断）

---

## 検証

- `npm run build` — 成功（TypeScript エラーなし、7ルートすべて静的プリレンダリング）
- ビルド出力 `.next/server/app/index.html` を直接確認:
  - `<title>`, `<meta name="description">`, `<meta name="keywords">`, `<link rel="canonical">`,
    OGP（`og:title/description/url/image/type/site_name/locale`）, Twitter Card,
    JSON-LD（Organization/WebSite/LocalBusiness）, favicon — すべて期待通り出力されていることを確認
  - 他ページ（`/learn` 等）は独自 metadata 未設定のため layout のデフォルトタイトル
    `line.omise.app` にフォールバックすることを確認（テンプレートは子が指定した場合のみ適用される）
- `npm run dev` + Claude in Chrome でブラウザ実描画を確認:
  - 初回は dev overlay が「3 Issues」を報告 → 上記の `<div> in <p>` バグを修正して「0 Issues」に
  - Hero / DemoApps（3カードの外部リンクボタン含む） / Features / WhyUs / Contact
    （QRコード誘導） / Profile（修正後の文章表示） / Footer をスクロールしながら目視確認、
    レイアウト崩れなし
  - デモアプリ3件の外部リンク（`membership.omise.app` 等）が正しい href で出力されていることを確認
  - 画像・動画（`/images/*.png`, `*.mp4`）がすべて `200` で配信されることを確認
- **検証中に判明した無関係な問題**: 検証序盤、`docker compose` で稼働し続けていた
  Step07 時点の `line-web` コンテナがポート3000を占有しており、`npm run dev` が
  ポート3001にフォールバック → バックエンドの `CORS_ALLOWED_ORIGINS`（3000のみ許可）に
  弾かれて `/auth/me` が `Failed to fetch` になった。本 Step のバグではなく検証環境側の
  ポート競合と判明したため、`docker compose stop line-web` でポートを解放して解消
  （最終確認は正しいポート3000で実施）
- `docker compose up -d --build line-web` — 新しいビルドで再構築し、`/`, `/robots.txt`,
  `/sitemap.xml`, `/images/membership.png` がいずれも `200` を返すことを確認
- `next lint` は本バージョンの Next.js では動作しない（`next lint` コマンド自体が
  ディレクトリ引数として `lint` を誤認識する既知の非互換 — dev-plan-07 時点からの
  ツーリング上の既存ギャップであり本 Step の対象外。`npm run build` の型チェックのみで代替）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| `https://line.omise.app/` で現行と同等の LP が表示される | ✅（ローカル環境で確認。本番デプロイは Step13） |
| 各ミニアプリカードから対応する外部ドメインへ遷移できる | ✅（3件とも正しい href で出力を確認） |
| `npm run build` の出力で主要メタタグが確認できる | ✅ |

## 次のステップ

- `dev-plan-09-frontend-learn.md`（`/learn/` — 本 Step で確立した「ページごとに metadata を持つ」
  パターンを踏襲できる）
- `dev-plan-10-frontend-usecase.md`
- `public/images/og-image.png` の作成（引き続きユーザー側の TODO）
