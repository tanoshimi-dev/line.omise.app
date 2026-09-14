# Step 06 実装結果 — 試験（exam）API・学習進捗（progress）API

**対応プラン:** [dev-plan-06-exam-progress-api.md](../dev-plan-06-exam-progress-api.md)

---

## 実施内容

### 6.1 試験の作成・管理（Admin）

- `internal/repository/exam.go` — `ExamRepository`
  - `Create`（lesson に対する試験作成）、`GetByLessonID` / `GetByID`
  - `CreateQuestion` / `UpdateQuestion` — 設問+選択肢をトランザクション（`pgx.Tx`）内で原子的に書き込み。
    `UpdateQuestion` は Step 05 の PUT 全置換方針を踏襲し、既存の選択肢を全削除してから作り直す
  - `DeleteQuestion`（`exam_choices` は `ON DELETE CASCADE` で自動削除）
  - `ListChoicesForExam` — 1回のJOINクエリで試験全体の選択肢（`is_correct`含む）を取得
    （Reader向けレスポンス生成時と採点時の両方で使用、後者のみ `is_correct` を利用）
- `internal/handler/exam.go` — `POST /api/admin/lessons/:lessonId/exam`,
  `POST /api/admin/exams/:examId/questions`, `PUT /api/admin/questions/:id`,
  `DELETE /api/admin/questions/:id`

**プラン外の判断（スキーマ追加）**: `exams.lesson_id` に元々 UNIQUE 制約がなく、
プランのルート設計（`/lessons/:lessonId/exam` が単数形）が前提とする「1レッスンにつき試験は1つ」を
DB レベルで保証できていなかった。`migrations/005_exam_lesson_unique.up.sql`／`.down.sql` を追加し
`UNIQUE (lesson_id)` 制約を付与。重複作成は `service.AsExamAlreadyExists` で検出し `409` を返す。

### 6.2 試験の受験（Reader）

- `GET /api/lessons/:lessonId/exam`（`RequireReader`）— `internal/handler/exam.go` の
  `examPublicJSON` は選択肢の `is_correct` を一切含まない別レンダリング関数
  （Admin向け `questionAdminJSON` とは完全に分離し、うっかり `is_correct` が漏れる経路を作らない設計）
- `POST /api/lessons/:lessonId/exam/submit` — `internal/service/exam.go` の `GradeExam` で採点
  - **採点方式（決定）**: `score` は **0〜100 のパーセンテージ**
    （正解数 ÷ 設問数 × 100、四捨五入）。`passing_score` はこの尺度に対する閾値として解釈
    （プランのスキーマに明記がなかったため、設問数の増減に強い％方式を採用）
  - 未回答の設問は不正解扱い。同じ設問への重複回答は最初の回答のみ採用
  - 設問に属さない `choice_id` の指定は `400`（`service.ErrInvalidAnswer`）
  - 結果は `user_exam_results` に保存し `{exam_id, score, passed, submitted_at}` を返す

### 6.3 学習進捗（Reader）

- `internal/repository/progress.go` — `ProgressRepository`
  - `MarkLessonComplete`（`ON CONFLICT (user_id, lesson_id) DO UPDATE` — 再完了は `completed_at` 更新）
  - `CompletedLessons`（講座内の複数レッスンの完了状況を1クエリで取得）
  - `CountCompletedByCourse`（`/api/me/progress` 用の件数集計）
  - `SaveExamResult` / `LatestExamResult`
- `internal/handler/progress.go`
  - `POST /api/lessons/:lessonId/complete`
  - `GET /api/courses/:slug/progress` — 完了数/全体数、レッスンごとの完了状態、
    レッスンに試験がある場合は最新の受験結果を同梱
  - `GET /api/me/progress`（任意項目として実装 — 公開講座ごとの完了数サマリー）

**6.4 の決定事項:**

- **再受験の扱い: 履歴を残す**（上書きしない）。`user_exam_results` に元々 `(user_id, exam_id)` の
  UNIQUE 制約がなく、素直に毎回 INSERT する設計がスキーマと自然に合致するため。
  進捗表示（`/api/courses/:slug/progress`, `/api/me/progress` 経由の `latest_result`）は
  「最新の受験結果」（`ORDER BY submitted_at DESC LIMIT 1`）を採用 — 「ベストスコア」ではない点に注意
