# Step — 自分のクイズ解答履歴・受験履歴を削除できるようにする

**対象:** マイページのクイズ履歴機能の拡張
**依存:** dev-plan-quiz-mode-selection（実装済み）

---

## 背景・方針（ユーザー確認済み）

現状、マイページ（`QuizProgressSection` → `QuizHistoryPanel`/`ExamAttemptsPanel`）は
単発モードの解答履歴・検定モードの受験履歴を**閲覧のみ**でき、削除する手段が無い。

- **対象範囲:** 単発モードの解答履歴（`user_quiz_answers`, `attempt_id IS NULL`）・
  検定モードの受験履歴（`user_quiz_attempts`, カスケードで紐づく `user_quiz_answers` も削除）の
  **両方**に削除機能を追加する。
- **粒度:** 1件ずつの削除（履歴の各行に削除ボタン）と、そのクイズの履歴を一括削除する操作の
  **両方**を用意する。
- 削除はログイン中の本人の履歴のみ（他ユーザーの履歴は削除できない — 既存の `GetOwnAttempt` と
  同じ「本人以外は 404」パターンを踏襲）。管理者による削除（Admin 側の別機能）は対象外。

---

## タスク

### 1. バックエンド — リポジトリ層に削除メソッドを追加

`sys/02_backend/internal/repository/quiz.go`

- [ ] `DeleteOwnAttempt(ctx, attemptID, userID int64) error` — `DELETE FROM user_quiz_attempts
      WHERE id = $1 AND user_id = $2`。0件なら `ErrNotFound`（`GetOwnAttempt` と同じ「本人以外は
      404」パターン）。紐づく `user_quiz_answers` は既存の `ON DELETE CASCADE`（
      `sys/02_backend/migrations/007_quiz.up.sql:45`）で自動削除される。
- [ ] `DeleteAttemptsForQuiz(ctx, userID, quizID int64) error` — そのユーザー・そのクイズの
      `user_quiz_attempts` を一括 `DELETE`（0件でもエラーにしない — 「既に空の履歴を消す」は成功扱い）。
- [ ] `DeleteOwnPracticeAnswer(ctx, answerID, userID int64) error` — `DELETE FROM
      user_quiz_answers WHERE id = $1 AND user_id = $2 AND attempt_id IS NULL`。0件なら
      `ErrNotFound`。
- [ ] `DeletePracticeHistoryForQuiz(ctx, userID, quizID int64) error` — そのユーザー・そのクイズの
      単発解答（`attempt_id IS NULL`、`question_id IN (SELECT id FROM quiz_questions WHERE
      quiz_id = $2)`）を一括 `DELETE`。

### 2. バックエンド — ハンドラ・ルーティング追加

`sys/02_backend/internal/handler/quiz.go`

- [ ] `DeleteAttempt`: `DELETE /api/me/quizzes/:slug/attempts/:attemptId` — `GetOwnAttempt` で
      本人所有かつ対象クイズに属することを確認してから `DeleteOwnAttempt` を呼ぶ（`GetAttempt` の
      認可パターンを踏襲）。成功時 204。
- [ ] `DeleteAllAttempts`: `DELETE /api/me/quizzes/:slug/attempts` — `DeleteAttemptsForQuiz`。
      成功時 204。
- [ ] `DeletePracticeAnswer`: `DELETE /api/me/quizzes/:slug/history/:answerId` —
      `DeleteOwnPracticeAnswer`。成功時 204、対象なし/他人の履歴なら 404。
- [ ] `DeleteAllPracticeHistory`: `DELETE /api/me/quizzes/:slug/history` —
      `DeletePracticeHistoryForQuiz`。成功時 204。
- [ ] `GetPracticeHistory` のレスポンス各項目に `"id"`（`user_quiz_answers.id`）を追加する
      （現状 `question_id`/`question_text`/`selected_choice_ids`/`is_correct`/`answered_at` のみで、
      行を一意に指す ID が無く削除対象を指定できないため）。

`sys/02_backend/internal/server/server.go`

- [ ] 上記4エンドポイントを `requireReader` ミドルウェアで `/me/quizzes/...` 配下に登録
      （既存の `GET /me/quizzes/:slug/history` 等と同じグループ）。

### 3. フロントエンド — 型・API 呼び出し

- [ ] `sys/03_frontend/web/src/lib/types.ts`: `QuizPracticeHistoryEntry` に `id: string` を追加。

### 4. フロントエンド — 履歴パネルに削除 UI を追加

`sys/03_frontend/web/src/components/mypage/QuizHistoryPanel.tsx`

- [ ] 各行に削除ボタンを追加し、`DELETE /api/me/quizzes/:slug/history/:answerId` を呼ぶ
      （既存の汎用 `components/admin/DeleteButton.tsx` を再利用 — path・確認メッセージ・
      `onDeleted` コールバックだけで完結する汎用コンポーネントのため、admin 専用ロジックは無い）
- [ ] 履歴が1件以上あるとき、「すべての解答履歴を削除」ボタンを一覧上部に表示し、
      `DELETE /api/me/quizzes/:slug/history` を呼ぶ
- [ ] 削除後、`QuizProgressSection`（親）の進捗サマリー（`answered_count`/`correct_count`）が
      古いまま残らないよう、新規 `onChanged?: () => void` prop を呼ぶ

`sys/03_frontend/web/src/components/mypage/ExamAttemptsPanel.tsx`

- [ ] 各行に削除ボタンを追加し、`DELETE /api/me/quizzes/:slug/attempts/:attemptId` を呼ぶ
- [ ] 受験履歴が1件以上あるとき、「すべての受験履歴を削除」ボタンを一覧上部に表示し、
      `DELETE /api/me/quizzes/:slug/attempts` を呼ぶ
- [ ] 削除後、同様に `onChanged?: () => void` を呼ぶ

`sys/03_frontend/web/src/components/mypage/QuizProgressSection.tsx`

- [ ] `QuizProgressCard` に、表示中の `quiz`（`QuizProgressSummary`）をローカルで上書きできる state
      を追加し、`onChanged` が呼ばれたら `GET /api/me/quizzes/:slug/progress` を再取得して
      サマリー（解答済み数・正答数・受験回数・最高/直近スコア・合否）を更新する

---

## 成果物

- `sys/02_backend/internal/repository/quiz.go`（削除メソッド4種）
- `sys/02_backend/internal/handler/quiz.go`（削除ハンドラ4種、`GetPracticeHistory` に `id` 追加）
- `sys/02_backend/internal/server/server.go`（ルート追加）
- `sys/03_frontend/web/src/lib/types.ts`（`QuizPracticeHistoryEntry.id` 追加）
- `sys/03_frontend/web/src/components/mypage/QuizHistoryPanel.tsx` /
  `ExamAttemptsPanel.tsx` / `QuizProgressSection.tsx`

## 完了条件

- マイページで、単発モードの解答履歴・検定モードの受験履歴のどちらも、1件ずつ削除できる
- マイページで、単発モード・検定モードそれぞれの履歴を一括削除できる
- 他ユーザーの履歴は削除できない（存在しないIDと同じ404になる）ことをテストで確認
- 削除後、進捗サマリー（解答済み数・受験回数など）が最新の状態に更新される
- 検定モードの受験を削除すると、紐づく解答詳細（`user_quiz_answers`）もカスケードで削除される
- 既存の Go テスト・Vitest が green（新規テスト追加含む）
