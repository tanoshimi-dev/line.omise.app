# Step 05 実装結果 — コンテンツ CMS API

**対応プラン:** [dev-plan-05-content-api.md](../dev-plan-05-content-api.md)

---

## 実施内容

### 5.1 講座・レッスン API

- `internal/repository/course.go` — `CourseRepository`（courses/lessons 両方を担当）
  - `ListPublished` / `GetPublishedBySlug` / `GetBySlug`（admin用、status問わず）/ `GetByID`
  - `Create` / `Update` / `Delete`（courses）、`CreateLesson` / `UpdateLesson` / `DeleteLesson` / `GetLessonBySlug` / `GetLessonByID` / `ListLessonsByCourse`
  - 公開系メソッドは SQL の `WHERE status = 'published'` で絞り込み、`draft` は決して返さない
- `internal/handler/course.go` — `GET /api/courses`, `GET /api/courses/:slug`（レッスン一覧を同梱）,
  `GET /api/courses/:slug/lessons/:lessonSlug`, および Admin 用 CRUD（courses/lessons）

### 5.2 記事 API

- `internal/repository/article.go` — `ArticleRepository`
  - `ListPublished(ctx, category, tagSlug)` — `tagSlug` 指定時は `article_tags`/`tags` を JOIN
  - `GetPublishedByCategoryAndSlug` / `GetByID` / `Create` / `Update` / `Delete`
  - `GetOrCreateTag`（`ON CONFLICT (slug) DO UPDATE`）/ `AttachTag`（`ON CONFLICT DO NOTHING`）/ `ListTagsForArticle`
- `internal/handler/article.go` — `GET /api/articles?category=&tag=`, `GET /api/articles/:category/:slug`,
  Admin CRUD、`POST /api/admin/articles/:id/tags`（タグ作成+付与を1回で行う）

### 5.3 導入事例 API

- `internal/repository/usecase.go` / `internal/handler/usecase.go` — courses と同型の CRUD
  （`client_name` を持つ点のみ異なる）

### 5.4 本文フォーマット（決定）

- **本文フォーマット: Markdown**（プラン推奨どおり。DB は `TEXT` のまま生の Markdown 文字列を保存 —
  レンダリングはフロント側の責務とし、バックエンドは検証・変換をしない）
- **画像埋め込み: 外部 URL 参照のみ**（Markdown 内に `![alt](https://...)` を直接書く運用。
  アップロード API は本 Step のスコープ外として見送り — プラン記載の「追加 Step として切り出す」の通り、
  必要になった時点で新しい dev-plan を起こす）

### 共通ロジック

- `internal/service/content.go`
  - `ValidateStatus` / `ValidateArticleCategory` — DB の `CHECK` 制約と同じ値域をハンドラー側で先に検証し、
    無効な値は `400` で弾く（DB 制約違反の生エラーが漏れるのを防ぐ）
  - `AsDuplicateSlug` — Postgres の unique_violation（`23505`）を検出して `ErrDuplicateSlug` に変換
    （スラッグ重複時に `500` ではなく `409` を返すため）
- `internal/repository/errors.go` — `ErrNotFound`（`pgx.ErrNoRows` を統一的に変換。
  DELETE の 0 件更新も `ErrNotFound` として扱う）
- `internal/handler/params.go` — `:id` パスパラメータの int64 パースを共通化

### ルーティング（`cmd/server/main.go`）

```
GET  /api/courses
GET  /api/courses/:slug
GET  /api/courses/:slug/lessons/:lessonSlug
GET  /api/articles
GET  /api/articles/:category/:slug
GET  /api/usecases
GET  /api/usecases/:slug

（以下すべて RequireAdmin 経由）
POST/PUT/DELETE /api/admin/courses[/:id]
POST/PUT/DELETE /api/admin/courses/:id/lessons, /api/admin/lessons/:id
POST/PUT/DELETE /api/admin/articles[/:id]
POST            /api/admin/articles/:id/tags
POST/PUT/DELETE /api/admin/usecases[/:id]
```

---

## プラン外で追加対応したこと

