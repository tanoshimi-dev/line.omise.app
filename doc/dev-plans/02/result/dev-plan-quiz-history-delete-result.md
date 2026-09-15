# Step — 自分のクイズ解答履歴・受験履歴を削除できるようにする — 実施結果

計画: [dev-plan-quiz-history-delete.md](../dev-plan-quiz-history-delete.md)

---

## 実施内容

計画のタスク1〜4を計画どおりに実施。

### 1. バックエンド — リポジトリ層

`sys/02_backend/internal/repository/quiz.go` に4メソッドを追加:

- `DeleteOwnAttempt(ctx, attemptID, userID)` — `user_id` をクエリ条件に含めることで所有者以外は
  `ErrNotFound`（`GetOwnAttempt` と同じパターン）。紐づく `user_quiz_answers` は既存の
  `ON DELETE CASCADE` で自動削除。
- `DeleteAttemptsForQuiz(ctx, userID, quizID)` — そのユーザー・そのクイズの受験履歴を一括削除。
- `DeleteOwnPracticeAnswer(ctx, answerID, userID)` — `attempt_id IS NULL` かつ本人所有のみ削除、
  それ以外は `ErrNotFound`。
- `DeletePracticeHistoryForQuiz(ctx, userID, quizID)` — そのユーザー・そのクイズの単発解答履歴を
  一括削除。

### 2. バックエンド — ハンドラ・ルーティング

`sys/02_backend/internal/handler/quiz.go` に4ハンドラを追加し、`GetOwnAttempt`/`GetAttempt` と
同じ「本人以外は404」パターンを踏襲:

- `DELETE /api/me/quizzes/:slug/attempts/:attemptId` → `DeleteAttempt`
- `DELETE /api/me/quizzes/:slug/attempts` → `DeleteAllAttempts`
- `DELETE /api/me/quizzes/:slug/history/:answerId` → `DeletePracticeAnswer`
- `DELETE /api/me/quizzes/:slug/history` → `DeleteAllPracticeHistory`

`GetPracticeHistory` のレスポンス各項目に `"id"`（`user_quiz_answers.id`）を追加— 従来は行を
一意に指す ID が無く、削除対象を指定できなかったため。

`sys/02_backend/internal/server/server.go` に上記4ルートを `requireReader` 配下に登録。

### 3〜4. フロントエンド

- `sys/03_frontend/web/src/lib/types.ts`: `QuizPracticeHistoryEntry` に `id: string` を追加。
- `QuizHistoryPanel.tsx` / `ExamAttemptsPanel.tsx`: 既存の汎用 `components/admin/DeleteButton.tsx`
  を再利用し、各行に削除ボタン、一覧上部に「すべて削除」ボタンを追加。どちらも削除後
  `onChanged?: () => void` を呼ぶ。
- `QuizProgressSection.tsx`: `QuizProgressCard` にローカル `quiz` state（初期値は props）を持たせ、
  `onChanged` から `GET /api/me/quizzes/:slug/progress` を再取得してサマリーを更新する
  `refreshProgress` を両パネルに渡す。また、件数が0になっても「閉じる」操作ができるよう、
  折りたたみボタンの表示条件を「件数>0 **または** 展開中」に変更（一括削除後にボタンごと消えて
  閉じられなくなる問題を回避）。

## 検証

- `go build ./...` / `go vet ./...` / `go test ./...`（`sys/02_backend`）— 全てパス。
  新規追加した6件（`TestDeletePracticeAnswer_OwnerCanDeleteOneEntry`,
  `TestDeletePracticeAnswer_OtherUsersEntryIs404`, `TestDeleteAllPracticeHistory_ClearsOwnHistoryOnly`,
  `TestDeleteAttempt_OwnerCanDeleteOneAttempt`, `TestDeleteAttempt_OtherUsersAttemptIs404`,
  `TestDeleteAllAttempts_ClearsOwnAttemptsOnly`）を含め、他ユーザーの履歴を削除できないこと・
  検定受験削除時に解答詳細がカスケード削除されること（`GetAttempt` が削除後404になることで確認）を
  検証済み。
- `npm run build` / `npm run test`（`sys/03_frontend/web`）— 型チェック含め成功、7ファイル32件
  全てパス。新規追加した2件（1件削除後の進捗再取得、一括削除後の進捗再取得）を含む。
  既存の `QuizProgressSection.test.tsx` は `DeleteButton` が `next/navigation` の `useRouter` を
  使うため、`vi.mock('next/navigation', ...)` の追加が必要だった（従来 Vitest 配下では
  `DeleteButton` を経由するコンポーネントが無く、このモックが存在しなかった）。

## 計画からの逸脱

- 完了条件にあった「Go テスト・Vitest が green（新規テスト追加含む）」は達成。それ以外は計画どおり
  実施し、設計上の逸脱はなし。
