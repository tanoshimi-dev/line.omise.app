# Step 10 実装結果 — `/usecase/` 一覧・詳細 UI

**対応プラン:** [dev-plan-10-frontend-usecase.md](../dev-plan-10-frontend-usecase.md)

---

## 実施内容

### スキーマ拡張（実装に必須だった判断）

プラン 10.1/10.2 は「サムネイル」表示と「関連ミニアプリへのリンク（があれば）」を要求しているが、
`usecases` テーブル（dev-plan-02-database）にはどちらの列も存在しなかった。
`migrations/006_usecase_thumbnail.up.sql`／`.down.sql` を追加:

- `thumbnail_url TEXT`（nullable）
- `related_demo_app TEXT`（nullable、`CHECK (... IN ('membership', 'salon-reservation', 'sweets-shop'))`
  — `src/data/demoApps.ts` の `DemoApp.id` と一致させる想定の閉じた列挙）

`internal/repository/usecase.go`（Create/Update/Get系すべて）・`internal/handler/usecase.go`
（リクエスト/レスポンス）・`internal/service/content.go`（`ValidateRelatedDemoApp`）を対応させて更新。
空文字は SQL `NULL` に変換して保存する（`nullIfEmpty` ヘルパー、CHECK制約が空文字を弾かないように）。

### 10.1 一覧ページ — `/usecase/`

- `src/app/usecase/page.tsx` — `GET /api/usecases` から取得し、店舗名・タイトル・サムネイルの
  カード一覧を表示

### 10.2 詳細ページ — `/usecase/{client}/`

- `src/app/usecase/[slug]/page.tsx` — `GET /api/usecases/:slug` から取得
- 本文レンダリングは dev-plan-09 で作った `src/components/learn/Markdown.tsx` を
  そのまま再利用（プラン記載の「共通コンポーネント化を検討」に対する回答 — 検討の結果、
  新規コンポーネントを作らずそのまま共有する形にした）
- 関連ミニアプリへのリンク: `related_demo_app` が設定されていれば
  `src/data/demoApps.ts`（dev-plan-08-frontend-lp で `externalUrl` を追加済み）から
  該当アプリを検索し、外部リンクボタンとして表示。未設定なら何も表示しない

### 10.3 SEO

- 一覧・詳細の両方に `generateMetadata`（詳細ページは店舗名を含むタイトル、
  本文からの抜粋を description に使用）
- `src/app/sitemap.ts`（dev-plan-09 で作成）に `/usecase` と公開中の `/usecase/:slug` を追加

---

## プラン外で追加対応したこと・プランからの逸脱

- **「静的生成寄り」（プラン冒頭のゴール）を字義通りには採用しなかった**: 当初
  `next: { revalidate: 3600 }`（ISR）で実装したが、`npm run build` 時に
  `/usecase` の静的プリレンダリングが試行され、ビルド時点では `line-api` に
  到達できず（`docker build` で `line-web` イメージを作る際、`docker-compose` の
  内部ネットワークは存在しないため）ビルドが失敗することが判明した。
  ISR は「ビルド時に一度成功させ、その後バックグラウンドで再生成する」方式のため、
  バックエンド依存のコンテンツを本プロジェクトのビルド構成でISR化するには
  フォールバックデータ等の追加実装が必要になり、複雑さに見合わないと判断。
  dev-plan-09 の `/learn/` と同じ `cache: 'no-store'`（常に動的・常に最新）に統一した。
  `src/lib/serverApi.ts` には経緯をコメントで残している
- 上記の検証中に見つかった知見: `serverApi.get` に一時的に `revalidate` オプションを
  追加する実装を試みたが、結果的に不採用として撤回。現在の `serverApi.ts` は
  dev-plan-09 時点のシンプルな実装のまま

## プラン未実施

なし。10.1〜10.3 すべて実装・検証済み。

---

## 検証

- `go build ./...` / `go vet ./...` / `gofmt -l .`（差分なし）— バックエンド成功
- マイグレーション: `go run ./cmd/migrate up` → `version=6`
- `npm run build` — 成功（`revalidate` 版は失敗することを確認した上で `no-store` に戻して再成功）
- Admin API で既存の seed usecase（`sample-salon`）に `thumbnail_url`（サロン予約アプリの
  スクリーンショット）と `related_demo_app: salon-reservation` を設定 → 正しく保存・返却されることを確認
- 不正な `related_demo_app`（スキーマの CHECK 制約にない値）を指定した作成リクエスト → `400`
- ブラウザ（Claude in Chrome）で実描画を確認:
  - `/usecase` — サムネイル付きカード一覧
  - `/usecase/sample-salon` — 店舗名・タイトル・サムネイル・Markdown本文・
    「サロン予約アプリのデモサイトを見る」外部リンクボタン（href が
    `https://salon-reservation.omise.app/` と正しく一致することを確認）
  - `/usecase/nonexistent` → `404`（ヘッダー等レイアウトは維持されたまま Next.js の
    not-found ページが表示されることを確認）
- `docker compose up -d --build line-web` で再構築し、`/usecase`, `/usecase/sample-salon`,
  `/sitemap.xml`（`/usecase` 系エントリ含む）がコンテナ経由で `200` を返すことを確認
- 検証で作成したテストセッションはすべて削除済み。`sample-salon` に設定した
  サムネイル/関連アプリ情報は実データとして有用なため保持（テスト用の使い捨てデータではない）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 事例一覧・詳細がコンテンツ API から取得して表示される | ✅ |
| 未ログインで問題なく閲覧できる | ✅（認証を一切要求しないページ構成） |

**Phase 1 — Frontend（Step 07〜10）完了。**

## 次のステップ

- `dev-plan-11-frontend-admin.md`（管理画面 — Step05/06/10 で実装した Admin API
  （courses/lessons/articles/tags/usecases/exams）を操作するフォームUIとして実装）
