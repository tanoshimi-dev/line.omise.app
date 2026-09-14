# Step 02 実装結果 — DB 選定・スキーマ・マイグレーション

**対応プラン:** [dev-plan-02-database.md](dev-plan-02-database.md)

---

## 実施内容

### 2.1 DB エンジン・マイグレーションツール

- DB エンジン: **PostgreSQL**（`postgres:16-alpine`）
- マイグレーションツール: **`golang-migrate/migrate/v4`**（`database/postgres` + `source/file` ドライバ）
- `sys/01_infra/docker-compose.yml` に `postgres` サービスを追加
  - `line-omise-internal` ネットワークのみ、`5432:5432` を公開（ローカルから `go run ./cmd/migrate` 等で直接接続できるように）
  - `postgres-data` 名前付きボリュームで永続化
  - `pg_isready` ヘルスチェックを追加し、`line-api` の `depends_on` を `condition: service_healthy` に変更
- `sys/01_infra/docker-compose.prod.yml` にも `postgres` サービスを追加
  - `line-omise-web`（Traefik）ネットワークには参加させない（DB は HTTP ルーティング不要）
  - 認証情報はコミットしない `../secrets/postgres.env`（`env_file`）から読む方針とし、サービス定義内にコメントで明記
- `sys/02_backend/.env.example` / `.env` の `DATABASE_URL` を実際の PostgreSQL 接続文字列に更新
  （`postgres://line_omise:line_omise@postgres:5432/line_omise_app?sslmode=disable`。
  `docker compose up` 実行時は `postgres` サービス名で解決、ホスト直接実行時は `localhost` に読み替える旨をコメントで明記）

### 2.2〜2.5 スキーマ

プラン記載のテーブルをそのまま実装（カラム構成・UNIQUE制約もプラン通り）:

- `sys/02_backend/migrations/001_users.up.sql`／`.down.sql` — `users`
- `sys/02_backend/migrations/002_content.up.sql`／`.down.sql` — `courses`, `lessons`, `articles`, `tags`, `article_tags`, `usecases`
- `sys/02_backend/migrations/003_exam_progress.up.sql`／`.down.sql` — `exams`, `exam_questions`, `exam_choices`, `user_lesson_progress`, `user_exam_results`

主キーは `BIGSERIAL`、タイムスタンプ系は `TIMESTAMPTZ DEFAULT now()`、`status` / `provider` / `role` / `category` は `CHECK` 制約で値を制限。外部キーは全て `ON DELETE CASCADE`。

Admin 権限の付与方法（2.2 の未決定事項）は引き続き Step 04 で確定（本 Step では触れず）。

### 2.6 マイグレーション基盤

- `sys/02_backend/cmd/migrate/main.go` を実装（プレースホルダーを置き換え）
  - `golang-migrate` を薄くラップした CLI: `go run ./cmd/migrate <up|down|version>`（`-path` フラグでマイグレーションディレクトリを変更可能、デフォルト `migrations`）
  - `DATABASE_URL` 未設定時は起動前にエラー終了
- `sys/02_backend/Makefile` を新規作成: `make migrate-up` / `migrate-down` / `migrate-version` / `migrate-seed`
  （`migrate-seed` は `docker compose exec postgres psql` 経由で `migrations/seed.sql` を投入）

### 2.7 開発用シード

- `sys/02_backend/migrations/seed.sql` — 開発用ダミーデータ（admin ユーザー1件、講座+レッスン1件、記事1件、事例1件）
  - golang-migrate のバージョン管理対象外の素の SQL とし、`ON CONFLICT ... DO NOTHING` で再実行可能にした

### アプリ側の DB 接続

- `internal/database/db.go` を TCP 到達性のみを見るスタブから、`pgxpool.Pool` を使った実コネクションプールに置き換え
  - `Connect(ctx, databaseURL)` — `DATABASE_URL` 未設定なら「configured=false」の空 `Pool` を返す（起動は失敗させない）
  - `Check(ctx, pool)` — 2秒タイムアウト付き `Ping` で `configured`/`reachable`/`detail` を返す（`/health` の応答形式は変更なし）
