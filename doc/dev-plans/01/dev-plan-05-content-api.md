# Step 05 — コンテンツ CMS API

**フェーズ:** Phase 1 — CMS API
**依存:** Step 04（認証）

---

## ゴール

講座（courses）・レッスン（lessons）・記事（articles: line-operation / ai）・
導入事例（usecases）の CRUD API を実装する。書き込みは Admin 限定、公開コンテンツの
閲覧は誰でも可能（Reader ログイン不要）とする。

---

## タスク

### 5.1 講座・レッスン API（`/learn/line-marketing/`）

- [ ] `GET /api/courses` — 公開講座一覧
- [ ] `GET /api/courses/:slug` — 講座詳細（レッスン一覧を含む）
- [ ] `GET /api/courses/:slug/lessons/:lessonSlug` — レッスン詳細
- [ ] `POST /api/admin/courses` — 講座作成（Admin）
- [ ] `PUT /api/admin/courses/:id` — 講座更新（Admin）
- [ ] `DELETE /api/admin/courses/:id` — 講座削除（Admin）
- [ ] `POST /api/admin/courses/:id/lessons` — レッスン作成（Admin）
- [ ] `PUT /api/admin/lessons/:id` — レッスン更新（Admin）
- [ ] `DELETE /api/admin/lessons/:id` — レッスン削除（Admin）
- [ ] `status = draft` のコンテンツは公開 API から除外する

### 5.2 記事 API（`/learn/line-operation/`, `/learn/ai/`）

- [ ] `GET /api/articles?category=line-operation` — カテゴリ別記事一覧
- [ ] `GET /api/articles?category=line-operation&tag=rich-menu` — タグ絞り込み
- [ ] `GET /api/articles/:category/:slug` — 記事詳細
- [ ] `POST /api/admin/articles` — 記事作成（Admin）
- [ ] `PUT /api/admin/articles/:id` — 記事更新（Admin）
- [ ] `DELETE /api/admin/articles/:id` — 記事削除（Admin）
- [ ] タグの作成・付与 API（`POST /api/admin/articles/:id/tags`）

### 5.3 導入事例 API（`/usecase/`）

- [ ] `GET /api/usecases` — 事例一覧
- [ ] `GET /api/usecases/:slug` — 事例詳細
- [ ] `POST /api/admin/usecases` — 事例作成（Admin）
- [ ] `PUT /api/admin/usecases/:id` — 事例更新（Admin）
- [ ] `DELETE /api/admin/usecases/:id` — 事例削除（Admin）

### 5.4 本文フォーマット

- [ ] `body` のフォーマットを決定（Markdown 推奨 — フロント側で `react-markdown` 等を使いレンダリング）
- [ ] 画像埋め込み方法を決定（外部 URL 参照 or アップロード API — アップロードが必要なら追加 Step として切り出す）

---

## 成果物

- `internal/handler/course.go`, `article.go`, `usecase.go`
- `internal/repository/course.go`, `article.go`, `usecase.go`
- `internal/service/content.go`（公開判定・スラッグ重複チェック等の共通ロジック）

## 完了条件

- Admin で講座・レッスン・記事・事例を作成・更新・削除できる
- 未ログイン/Reader で公開コンテンツの一覧・詳細が取得できる
- `draft` ステータスのコンテンツが公開 API に出てこない
- Reader が Admin API を叩くと `403` が返る
