# Step 2-1 実装結果 — DB スキーマ設計・マイグレーション

**対応プラン:** [dev-plan-2-1-db-migration.md](../dev-plan-2-1-db-migration.md)

---

## 実施内容

### 2-1.1〜2-1.2 スキーマ・マイグレーションファイル

プラン記載のテーブルをそのまま実装（連番はディレクトリの最新 `006_usecase_thumbnail` に続けて `007`）:

- `sys/02_backend/migrations/007_quiz.up.sql`
  - `quizzes`（`slug` UNIQUE、`mode` は `CHECK (mode IN ('practice','exam'))`、`passing_score` は NULL 許容）
  - `quiz_questions`（`allow_multiple`、`explanation` NOT NULL、`reference_url` NULL 許容）
  - `quiz_choices`（`is_correct`）
  - `user_quiz_attempts`（検定モードの受験セッション。`passed` は NULL 許容）
  - `user_quiz_answers`（`attempt_id` NULL 許容 — 練習モードは NULL、検定モードは受験セッションに紐付け）
- `sys/02_backend/migrations/007_quiz.down.sql`
  - `user_quiz_answers` → `user_quiz_attempts` → `quiz_choices` → `quiz_questions` → `quizzes` の順で `DROP TABLE`（既存 `003_exam_progress.down.sql` と同じ「作成の逆順」の流儀に合わせた）

主キーは `BIGSERIAL`、外部キーは全て `ON DELETE CASCADE`。既存の `exams` / `exam_questions` / `exam_choices` / `user_exam_results` には一切変更なし。

### 2-1.3 開発用シード

プランでは任意タスクとして「LINE運用クイズ（practice）」「基礎LINE検定（exam）」のサンプル投入が挙がっていたが、**今回は見送った**。
理由: `quiz_questions` / `quiz_choices` には（`courses`/`lessons` の `(course_id, slug)` のような）自然な UNIQUE キーがなく、既存 `seed.sql` の `ON CONFLICT DO NOTHING` パターンでは冪等な再投入ができない。
Admin CRUD API（Step 2-2）が揃った時点で、API 経由 or それに合わせた `NOT EXISTS` ガード付きシードを追加する方が安全と判断した。

---

## プラン外で追加対応したこと

なし。

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- サンプルクイズ・検定のシード投入（任意タスク） — Step 2-2（Admin API）以降で追加を検討
- `internal/repository` を使った実際のクエリ層 — Step 2-2〜2-3 で実装（本 Step はスキーマのみ）

---

## 検証

- `DATABASE_URL=postgres://line_omise:line_omise@localhost:5432/line_omise_app?sslmode=disable go run ./cmd/migrate up` → `migrate: up complete`
  - `go run ./cmd/migrate version` → `version=7 dirty=false`
  - `docker compose exec postgres psql -c '\dt'` → `quizzes` / `quiz_questions` / `quiz_choices` / `user_quiz_attempts` / `user_quiz_answers` の5テーブルを確認
- CHECK 制約の検証: `INSERT INTO quizzes (..., mode) VALUES (..., 'bogus')` → `quizzes_mode_check` 違反で正しく拒否されることを確認
- カスケード削除の検証（トランザクション内で作成 → 検証 → `ROLLBACK` し、本番データには影響なし）:
  - quiz → question → choice(2件) / attempt(1件) / answer(1件) を作成後、`DELETE FROM quizzes` → 関連する question/choice/attempt/answer が全て 0 件になることを確認
- ロールバック往復の検証:
  - `go run ./cmd/migrate down` → 全マイグレーション（001〜007）がロールバックされ `version` コマンドが `no migration` で終了（既存の `m.Down()` の挙動どおり、001〜006 も含めて全ロールバックする既存仕様。007 単体の down.sql 自体はクイズ関連5テーブルのみを対象にしている）
  - 直後に `\dt` で `quiz_*` / `user_quiz_*` テーブルが消えていることを確認
  - `go run ./cmd/migrate up` で再適用 → `version=7 dirty=false` に復帰、`\dt` で19テーブル（既存14 + 新規5）全てが揃っていることを確認

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| ローカル DB に `quizzes` / `quiz_questions` / `quiz_choices` / `user_quiz_attempts` / `user_quiz_answers` が作成される | ✅ |
| マイグレーションの適用・ロールバックが両方エラーなく完了する | ✅ |
| Step 2-2（Admin API）・Step 2-3（解答・採点 API）がこのスキーマに対して実装できる状態になっている | ✅ スキーマのみ用意（クエリ層は Step 2-2〜2-3 で実装） |

## 次のステップ

- [dev-plan-2-2-admin-api.md](../dev-plan-2-2-admin-api.md)（クイズ／検定・設問・選択肢の Admin CRUD API）
