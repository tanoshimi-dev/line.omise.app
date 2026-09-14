# Step 04 実装結果 — LINE Login / Google OAuth 統合・Admin/Reader ロール

**対応プラン:** [dev-plan-04-auth.md](../dev-plan-04-auth.md)

> **追記（Step 05 実装時に判明）**: この Step で実装した `internal/middleware.RequireAdmin` に、
> Reader が Admin 限定 API を実行できてしまう認可バイパスのバグがあった
> （`RequireReader` を関数として直接呼び出したことによる Gin ハンドラーチェーンの意図しない先行実行）。
> 本 Step の検証時は応答本文を書かない簡易ルートでしかテストしておらず、このバグを再現できていなかった。
> 詳細と修正内容は [dev-plan-05-content-api-result.md](dev-plan-05-content-api-result.md) を参照。

---

## 実施内容

### 4.1 プロバイダー設定 — ユーザー側で完了・実ログイン確認済み

LINE Developers コンソール / Google Cloud Console でのチャネル・OAuth クライアント作成は
本セッションの権限では実施できないため（ユーザー自身のアカウントでの手動作業）、
コード側は環境変数を読み込むだけの構成にして受け渡した。ユーザー側で以下を実施し、
LINE・Google 両方のログインが実際に成功することを確認済み:

- LINE Developers コンソールで LINE Login チャネルを作成、Channel ID / Secret を取得
- Google Cloud Console で OAuth クライアントを作成
  - 途中「Authorized redirect URIs」に開発用コールバック URL
    （`http://localhost:8080/auth/google/callback`）が未登録で `redirect_uri_mismatch`
    が発生したが、登録後に解消
- 発行された値を `sys/02_backend/.env` の `LINE_LOGIN_CHANNEL_ID` / `LINE_LOGIN_CHANNEL_SECRET` /
  `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET` に設定

### 4.2 認可コードフロー

- `internal/auth/provider.go` — `ProviderClaims` / `Provider` インターフェース
- `internal/auth/line.go` — LINE Login
  - 認可 URL 生成 → `/oauth2/v2.1/token` でコード交換 → **LINE の `/oauth2/v2.1/verify` エンドポイントで
    ID トークン検証**（JWKS を自前で取得して署名検証する代わりに、LINE 公式が提供する
    サーバーサイド検証エンドポイントを利用。`nonce` も渡して LINE 側にリプレイをチェックさせる）
- `internal/auth/google.go` — Google OAuth
  - 認可 URL 生成 → `/token` でコード交換 → **Google の `tokeninfo` エンドポイントで ID トークン検証**
    （同様に JWKS ローカル検証を避けた簡略化。`aud`・`nonce` はこちらで突き合わせて確認）
  - **プラン外の判断（簡略化）**: 本番運用でアクセス数が増えた場合、LINE/Google 双方とも
    JWKS ベースのローカル署名検証への切り替えを検討する余地あり（現状は依存ライブラリを
    増やさない方針を優先）

### 4.3 セッション方式（決定済み）

- **決定: Cookie ベースのサーバーサイドセッション**（プラン推奨どおり）
- `sys/02_backend/migrations/004_sessions.up.sql`／`.down.sql` — `sessions` テーブル
  （`id TEXT PRIMARY KEY` に暗号学的乱数トークンを使用、`user_id` FK、`expires_at`、TTL 30日）
- `internal/repository/session.go` — `Create` / `GetValid` / `Delete`
- `internal/session/cookie.go` — Cookie に入れる値は `<session_id>.<HMAC-SHA256>`
  （`SESSION_SECRET` で署名。DB に到達する前に改ざんされた Cookie を弾けるようにするための追加防御で、
  セッション ID 自体は乱数なので署名がなくても安全だが、`SESSION_SECRET` をプランの環境変数リストに
  合わせて実際に使う設計とした）

### 4.4 認証ミドルウェア

