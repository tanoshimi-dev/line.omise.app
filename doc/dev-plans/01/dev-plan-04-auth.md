# Step 04 — LINE Login / Google OAuth 統合・Admin/Reader ロール

**フェーズ:** Phase 1 — Foundation
**依存:** Step 03（バックエンド基盤）

---

## ゴール

LINE Login と Google OAuth によるログインを実装し、バックエンドでのトークン検証・
セッション発行・ユーザー管理（Admin/Reader ロール判定）の仕組みを構築する。
Firebase 等の仲介 IdP は使わず、各プロバイダーと直接連携する。

---

## タスク

### 4.1 プロバイダー設定

- [ ] LINE Developers コンソールで LINE Login チャネルを作成
  - コールバック URL: `https://line.omise.app/auth/line/callback`（開発用は localhost も登録）
  - スコープ: `profile`, `openid`, （必要なら `email`）
- [ ] Google Cloud Console で OAuth 2.0 クライアントを作成
  - 承認済みリダイレクト URI: `https://line.omise.app/auth/google/callback`
- [ ] クライアント ID / シークレットは `sys/secrets/` またはサーバー環境変数として管理（コミットしない）

### 4.2 認可コードフロー（バックエンド検証方式を決定）

- [ ] フローを決定: Next.js（フロント）が認可コードを受け取り → バックエンドにコードを渡す →
  バックエンドがプロバイダーとトークン交換 → ユーザー情報取得、という構成を基本方針とする
- [ ] `internal/auth/line.go` — LINE Login トークン交換・ID トークン検証（JWKS）
- [ ] `internal/auth/google.go` — Google OAuth トークン交換・ID トークン検証

```go
type ProviderClaims struct {
    Provider       string // "line" | "google"
    ProviderUserID string
    Email          string
    DisplayName    string
    AvatarURL      string
}
```

### 4.3 セッション方式（未決定事項 — ここで決定する）

- [ ] Cookie ベースのサーバーサイドセッション vs JWT（ステートレス）のどちらを採用するか決定
  - 推奨: HttpOnly Secure Cookie + サーバー側セッションストア（シンプルで失効制御がしやすい）
- [ ] セッションストアの実装（DB テーブル `sessions` or 別ストア）

### 4.4 認証ミドルウェア

- [ ] `internal/middleware/auth.go`
  - Reader 認証: ログイン必須エンドポイント用（試験受験・進捗保存など）
  - Admin 認証: `role = admin` のユーザーのみ許可するミドルウェア（コンテンツ CRUD 用）
  - 未認証時は `401 Unauthorized`、権限不足時は `403 Forbidden`

### 4.5 ユーザー登録（upsert）・Admin 昇格

- [ ] `internal/repository/user.go`
  - `UpsertByProvider(ctx, provider, providerUserID, email, displayName, avatarURL)`
  - 初回ログイン時に `users` へ自動登録（デフォルト `role = reader`）
- [ ] Admin 昇格方法を決定・実装
  - 例: 環境変数 `ADMIN_EMAILS`（カンマ区切り）に一致するメールアドレスでログインした際に `role = admin` を自動付与
  - もしくは運用者が DB を直接更新（初期は手動でも可）

### 4.6 認証 API エンドポイント

- [ ] `GET /auth/line/login` — LINE Login 認可 URL へリダイレクト
- [ ] `GET /auth/line/callback` — コールバック処理・ユーザー upsert・セッション発行
- [ ] `GET /auth/google/login` / `GET /auth/google/callback` — 同様
- [ ] `POST /auth/logout` — セッション破棄
- [ ] `GET /auth/me` — ログイン中ユーザー情報取得（`role` を含む）

### 4.7 環境変数

```bash
LINE_LOGIN_CHANNEL_ID=
LINE_LOGIN_CHANNEL_SECRET=
LINE_LOGIN_CALLBACK_URL=https://line.omise.app/auth/line/callback

GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
GOOGLE_OAUTH_CALLBACK_URL=https://line.omise.app/auth/google/callback

ADMIN_EMAILS=owner@example.com
SESSION_SECRET=
```

---

## 成果物

- `internal/auth/line.go`, `internal/auth/google.go`
- `internal/handler/auth.go`
- `internal/repository/user.go`
- `internal/middleware/auth.go`（Reader / Admin 認証）
- LINE Developers / Google Cloud Console 設定完了

## 完了条件

- LINE / Google それぞれでログインでき、`users` テーブルにレコードが作成される
- `ADMIN_EMAILS` に一致するユーザーが `role = admin` になる
- Admin 限定エンドポイントに Reader ユーザーがアクセスすると `403` が返る
- 未ログインで保護エンドポイントにアクセスすると `401` が返る