- **重大バグの発見・修正**: `internal/middleware/auth.go` の `RequireAdmin` が `RequireReader` を
  **ただの関数として直接呼び出して**いたため、`RequireReader` 内部の `c.Next()` が
  Gin のハンドラーチェーンを先送りし、`RequireAdmin` が Admin 判定を行う**前に**
  実際のルートハンドラー（例: `AdminCreateCourse`）が実行されてしまうバグがあった。
  `c.JSON(...)` で応答を書き込むハンドラーでは、後から `RequireAdmin` が `403` を返そうとしても
  レスポンスは既に送信済みのため、**Reader ユーザーが Admin 限定 API を実行できてしまっていた**
  （dev-plan-04-auth の検証時は応答本文を書かない空ハンドラーでテストしたため、この不具合を
  見逃していた）。
  - 修正: セッション検証ロジックを `resolveUser`（Gin のフロー制御を一切行わない純粋関数）に切り出し、
    `RequireReader` と `RequireAdmin` はそれぞれ独立して `resolveUser` を呼び、自分自身の
    `c.Next()`/`c.AbortWithStatus` のみを扱うように変更（ネストした `c.Next()` を完全に排除）
  - 実際に「Reader が admin API で作成したデータが DB に残る」ことを再現した上で修正し、
    修正後は当該データが作成されないこと・`403`/`401`/`201` それぞれが正しく返ることを確認済み
- 上記修正は `dev-plan-04-auth` が作った `internal/middleware/auth.go` への変更だが、
  影響範囲は `RequireAdmin` のみ（`RequireReader` の外部インターフェース・戻り値は変更なし、
  `/auth/me` の動作も再確認して問題ないことを確認済み）

## プラン未実施（意図的にスキップ）

- 画像アップロード API（5.4 で「外部 URL 参照のみ」と決定したため対象外）

---

## 検証

- `go build ./...` / `go vet ./...` / `gofmt -l .`（差分なし）— 成功
- `docker compose up -d --build line-api` → `/health` 引き続き `200 OK`
- **公開 API**（seed データに対して）:
  `GET /api/courses`, `/api/courses/line-marketing`（レッスン同梱）,
  `/api/courses/line-marketing/lessons/intro`, `/api/articles?category=line-operation`,
  `/api/articles/line-operation/rich-menu-basics`, `/api/usecases`, `/api/usecases/sample-salon`
  — すべて期待通りの JSON を返すことを確認
- **Admin CRUD 一連**（テスト用 admin セッションを DB に直接投入して確認）:
  - `POST /api/admin/courses`（`status:draft`）→ `201`、直後に公開一覧・詳細から見えない（`404`）ことを確認
  - `PUT` で `status:published` に更新 → 公開詳細が `200` になることを確認
  - 同一 slug で再作成 → `409`（`AsDuplicateSlug` の変換を確認）
  - `DELETE` → `204`、再度 `DELETE` → `404`（`ErrNotFound` 経由）
  - 記事作成 → タグ作成+付与（`POST /api/admin/articles/:id/tags`）→
    `?tag=` フィルタで一致/不一致それぞれ正しい結果を確認
  - 不正な `category` クエリ → `400`
- **認可境界**（重大バグ修正前後で再テスト）:
  - 未ログインで Admin API → `401`
  - Reader ユーザーで Admin API → **修正前は `201`（バグ）** → **修正後は `403`、かつ DB にデータが
    作成されないことを確認**
  - Admin ユーザーで Admin API → `201`（正常動作を維持していることを確認）
  - 修正後に `GET /auth/me` が admin セッションで引き続き `200` を返すことを確認（`RequireReader` の
    リファクタが auth 機能を壊していないことの再確認）
- 検証で作成したテストデータ（courses `x`/`y`/`admin-ok-check`、reader テストユーザー、
  テストセッション、テスト記事・タグ）はすべて DB から削除済み。seed データ（`line-marketing` 講座・
  `rich-menu-basics` 記事等）とユーザーの実ログインデータ（LINE/Google 各1件）は保持

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| Admin で講座・レッスン・記事・事例を作成・更新・削除できる | ✅ |
| 未ログイン/Reader で公開コンテンツの一覧・詳細が取得できる | ✅ |
| `draft` ステータスのコンテンツが公開 API に出てこない | ✅（一覧・詳細とも `WHERE status='published'` で確認） |
| Reader が Admin API を叩くと `403` が返る | ✅（バグ修正後に確認。修正前は認可バイパスが発生していた） |

## 次のステップ

- `dev-plan-06-exam-progress-api.md`（試験・進捗 API — 同じ `internal/middleware`（Reader/Admin）・
  `internal/service` のパターンを踏襲できる）
- フロントエンド側（Step 09/10）でこの API を消費する際、`RequireAdmin` のバグ修正により
  Admin 判定が仕様通りであることを前提にしてよい
