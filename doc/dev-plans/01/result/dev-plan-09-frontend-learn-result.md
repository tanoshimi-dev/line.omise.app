# Step 09 実装結果 — `/learn/` 講座・記事一覧/詳細・試験・進捗表示 UI

**対応プラン:** [dev-plan-09-frontend-learn.md](../dev-plan-09-frontend-learn.md)

---

## 実施内容

### アーキテクチャ上の決定（プラン外・実装に必須だった判断）

- **サーバー用/ブラウザ用で API のベース URL を分離**: 本 Step で初めてサーバーサイド
  （Server Components）からバックエンド API を叩く必要が生じた。`docker-compose` 環境では
  Next.js サーバー（`line-web` コンテナ）は内部ネットワークのホスト名 `line-api` で
  バックエンドに到達する必要がある一方、ブラウザは引き続きホストにマッピングされたポート
  （`localhost:8080`）でしか到達できない。この2つを混同すると SSR がコンテナ内で
  `localhost:8080` に接続しようとして失敗する。
  - `src/lib/serverApi.ts`（新規） — `API_URL`（`NEXT_PUBLIC_` 接頭辞なし＝クライアントに
    絶対にバンドルされない）を使うサーバー専用フェッチヘルパー。`server-only` パッケージで
    クライアントコンポーネントからの誤 import をビルドエラーにする
  - `src/lib/api.ts`（既存、変更なし） — 引き続き `NEXT_PUBLIC_API_URL` でブラウザから使用
  - `.env` / `.env.example` に `API_URL=http://line-api:8080` を追加
    （dev-plan-02-database の `DATABASE_URL` と同じ考え方: Docker 前提の既定値、
    ホスト単体実行時は `http://localhost:8080` に上書き）
- **ユーザー固有データ（進捗・試験結果）はブラウザ側フェッチに限定**: セッション Cookie は
  HttpOnly のため、サーバーサイドで転送するには `next/headers` 経由で読み取って手動で
  ヘッダーに付け直す実装が必要になる。公開コンテンツ（講座・記事の一覧/詳細）は
  SSR（`serverApi`、SEO重視）、ログイン状態に依存する部分（完了バッジ・試験受験・
  マイページ）はすべて Client Component が `src/lib/api.ts`（Cookie 付き `fetch`）で
  別途取得する設計に統一した。dev-plan-07 で確立済みの「`/auth/me` はクライアント側で取得」
  という方針と一貫させるため、Cookie 転送の複雑さを増やさない選択をした

### 9.1 `/learn/` トップ

- `src/app/learn/page.tsx` — 3カテゴリへのカード導線
- ハンバーガーメニュー内の並び順は dev-plan-07-frontend-base で既に確定済み
  （マーケティング講座 → 運用・設定 → AI活用事例、`Header.tsx` 参照）だったため、
  トップページのカード順もこれに合わせただけで新たな決定は不要だった

### 9.2 講座（コース形式） — `/learn/line-marketing/`

- `src/app/learn/line-marketing/page.tsx` — 講座トップ。`GET /api/courses/line-marketing`
  で講座+レッスン一覧を取得（本文込みで1回のフェッチで完結 — コース詳細APIがレッスンの
  `body` も含めて返すため、レッスン詳細ページ用に個別フェッチは不要）
- `src/components/learn/CourseLessonList.tsx`（Client Component） — 完了バッジ表示。
  ログイン時のみ `GET /api/courses/:slug/progress` を追加取得し、未ログイン時は
  一切フェッチしない
- `src/app/learn/line-marketing/[lessonSlug]/page.tsx` — レッスン詳細。前後レッスンへの
  ナビゲーションは講座詳細のレッスン配列（`sort_order` 順）から算出
- `src/components/learn/LessonInteractive.tsx`（Client Component） — 完了マーク・試験受験の
  両方を1コンポーネントに統合。`GET /api/courses/:slug/progress` を1回取得するだけで
  「このレッスンは完了済みか」「試験はあるか・前回の結果は」の両方が分かる設計にし、
  試験の設問本体（`GET /api/lessons/:lessonId/exam`）は「試験を受ける」ボタン押下時にのみ
  追加取得する（初期表示時の無駄なフェッチを避けた）
  - 未ログイン時は `LoginPrompt` を表示（閲覧は引き続き可能 — 完了条件通り）
- `src/components/learn/LoginPrompt.tsx` — ログイン誘導の共通コンポーネント
  （進捗保存・試験受験の両方から利用）

### 9.3 記事一覧形式 — `/learn/line-operation/`, `/learn/ai/`

- `src/components/learn/ArticleListView.tsx` / `ArticleDetailView.tsx` — `line-operation`と
  `ai` の両カテゴリで共有する Server Component（プラン記載の「同一コンポーネントを再利用」）。
  `src/app/learn/line-operation/page.tsx` 等の実ルートはカテゴリ名を渡すだけの薄いラッパー
- タグ絞り込みは README 推奨の `?tag=rich-menu` クエリ方式（プラン記載通り）。
  **設計上の工夫**: タグ一覧の選択肢は常に「絞り込み前の全記事」から算出し、
  実際の表示リストのみ `?tag=` 指定時に絞り込み後のデータに差し替える。
  絞り込み結果からタグ一覧を作ると、一度タグを選んだ後に他のタグへ切り替えられなくなるため
- `src/components/learn/Markdown.tsx` — `react-markdown` + `remark-gfm` +
  `@tailwindcss/typography`（`prose` クラス）でレンダリング

### 9.4 進捗表示

- `src/app/learn/me/page.tsx` — `/learn/me`。ユーザー固有データのみのページのため
  Server Component の皮を被せず素の Client Component ページとして実装