- `internal/handler/health.go` — 引数を `databaseURL string` から `*database.Pool` に変更
- `cmd/server/main.go` — 起動時に `database.Connect` でプールを作成し、`defer dbPool.Close()`

---

## プラン外で追加対応したこと

- `sys/01_infra/docker-compose.yml` に `adminer`（`adminer:4`）サービスを追加。
  DB テーブルを GUI で確認したいというユーザー要望への対応（開発用途のみ、`docker-compose.prod.yml` には追加していない —
  認証なしの DB 管理 UI を本番で公開しない判断）。`http://localhost:8081` で接続、
  System: PostgreSQL / Server: `postgres` / User: `line_omise` / Password: `line_omise` / DB: `line_omise_app`

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- Admin 権限付与方法の最終決定（2.2 内の未決定事項） — Step 04 で確定
- `internal/repository` を使った実際のクエリ層 — Step 05〜06 で実装（本 Step はスキーマとコネクションプールの用意まで）

---

## 検証

- `go vet ./...` / `go build ./...` — 成功
- `go mod tidy` — `golang-migrate/migrate/v4`・`jackc/pgx/v5` 系の依存関係が正しく解決されることを確認（`go.sum` 更新、docker/testcontainers 等の不要な間接依存は残っていないことを `go.mod` で確認済み）
- `docker compose -f sys/01_infra/docker-compose.yml up -d --build`
  - `postgres` が healthy になった後 `line-api` が起動することを確認
  - `curl localhost:8080/health` → `200 OK`、`database.configured=true, reachable=true`（マイグレーション適用前でも接続自体は成功）
- `DATABASE_URL=postgres://line_omise:line_omise@localhost:5432/line_omise_app?sslmode=disable go run ./cmd/migrate up` → `migrate: up complete`
  - `go run ./cmd/migrate version` → `version=3 dirty=false`
  - `docker compose exec postgres psql \dt` → 13テーブル（`schema_migrations` 含む）を確認
- シード投入（`migrations/seed.sql` を `psql` に直接投入して `make migrate-seed` と同等の動作を確認）→ 各テーブルに1件ずつ想定通り挿入されたことを確認
- `go run ./cmd/migrate up`（適用済み状態での再実行）→ `ErrNoChange` を正常系として扱い `migrate: up complete` を出力することを確認
- `go run ./cmd/migrate down` → 全テーブルロールバック → `go run ./cmd/migrate version` → `no migration`（想定通りのエラー終了）
- 再度 `go run ./cmd/migrate up` → `version=3` に復帰することを確認（up/down の往復が安全であることを確認）
- `docker compose -f docker-compose.prod.yml config -q` — `../secrets/postgres.env` 不在時はエラーになることを確認（想定通り、本番運用時にユーザーが用意する）。一時的にダミーの `sys/secrets/postgres.env` を作成して YAML 構文自体は妥当であることを確認後、削除済み（コミット対象なし）
- 検証用に起動したコンテナ・ボリューム・イメージ（`line-api`, `line-web`, `postgres`）はすべて `docker compose down -v` + `docker rmi` で後片付け済み

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| ローカルで DB コンテナが起動し、マイグレーションが適用できる | ✅ |
| `users` / `courses` / `lessons` / `articles` / `usecases` / `exams` 系テーブルが作成される | ✅ |
| Step 05〜06 のコンテンツ・試験・進捗 API がこのスキーマに対して実装できる状態になっている | ✅ スキーマ・接続プールとも用意済み（クエリ層は Step 05〜06 で実装） |

## 次のステップ

- `dev-plan-04-auth.md`（LINE Login / Google OAuth・Admin/Reader ロール — `users` テーブルを使用）
- `dev-plan-05-content-api.md` / `dev-plan-06-exam-progress-api.md`（本 Step のスキーマに対する CRUD・API 実装）
