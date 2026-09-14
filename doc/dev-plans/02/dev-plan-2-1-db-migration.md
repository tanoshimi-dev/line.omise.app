# Step 2-1 — DB スキーマ設計・マイグレーション

**フェーズ:** Phase 2 — クイズ・検定機能
**依存:** Phase 1 完了（[dev-plan-02-database.md](../01/dev-plan-02-database.md) DB 基盤・`golang-migrate`）

---

## ゴール

クイズ（練習モード）・検定（試験モード）機能に必要なテーブルを設計し、開発環境に適用できる
マイグレーションを用意する。本番への適用手順・確認は [dev-plan-2-7-deploy-production.md](dev-plan-2-7-deploy-production.md) で扱う。

> **用語の整理（重要・混同注意）**
> 本 Step で新設する `quizzes` は、`mode` カラムで「クイズ（`practice`）」と「検定（`exam`）」を区別する。
> Phase 1 の `exams` / `exam_questions` / `exam_choices`（レッスン合否判定用、`lesson_id` 必須）とは別物であり、
> 既存テーブルには一切変更を加えない。

---

## タスク

### 2-1.1 スキーマ設計

- [ ] `quizzes`
  - `id`, `slug`（UNIQUE）, `title`, `description`, `mode`（`practice` / `exam`, CHECK制約）,
    `passing_score`（NULL可、`exam` のみ使用）, `published`
- [ ] `quiz_questions`
  - `id`, `quiz_id`（FK）, `question_text`, `allow_multiple`（複数回答可否）,
    `explanation`（解説本文）, `reference_url`（NULL可）, `sort_order`
- [ ] `quiz_choices`
  - `id`, `question_id`（FK）, `choice_text`, `is_correct`, `sort_order`
- [ ] `user_quiz_attempts`（検定モードの受験セッション。1回の受験 = 1レコード）
  - `id`, `user_id`（FK）, `quiz_id`（FK）, `score`, `total_questions`,
    `passed`（NULL可。`passing_score` 未設定時は NULL）, `started_at`, `submitted_at`
- [ ] `user_quiz_answers`（設問単位の解答履歴。練習モードは `attempt_id = NULL`、検定モードは受験セッションに紐付け）
  - `id`, `user_id`（FK）, `question_id`（FK）, `attempt_id`（FK, NULL可）,
    `selected_choice_ids`（`BIGINT[]`）, `is_correct`, `answered_at`

### 2-1.2 マイグレーションファイル作成

- [ ] `sys/02_backend/migrations/007_quiz.up.sql`
  - `quizzes` → `quiz_questions` → `quiz_choices` → `user_quiz_attempts` → `user_quiz_answers` の順に作成
    （既存の `001`〜`006` の連番に続ける。実装時点の最新連番を確認して番号を確定する）
- [ ] `sys/02_backend/migrations/007_quiz.down.sql`
  - 上記と逆順に `DROP TABLE`
- [ ] 開発用シード（任意）: `sys/02_backend/migrations/seed.sql` にサンプルクイズ・検定を追記
  - 例: 「LINE運用クイズ」（practice、複数回答問題を含む）、「基礎LINE検定」（exam、`passing_score` 設定あり）

### 2-1.3 ローカル動作確認

- [ ] `make migrate-up` 相当のコマンドでローカル DB に適用できる
- [ ] `.down.sql` で正常にロールバックできる（適用 → ロールバック → 再適用のサイクルを確認）
- [ ] 外部キー制約（`ON DELETE CASCADE`）がクイズ/設問/選択肢の削除時に正しく連鎖することを確認

---

## 成果物

- `sys/02_backend/migrations/007_quiz.up.sql` / `.down.sql`
- （任意）`seed.sql` へのサンプルデータ追記

## 完了条件

- ローカル DB に `quizzes` / `quiz_questions` / `quiz_choices` / `user_quiz_attempts` / `user_quiz_answers` が作成される
- マイグレーションの適用・ロールバックが両方エラーなく完了する
- Step 2-2（Admin API）・Step 2-3（解答・採点 API）がこのスキーマに対して実装できる状態になっている
