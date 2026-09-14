# Step 12 実装結果 — Phase 1 テスト

**対応プラン:** [dev-plan-12-test-phase1.md](../dev-plan-12-test-phase1.md)

---

## 実施内容

### アーキテクチャ上の決定（実装に先立って必要だった判断）

- **DB統合テストの方式: `testcontainers-go` を採用**（プランの「検討」に対する決定）。
  実機の Docker（Docker Desktop / Windows）上で実際に動作することを事前にスモークテストで確認した上で採用。
  各パッケージの最初のテストでコンテナを1つ起動し（`sync.Once`）、以降のテストはそのプールを
  共有して `TRUNCATE ... RESTART IDENTITY CASCADE` で毎テスト後にクリーンアップする方式にした
  （テストごとにコンテナを起動すると約15秒/テストかかり非現実的なため）
  - Windows の Docker Desktop 環境では、testcontainers が「コンテナ起動完了」と報告した直後に
    Postgres へ接続すると `EOF` になる既知のタイミング問題があったため、
    マイグレーション適用に最大10回・500ms間隔のリトライを実装（`internal/testutil/db.go`）
- **ハンドラーテストの方式: リポジトリのインターフェース化・モック化はしない**。
  既存コードはリポジトリを具象構造体のまま Handler に注入する設計（意図的にシンプルに保たれている）。
  これをテストのためだけにインターフェース化するのは本 Step の目的に対してやりすぎと判断し、
  代わりに `cmd/server/main.go` のルーティング構築ロジックを `internal/server.New()` として
  切り出し、テストからも本番と全く同じルーター（実際の認証ミドルウェアチェーンを含む）を
  構築できるようにした。ハンドラーテストは実質的に「本物のDBに対する結合テスト」になるが、
  401/403 の挙動を含めミドルウェアチェーンごと検証できる利点を優先した
- **E2E でのログイン方式: 実 OAuth は行わず、セッションを直接シードする**（プランの
  「テスト用アカウント or モックプロバイダー方針を決定」に対する決定）。
  LINE/Google の認可コード交換自体は dev-plan-04-auth で実アカウントによる手動検証済み、かつ
  Go のハンドラーテストでもカバーされているため、E2E で再現する価値は低い一方、
  CI で安定して繰り返し実行できる実アカウントやモックプロバイダーを用意するコストは高いと判断した。
  `sys/04_e2e/tests/helpers/db.ts` が `internal/session/cookie.go` と同じ HMAC-SHA256 署名を
  Node の `crypto` で再現し、Postgres に直接ユーザー+セッション行を作成してブラウザに
  Cookie を注入する。ログインボタン自体が正しいプロバイダーの認可URLへ遷移することだけは
  別途 E2E で確認する（`login-redirect.spec.ts`）

### 12.1 Go テスト基盤

- `internal/testutil/`
  - `db.go` — `TestDB(t)`（共有 testcontainers Postgres プール、マイグレーション適用込み）、
    `TruncateAll(t, pool)`
  - `auth.go` — `CreateUser(t, pool, role)`、`LoginCookieValue(t, pool, userID)`
    （署名付き Cookie 値を生成 — 手動検証で使っていた `curl -H "Cookie: ..."` パターンをそのまま関数化）
  - `server.go` — `NewRouter(t, pool)`（`internal/server.New` を試験用DBで構築）、
    `DoRequest(t, router, method, path, cookie, body)`（`httptest` によるインプロセスHTTPリクエスト）
- `internal/database` に `NewPoolForTest` を追加（既存の `pgxpool.Pool` を `*database.Pool` へラップ —
  テストが `Connect`（DSN文字列前提）を経由せずに済むようにするための最小限のテスト用エクスポート）
- Repository層テスト（`internal/repository/*_test.go`、27件）: 公開判定フィルタ（course/article/usecase の
  `ListPublished` が draft を除外）、Admin昇格・非降格（`UpsertByProvider`）、NULL許容カラムのスキャン
  （dev-plan-04で見つけたバグの回帰テスト）、試験の1レッスン1試験制約、設問+選択肢のトランザクション整合性、
  進捗のupsert・他ユーザー分離・試験結果の履歴保持（最新のみ返す）等
- Service層テスト（`internal/service/*_test.go`、10件）: `ValidateStatus`/`ValidateArticleCategory`/
  `ValidateRelatedDemoApp`、`GradeExam`（全問正解・部分正解・未回答・不正な設問/選択肢ID・
  設問0件・合格点の境界値を含む7ケース）
- Session層テスト（`internal/session/cookie_test.go`、5件）: 署名の往復・改ざん検知・秘密鍵不一致検知
- Handler層テスト（`internal/handler/*_test.go`、12件）: **dev-plan-05 で見つけた `RequireAdmin`
  認可バイパスバグの回帰テスト**（Reader が 403 を受け取り、かつ実際に DB へ書き込みが
  発生していないことを確認）、Admin/Reader/未ログインの 401/403/200/201 系統、重複スラッグの 409、
  試験の `is_correct` が Reader 向けレスポンスに一切含まれないことの直接検証、採点結果の検証
- カバレッジ: `go test ./... -coverprofile=coverage.out`。`internal/repository` 63.1%、
  `internal/service` 73.7%、`internal/session` 93.8%、`internal/handler` 21.9%
  （多数あるハンドラーのうち代表的な経路のみを検証したため）。`internal/auth`・`internal/config`・
  `internal/database`・`internal/middleware`・`cmd/*` は直接のテストファイルを持たないため
  0% と表示されるが、`internal/middleware` は全ハンドラーテストの認証チェーンを通じて
  実質的に検証されている（`go test` の per-package カバレッジ集計の性質上、数値には出ない）

