# Step 03 — Go プロジェクト初期化・Gin セットアップ

**フェーズ:** Phase 1 — Foundation
**依存:** Step 02（DB）

---

## ゴール

`sys/02_backend/` に Go + Gin の API サーバーの土台を構築する。
`hannari.dev/cloudflare/log`（Chi + Huma 構成）とは異なり、本プロジェクトは Gin を採用する
（OpenAPI 自動生成は必須要件ではないため、シンプルな Gin ルーターで構成する）。

---

## タスク

### 3.1 プロジェクト初期化

- [ ] `sys/02_backend/go.mod` 初期化（Go 1.22+ 想定）
- [ ] Gin 導入（`github.com/gin-gonic/gin`）
- [ ] レイヤー構成を決定・ディレクトリ作成:

```
sys/02_backend/
├── cmd/
│   ├── server/main.go     # エントリーポイント
│   └── migrate/main.go    # マイグレーション CLI
├── internal/
│   ├── auth/              # LINE Login / Google OAuth トークン検証（Step 04）
│   ├── database/          # DB 接続プール
│   ├── handler/           # Gin ハンドラー
│   ├── middleware/        # CORS, 認証, ロギング
│   ├── repository/        # データアクセス層
│   ├── service/           # ビジネスロジック
│   └── testutil/          # テストヘルパー
├── migrations/
├── go.mod
└── Dockerfile
```

- [ ] `handler → service → repository` の3層構成を採用（`hannari.dev/cloudflare/log` の規約を踏襲）

### 3.2 共通ミドルウェア

- [ ] ロギングミドルウェア（`gin.Logger()` または構造化ログ）
- [ ] リカバリーミドルウェア（`gin.Recovery()`）
- [ ] CORS ミドルウェア（`line.omise.app` からのリクエストを許可）
- [ ] リクエスト ID / トレーシング（任意）

### 3.3 設定・環境変数読み込み

- [ ] `internal/config`（または同等）で環境変数を構造体にロード
- [ ] `DATABASE_URL`, `PORT`, `CORS_ALLOWED_ORIGINS` 等

### 3.4 ヘルスチェック

- [ ] `GET /health` — DB 接続確認を含むヘルスチェックエンドポイント

### 3.5 Dockerfile

- [ ] マルチステージビルドの `Dockerfile`（ビルド用 Go イメージ → 軽量な実行イメージ）

---

## 成果物

- `sys/02_backend/` 一式（上記ディレクトリ構成）
- `GET /health` が動作する最小限の Gin サーバー
- `sys/02_backend/Dockerfile`

## 完了条件

- `go run ./cmd/server` でサーバーが起動する
- `curl http://localhost:8080/health` が `200 OK` を返す
- Docker イメージがビルドできる