- 存在しない `lesson_id`（未公開含む）/ `question_id` へのアクセスは `404`
  （Reader 向けエンドポイントは lesson・course の両方が `published` であることも確認し、
  下書きコンテンツの試験を lesson_id 推測で覗けないようにした）
- 他人の進捗へのアクセス防止: 全ての progress/exam エンドポイントで `user_id` は
  `middleware.CurrentUser(c)`（セッション由来）からのみ取得し、URL/ボディparamでは一切受け取らない

---

## プラン外で追加対応したこと

- `migrations/005_exam_lesson_unique.up.sql`／`.down.sql`（上記 6.1 参照）
- `internal/service/errors.go` — Step 05 の `AsDuplicateSlug` にあった Postgres unique_violation 判定を
  `isUniqueViolation(err, constraintName)` として切り出し、Step 06 の `AsExamAlreadyExists` と共有
  （Step 05 の動作・戻り値は変更なし、内部実装のみ整理）

## プラン未実施

なし。6.1〜6.4 すべて実装・検証済み。

---

## 検証

- `go build ./...` / `go vet ./...` / `gofmt -l .`（差分なし）— 成功
- マイグレーション: `go run ./cmd/migrate up` → `version=5`
- `docker compose up -d --build line-api` → `/health` 引き続き `200 OK`
- Admin で試験作成 → 同一レッスンへの再作成が `409` になることを確認
- 設問+選択肢の作成（トランザクション経由）→ 正しく採番・保存されることを確認
  （日本語本文は当初ターミナル表示が文字化けしたが、これは検証端末（Windows, コードページ 932）の
  シェル入力エンコーディングの問題と判明 — UTF-8 ファイル経由で送信し直したところ完全に正しく
  往復することを確認。アプリ・DB・JSON 経路には問題なし）
- Reader での `GET /api/lessons/:lessonId/exam` → レスポンスに `is_correct` が一切含まれないことを確認
- 未ログインで exam 取得/受験/完了マーク → すべて `401`
- 採点ロジック: 全問正解 → `score=100, passed=true`、1問不正解（設問2問中）→ `score=50`、
  合格点70に対し `passed=false` となることを確認
- 存在しない設問に属す `choice_id` を指定した受験 → `400`
- レッスン完了マーク → `POST .../complete` → `completed_at` が返る
- 講座進捗取得 → 完了済みレッスン・最新試験結果（直近の受験、ベストではなく最新であることを確認）が
  正しく反映されることを確認
- **進捗の他ユーザー分離を確認**: 別の Reader ユーザーで同じ講座の進捗を取得すると
  `completed:false`, `latest_result:null` となり、最初のユーザーの進捗が漏れていないことを確認
- `/api/me/progress` — 公開講座ごとの完了数サマリーが正しく返ることを確認
- Reader で Admin 専用試験作成 API を叩くと `403`（dev-plan-05 で修正した `RequireAdmin` が
  本 Step の新規ルートでも正しく機能することを確認）
- 検証で作成したテストデータ（試験・設問・選択肢・Reader テストユーザー2名・関連セッション・
  進捗/受験結果レコード）はすべて DB から削除済み。seed データと実ログインユーザーは保持

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| Admin が試験・設問を作成できる | ✅ |
| Reader が試験を受験し、採点結果が返る | ✅ |
| レッスン完了マークと進捗取得が動作する | ✅ |
| 正解フラグが受験前のレスポンスに含まれない（カンニング防止） | ✅（Admin向け/Reader向けレンダリング関数を分離して確認） |
| 他ユーザーの進捗にアクセスできない | ✅（別ユーザーでの取得結果を確認） |

**Phase 1 — CMS API（Step 05〜06）完了。**

## 次のステップ

- `dev-plan-07-frontend-base.md` は完了済みのため、`dev-plan-08-frontend-lp.md`〜
  `dev-plan-11-frontend-admin.md`（フロントエンド実装）が本 Step の API を消費する形で着手可能