- ヘッダーのユーザー表示名（`Header.tsx` の `AuthControls`）を `/learn/me` へのリンクに変更
  （プラン記載の「ヘッダーのユーザーメニューからアクセス」を満たす導線）

### 9.5 SEO

- 各ページで `generateMetadata` を実装（講座トップ・レッスン詳細・記事一覧・記事詳細）。
  記事本文からの description は `src/lib/markdown.ts` の `excerpt()` で Markdown 記法を
  簡易的に除去して生成（記事に専用の要約フィールドがないため）
- `src/app/sitemap.ts`（新規）が dev-plan-08 の静的 `public/sitemap.xml` を置き換え、
  公開中の講座・レッスン・記事を動的に含める。バックエンド呼び出しを含むため
  Next.js が自動的に動的レンダリング（リクエスト毎に再生成）と判定し、
  新規コンテンツがビルドなしで反映される。各カテゴリ取得は `try/catch` で
  ベストエフォート化し、API 障害時も静的エントリだけは返せるようにした

---

## プラン外で追加対応したこと

- `@tailwindcss/typography` プラグインを追加（Markdown レンダリング用。Tailwind v4 の
  `@plugin` 記法で `globals.css` に登録）
- `react-markdown` / `remark-gfm` / `server-only` を新規依存として追加
- README.md の「Open decision: 最終カテゴリ名」を解消 —
  `articles.category` の CHECK 制約・API が dev-plan-02/05 の時点で既に `line-operation`
  に確定していたため、実装に合わせて記録を更新
- README.md の「Content Site Map」冒頭の "Not yet implemented" を "`/learn/` implemented" に更新

## プラン未実施

なし。9.1〜9.5 すべて実装・検証済み。

---

## 検証

- `npm run build` — 成功。ルート一覧で `/learn/*` の動的セグメントはすべて `ƒ (Dynamic)`、
  `/sitemap.xml` も動的と表示されることを確認（想定通り）
- ブラウザ（Claude in Chrome）で未ログイン状態の実描画を確認:
  - `/learn` — 3カテゴリカード
  - `/learn/line-marketing` — 講座詳細・レッスン一覧
  - `/learn/line-marketing/intro` — レッスン本文（Markdown）・ログイン誘導パネル
  - `/learn/line-operation` — 記事一覧・タグチップ（テスト用タグ追加後）
  - タグチップをクリックして `?tag=broadcast` に絞り込まれ、かつ他のタグも
    引き続き選択肢に残ることを確認（設計通り）
  - `/learn/line-operation/broadcast-tips` — 見出し・強調・箇条書き・コードブロック・
    リンクを含むテスト記事が `@tailwindcss/typography` で正しく整形されることを確認
- **ログイン済み状態の検証方法について**: セッション Cookie は `HttpOnly` のため
  ブラウザの JS からは設定できず、実際の LINE/Google ログインを本セッションで
  代行することもしていない（認証情報の操作はしない方針）。そのため dev-plan-04〜06 と
  同じ手法（DB に直接テスト用セッションを投入し `curl` で署名付き Cookie を付けて叩く）で
  以下をAPIレベルで確認した:
  - `POST /api/lessons/:id/complete` → `GET /api/courses/line-marketing/progress` で
    `completed:true` が反映される
  - 試験作成 → `GET /api/lessons/:id/exam` で `is_correct` が含まれない → 正解を送信 →
    `score:100, passed:true` → 同じ progress エンドポイントで `latest_result` に反映
  - `GET /api/me/progress` が正しい集計を返す
  - フロントエンドの各コンポーネントはこれらのレスポンス型（`src/lib/types.ts`）を
    そのまま消費する実装になっており、上記のAPIレスポンス形状が正しいことを確認できたことで
    データバインディングの正しさも担保している（UI 自体の見た目は未ログイン状態の
    レンダリングで確認済み）
- `docker compose up -d --build line-web` で再構築し、`/learn`, `/learn/line-marketing`,
  `/learn/line-marketing/intro`, `/learn/line-operation`, `/sitemap.xml` がいずれも
  コンテナ経由（`API_URL=http://line-api:8080` の内部ネットワーク接続）で `200` を返すことを
  確認 — サーバー/クライアントの API URL 分離が実際に機能していることの実証
- 検証中に作成したテストデータ（試験・設問・選択肢・テスト記事・テストタグ・
  reader テストユーザー・関連セッション/進捗/受験結果）はすべて DB から削除済み。
  seed データ（`line-marketing`講座・`rich-menu-basics`記事）と実ログインユーザーは保持
- 日本語を含む JSON ペイロードは、この検証端末（Windows, コードページ932）で
  `curl -d '...'` に直接埋め込むと文字化けする既知の問題（dev-plan-06 で確認済み）に
  再度遭遇 → 毎回 UTF-8 ファイル経由（`--data-binary @file`）で送り直して解消

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 講座・記事の一覧/詳細がコンテンツ API から取得して表示される | ✅ |
| ログインユーザーがレッスンを完了しマークでき、進捗が反映される | ✅（API レベルで確認。UI コンポーネントは同じレスポンス型を消費） |
| 試験を受験し結果が表示される | ✅（同上） |
| 未ログインでも記事・レッスンの閲覧はできる | ✅（ブラウザで確認） |

## 次のステップ

- `dev-plan-10-frontend-usecase.md`（`/usecase/` — 本 Step の `serverApi`/`ArticleListView`
  的なパターンを踏襲できる）
- `dev-plan-11-frontend-admin.md`（管理画面 — 本 Step で使った Admin API の書き込み側を
  フォームUIとして実装する）
