# Step 03 実装結果 — Go プロジェクト初期化・Gin セットアップ

**対応プラン:** [dev-plan-03-backend-base.md](dev-plan-03-backend-base.md)

---

## 実施内容

### 3.1 プロジェクト初期化

- `go mod init github.com/tanoshimi-dev/line.omise.app/sys/02_backend`（Go 1.22 を `go.mod` に指定。
  `go mod tidy` が依存関係の要求により `1.25.0` に引き上げた — 開発機のツールチェーンは 1.26.1 のため
  問題なくビルドできる。Dockerfile 側は `golang:1.25-alpine` に合わせた）
- `github.com/gin-gonic/gin` を追加
- プラン記載のディレクトリ構成を作成:
  `cmd/server`, `cmd/migrate`, `internal/{auth,config,database,handler,middleware,repository,service,testutil}`, `migrations/`
  - `internal/auth`, `internal/repository`, `internal/service`, `internal/testutil` は Step 04〜06 まで空
  - `internal/config` はプランになかったが、環境変数ロード（タスク 3.3）の置き場所として追加

### 3.2 共通ミドルウェア

- `internal/middleware/cors.go` — 許可オリジンのみ `Access-Control-Allow-*` を返す CORS ミドルウェア
- ロギング / リカバリーは Gin 標準の `gin.Logger()` / `gin.Recovery()` を `cmd/server/main.go` で使用

### 3.3 設定・環境変数読み込み

- `internal/config/config.go` — `PORT` / `DATABASE_URL` / `CORS_ALLOWED_ORIGINS` を読み込み、
  `PORT` はデフォルト `8080`

### 3.4 ヘルスチェック

- `GET /health` — `internal/handler/health.go`
  - プロセスの生存確認に加え、`internal/database` パッケージで DB 到達性を best-effort 報告
  - **プラン外の判断**: Step 02（DB エンジン決定）が未着手のため、実際の DB ドライバは使わず
    `DATABASE_URL` の host:port へ TCP 接続を試みるだけのスタブとした
    (`configured` / `reachable` / `detail` を返す)。`DATABASE_URL` 未設定でも `/health` は `200 OK` を返す
    — Step 02 実装時にこのスタブを実ドライバ（`database/sql` 等）に置き換える

### 3.5 Dockerfile

- マルチステージビルド（`golang:1.25-alpine` → `alpine:3.20`）
- 非 root ユーザー（`appuser`）で実行

---

## 検証

- `go vet ./...` / `go build ./...` — 成功
- `docker build` — 成功（イメージサイズ確認は未実施）
- コンテナ単体起動 + `curl /health` → `200 OK`、`database.configured=false`（`.env` なし）
- `docker compose up -d --build line-api`（`sys/01_infra/docker-compose.yml`）→ `curl localhost:8080/health`
  → `200 OK`、`database.configured=true, reachable=false`（`DATABASE_URL` はサンプル値のみで実 DB なし、想定通り）
- 検証用に作成した `.env`・Docker イメージ・コンテナはすべて後片付け済み（コミット対象なし）

---

## プラン外で追加対応したこと

- `internal/config` パッケージを追加（プランのディレクトリ構成表になかったが、3.3 のタスクの置き場所として必要だったため）
- `cmd/migrate/main.go` はプランの成果物に明記がなかったが、ディレクトリ構成表に含まれていたため
  「未実装」の旨をログ出力するだけのプレースホルダーとして作成（Step 02 で実装）

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- 実際の DB ドライバ・コネクションプール（Step 02 の DB エンジン決定後に `internal/database` を置き換え）
- 認証・CRUD 系ハンドラー（Step 04〜06）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| `go run ./cmd/server` でサーバーが起動する | ✅ |
| `curl http://localhost:8080/health` が `200 OK` を返す | ✅（standalone / docker compose 両方で確認） |
| Docker イメージがビルドできる | ✅ |

## 次のステップ

- `dev-plan-02-database.md`（DB エンジン決定 → `internal/database` を実ドライバに置き換え）
- `dev-plan-04-auth.md`（LINE Login / Google OAuth）