### 12.2 Next.js テスト

- Vitest + React Testing Library を導入（`vitest.config.ts`、`@/*` エイリアスを `tsconfig.json` と合わせた）
- `src/lib/api.test.ts` — `fetch` をモックし、`get`/`post`/`put`/`delete` のメソッド・Cookie送信・
  `ApiError` の送出を検証
- `src/lib/markdown.test.ts` — `excerpt()`（Markdown記法の除去・切り詰め）
- `src/components/learn/LessonInteractive.test.tsx` — プランが例示する「試験受験フォーム」
  「進捗表示」の両方をカバー: 未ログイン時のログイン誘導、完了マークのクリック→状態遷移、
  試験の開始→回答選択→送信→採点結果表示（`api.post` に送られる実際のペイロードまで検証）
- 記事一覧のタグ絞り込み（Server Component）は Vitest では直接カバーしていない
  （後述「プラン未実施」参照）— 同機能は E2E（Playwright）で実ブラウザ・実APIに対して検証済み

### 12.3 E2E テスト（Playwright）

- `sys/04_e2e/`（CLAUDE.md で予約されていたディレクトリ）に独立した Node プロジェクトとして構築
- `tests/helpers/db.ts` / `auth.ts` — 上記のセッション直接シード方式の実装
- `tests/login-redirect.spec.ts` — LINE/Googleログインボタン→各プロバイダーの認可エンドポイントへの遷移
- `tests/public-content.spec.ts` — `/learn/` トップ・講座受講導線・記事一覧のタグ絞り込み
  （テストが自分でタグを1件シードし、終了後に削除）・記事詳細・`/usecase/` 一覧・詳細・404系
- `tests/auth-and-progress.spec.ts` — レッスン完了→進捗バッジ反映、マイページの進捗表示、
  未ログイン/Reader/Adminそれぞれの `/admin` アクセス制御
- `tests/admin-content.spec.ts` — Admin が事例を新規作成（下書き）→非公開確認→公開に変更→
  公開ページに反映→UI から削除→非公開ページ・一覧の両方から消えることを確認
  （講座ではなく事例を使った理由: `/learn/` トップは固定3カテゴリカードで動的な講座一覧ページが
  存在しないため、動的な公開一覧を持つ `/usecase/` の方が「作成→公開→公開ページに表示」の
  検証に適している）
- `tests/lp-regression.spec.ts` — 12.4 のLP回帰確認（後述）
- 全17件のテストは共有の開発用DBに対して実行され、作成したテストデータ（ユーザー・セッション・
  タグ・事例）はすべて `finally` ブロックで削除する設計にした（実行後にDBを検証しseedデータ・
  実ログインユーザーが無事であることを確認済み）

### 12.4 回帰確認

- `lp-regression.spec.ts` — title/description/canonical/OGP/JSON-LD、LPの各セクション
  （デモ・お問い合わせ・プロフィール）、デモアプリの外部リンク、`robots.txt`/`sitemap.xml` の配信
- 未ログイン状態でのアクセス制御は `auth-and-progress.spec.ts` の Access control グループと
  Go の Handler テスト（401/403系）の両方でカバー

---

## プラン外で追加対応したこと

- `cmd/server/main.go` のルーティング構築を `internal/server.New()` に抽出（テスト基盤のための
  リファクタ。`main.go` 自体の挙動は変更なし、`go build`/`docker build` で動作確認済み）
- `internal/database.NewPoolForTest` を追加（テスト専用のプールラッパー）
- ルートの `.gitignore` に `coverage.out` を追加

## プラン未実施

- 記事一覧のタグ絞り込みの Vitest（コンポーネント）テスト — `ArticleListView` は非同期の
  Server Component で `serverApi`（`server-only`）に依存するため、Vitest + RTL でのユニットテストの
  相性が悪い（モックは可能だが、Server Component の非同期関数を直接呼び出す形になり実利用時の
  レンダリング経路と乖離する）。同機能は Playwright E2E（`public-content.spec.ts`）で
  実ブラウザ・実APIに対して検証済みのため、Vitest 側は見送った
- Go の `internal/auth`（LINE/Google クライアント）・`internal/config`・`cmd/*` の直接テストは
  未実装（前者は外部サービスへの実HTTP呼び出しが前提でモックが必要になり、後者は薄い
  設定読み込み/エントリポイントであるため、本 Step の優先度からは見送った）

---

## 検証

- `go vet ./...` / `go build ./...` / `gofmt -l .`（差分なし）— 成功
- `go test ./...` — **全パッケージ成功**（テストのあるパッケージはすべて `ok`、
  テストのないパッケージは `FAIL` ではなく `coverage: 0.0%` 表示のみで正常）
- `npm run build`（Next.js）— 成功
- `npm test`（Vitest）— **15件全て成功**
- `npx playwright test`（`sys/04_e2e`）— **17件全て成功**（Docker Compose で全スタックを
  起動した状態で実行。実行後、テストが作成したデータがすべて削除され、seed データと
  実ログインユーザーが無傷であることを確認済み）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| `go test ./...` が全て成功する | ✅ |
| `npm test`（Vitest）が全て成功する | ✅ |
| `npx playwright test` の主要シナリオが成功する | ✅ |

## 次のステップ

- `dev-plan-13-deploy-phase1.md`（本番 VPS デプロイ）
- 任意の改善候補（未実施、必要になれば別途）: Handler層カバレッジの拡充、
  `internal/auth` プロバイダークライアントのHTTPモックテスト、CI（GitHub Actions等）への
  この3種のテストスイートの組み込み