- `internal/middleware/auth.go`
  - `RequireReader` — Cookie検証 → セッション有効性確認 → ユーザー取得 → Gin コンテキストに格納。
    いずれかの失敗で `401`
  - `RequireAdmin` — `RequireReader` を通した上で `role != admin` なら `403`
  - `CurrentUser(c)` — ハンドラー側でログイン中ユーザーを取得するヘルパー

### 4.5 ユーザー登録（upsert）・Admin 昇格

- `internal/repository/user.go` — `UpsertByProvider`
  - **Admin 昇格方法（決定済み）**: 環境変数 `ADMIN_EMAILS`（カンマ区切り、大文字小文字無視）に
    一致するメールアドレスでログインした際に自動的に `role = admin` を付与
  - **降格はしない**方針: 既存ユーザーが `ADMIN_EMAILS` に含まれなくなっても、ログイン時に
    自動で `reader` に戻すことはしない（誤って env から消した場合に意図せず権限を失わないため）

### 4.6 認証 API エンドポイント

`cmd/server/main.go` に実装・登録:

- `GET /auth/line/login` / `GET /auth/line/callback`
- `GET /auth/google/login` / `GET /auth/google/callback`
- `POST /auth/logout`
- `GET /auth/me`（`RequireReader` 経由）

Login → state(CSRF) と nonce(リプレイ対策) を乱数生成し `line_omise_oauth_state` Cookie
（`Path=/auth`, 10分, HttpOnly, SameSite=Lax）に格納してプロバイダーへリダイレクト。
Callback → state 照合 → コード交換 → user upsert → session 発行 → `line_omise_session` Cookie
（HttpOnly, SameSite=Lax, `Secure` は `APP_ENV=production` の時のみ）をセットして
`FRONTEND_URL/auth/callback` へリダイレクト（dev-plan-07-frontend-base で実装済みのページに合流する）。

### 4.7 環境変数

`internal/config/config.go` に追加。`.env.example` / `.env` を更新（プレースホルダー値、
`.env` は開発用ダミー値 + `openssl rand -base64 32` で生成した `SESSION_SECRET`）。

---

## プラン外で追加対応したこと

- **`FRONTEND_URL` 環境変数を追加**（プランの 4.7 リストにはなかった）。
  dev-plan-07-frontend-base で `/auth/callback` ページが既に「バックエンドがログイン後に
  ここへリダイレクトしてくる」前提で実装されていたため、バックエンド側がフロントエンドの
  オリジンを知る必要があった。デフォルト `http://localhost:3000`
- **`APP_ENV` 環境変数を追加**。セッション Cookie の `Secure` フラグを本番(HTTPS)でのみ
  有効にするため（開発は `http://localhost` なので `Secure` を強制すると Cookie が保存されない）
- **`cmd/server/main.go` に `DATABASE_URL` 未設定時の起動失敗チェックを追加**。
  dev-plan-02-database までは DB 未設定でも `/health` が動く設計だったが、認証機能は
  DB（users/sessions テーブル）なしには成立しないため、`SESSION_SECRET` 同様に起動時点で
  `log.Fatal` するよう変更（DB 到達不能 = 起動失敗、DB 未設定 = 従来通り起動、という
  従来の使い分けは維持: `DATABASE_URL` が空文字の場合のみ Fatal、値はあるが疎通できない場合は
  これまで通り `/health` が `reachable=false` を返すのみで起動は継続する）
- **バグ修正（実装中に発見）**: `users` テーブルの `email` / `display_name` / `avatar_url` は
  NULL 許容カラムだが、`internal/repository/user.go` の Scan は素の `string` に直接読み込んでおり、
  値が NULL の行（例: `migrations/seed.sql` の admin ユーザーは `avatar_url` 未設定）で
  `cannot scan NULL into *string` エラーになっていた。SELECT/RETURNING 側で
  `COALESCE(col, '')` を使うよう修正（動作確認中に実際に踏んで発見・修正した）

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

なし。4.1〜4.7 すべて完了・実ログインで確認済み。

