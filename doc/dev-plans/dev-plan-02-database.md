# Step 02 — DB 選定・スキーマ・マイグレーション

**フェーズ:** Phase 1 — Foundation
**依存:** Step 01（インフラ）

---

## ゴール

コンテンツ CMS（講座・レッスン・記事・事例）と、認証（LINE Login / Google OAuth、
Admin/Reader ロール）・試験（exam）・学習進捗（progress）に必要なデータストアを選定し、
初期スキーマとマイグレーション基盤を用意する。

> 会員管理・サロン予約・スイーツショップの3ミニアプリのデータは対象外（別ドメインで運用中）。

---

## タスク

### 2.1 DB エンジン選定（未決定事項 — ここで決定する）

- [ ] DB エンジンを決定する。候補:
  - **PostgreSQL**（推奨 — `hannari.dev/cloudflare/log` と揃えられ、運用ノウハウを共有できる）
  - MySQL / SQLite（小規模なら検討可）
- [ ] マイグレーションツールを決定（例: `golang-migrate`）

### 2.2 ユーザー・認証

- [ ] `users` テーブル
  - `id`, `provider`（`line` / `google`）, `provider_user_id`, `email`, `display_name`, `avatar_url`
  - `role`（`admin` / `reader`, デフォルト `reader`）
  - `created_at`, `updated_at`
  - `(provider, provider_user_id)` に UNIQUE 制約
- [ ] Admin 権限の付与方法を決定（Step 04 で詳細確定。候補: 起動時の許可リスト env var でログイン時に自動昇格 / DB を手動で直接更新）

### 2.3 コンテンツ（講座・レッスン・記事・事例）

- [ ] `courses` — 講座（例: LINEマーケティング講座）
  - `id`, `slug`, `title`, `description`, `sort_order`, `status`（`draft`/`published`）, `created_at`, `updated_at`
- [ ] `lessons` — 講座内の各レッスン
  - `id`, `course_id` (FK), `slug`, `title`, `body`（Markdown/リッチテキスト）, `sort_order`, `status`, `created_at`, `updated_at`
  - `(course_id, slug)` UNIQUE
- [ ] `articles` — 記事一覧形式のコンテンツ（`/learn/line-operation/`, `/learn/ai/`）
  - `id`, `category`（`line-operation` / `ai`）, `slug`, `title`, `body`, `status`, `published_at`, `created_at`, `updated_at`
  - `(category, slug)` UNIQUE
- [ ] `article_tags` / `tags` — タグ絞り込み用（`/learn/line-operation/?tag=rich-menu` 相当）
  - `tags`: `id`, `name`, `slug`
  - `article_tags`: `article_id` (FK), `tag_id` (FK)
- [ ] `usecases` — 導入店舗インタビュー
  - `id`, `slug`, `client_name`, `title`, `body`, `status`, `published_at`, `created_at`, `updated_at`

### 2.4 試験（exam）

- [ ] `exams` — レッスンに紐づく試験
  - `id`, `lesson_id` (FK), `title`, `passing_score`
- [ ] `exam_questions`
  - `id`, `exam_id` (FK), `question_text`, `sort_order`
- [ ] `exam_choices`
  - `id`, `question_id` (FK), `choice_text`, `is_correct`, `sort_order`

### 2.5 学習進捗（progress）

- [ ] `user_lesson_progress`
  - `id`, `user_id` (FK), `lesson_id` (FK), `completed_at`
  - `(user_id, lesson_id)` UNIQUE
- [ ] `user_exam_results`
  - `id`, `user_id` (FK), `exam_id` (FK), `score`, `passed`, `submitted_at`

### 2.6 マイグレーション基盤

- [ ] `sys/02_backend/migrations/` ディレクトリを作成
- [ ] `001_users.up.sql` / `.down.sql`
- [ ] `002_content.up.sql` / `.down.sql`（courses, lessons, articles, tags, article_tags, usecases）
- [ ] `003_exam_progress.up.sql` / `.down.sql`（exams, exam_questions, exam_choices, user_lesson_progress, user_exam_results）
- [ ] マイグレーション実行用の CLI or Makefile ターゲットを用意（例: `make migrate-up`）

### 2.7 開発用シード

- [ ] `sys/02_backend/migrations/seed.sql`（開発用ダミーユーザー・講座・記事、任意）

---

## 成果物

- DB エンジン・マイグレーションツールの決定記録（このファイルに追記 or `doc/architecture-decision-record/` に ADR として記録）
- `sys/02_backend/migrations/001_users.up.sql` 〜 `003_exam_progress.up.sql`（および `.down.sql`）
- マイグレーション実行手順

## 完了条件

- ローカルで DB コンテナが起動し、マイグレーションが適用できる
- `users` / `courses` / `lessons` / `articles` / `usecases` / `exams` 系テーブルが作成される
- Step 05〜06 のコンテンツ・試験・進捗 API がこのスキーマに対して実装できる状態になっている