---

## 検証

- `go vet ./...` / `go build ./...` / `gofmt -l .`（差分なし）— 成功
- `docker compose up -d --build line-api` → `/health` 引き続き `200 OK`
- マイグレーション: `go run ./cmd/migrate up` → `version=4`（`sessions` テーブル追加を確認）
- **未ログイン確認**: `GET /auth/me` → `401`、`POST /auth/logout`（セッションなし）→ `204`
- **ログインリダイレクト確認**: `GET /auth/line/login` / `GET /auth/google/login` →
  それぞれ正しい認可URL（`client_id`, `redirect_uri`, `state`, `nonce`, `scope` 含む）へ `302`。
  `line_omise_oauth_state` Cookie が `HttpOnly; SameSite=Lax; Path=/auth; Max-Age=600` で発行される
- **CSRF state 不一致確認**: 発行された Cookie と異なる `state` で `/auth/line/callback` を叩くと `400`
- **プロバイダー交換失敗確認**: 正しい `state`・偽の `code` で callback を叩くと、
  LINE の実サーバーへの交換リクエストが失敗し `401`（想定通りのエラーハンドリング）
- **ログイン成功後の状態を手動再現して確認**（実コンソールがないため、DB に直接セッション行を
  投入して代替検証）:
  - `psql` で `sessions` テーブルに admin ユーザーのセッションを直接 INSERT し、
    `internal/session.Sign` と同じ HMAC 署名を付けた Cookie で `GET /auth/me` → `200`、
    フロントエンドの `CurrentUser` 型と一致する JSON
    （`{"id":"1","provider":"google","email":"admin@example.com","display_name":"Dev Admin","avatar_url":"","role":"admin"}`）
  - `POST /auth/logout` → `204`、DB の該当セッション行が削除されることを確認
  - ログアウト後に同じ Cookie で `/auth/me` → `401`
  - reader ロールのテストユーザー・セッションを追加投入し、一時的な `RequireAdmin` 検証用ルートで
    reader → `403`、admin → `200`、Cookie なし → `401` を確認。検証用ルート・テストデータは
    確認後にコードとDBの両方から削除済み（コミット対象なし）
- 検証中に使用した一時テストデータ（`test-session-id-*`, `dev-reader` ユーザー等）はすべて
  DB から削除済み。`docker compose` スタック自体はユーザーが Adminer で確認作業中のため稼働継続
- **実ログイン E2E 確認（ユーザー実施）**: LINE ログイン・Google ログインの両方をブラウザから
  実際に実行し、成功を確認。`users` テーブルに実プロバイダー経由のレコードが作成されたことを確認済み
  （`provider=line, role=reader` 1件・`provider=google, role=reader` 1件 — 事前に投入していた
  `provider=google, role=admin` の seed ユーザーとは別に新規作成されている。ログインしたメールアドレスは
  `ADMIN_EMAILS` と一致しなかったため、想定通り `role=reader` で登録された）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| LINE / Google それぞれでログインでき、`users` テーブルにレコードが作成される | ✅ ユーザーが実ログインで確認（両プロバイダーとも `users` に新規レコード作成を確認） |
| `ADMIN_EMAILS` に一致するユーザーが `role = admin` になる | ✅（ロジック実装済み。実ログインしたアカウントは `ADMIN_EMAILS` 対象外だったため `role=reader` で登録され、これも想定通りの動作として確認） |
| Admin 限定エンドポイントに Reader ユーザーがアクセスすると `403` が返る | ✅（一時検証ルートで確認） |
| 未ログインで保護エンドポイントにアクセスすると `401` が返る | ✅（`/auth/me` で確認） |

**Step 04 完了。**

## 次のステップ

- `dev-plan-05-content-api.md` / `dev-plan-06-exam-progress-api.md`
  （`internal/repository`・`internal/middleware`（Reader/Admin）を使ったコンテンツ CRUD・試験/進捗 API）
