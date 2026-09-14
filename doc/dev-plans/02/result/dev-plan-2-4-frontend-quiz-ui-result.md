# Step 2-4 実装結果 — フロントエンド: クイズ（練習）・検定 受験 UI

**対応プラン:** [dev-plan-2-4-frontend-quiz-ui.md](../dev-plan-2-4-frontend-quiz-ui.md)

---

## 実施内容

### 2-4.1 配置の決定

- クイズ／検定は **`/learn/` 配下**（`/learn/quiz/`）に配置。既存の3カテゴリ（マーケティング講座・運用/設定・AI活用事例）と並ぶ
  4つ目のカテゴリとして `src/app/learn/page.tsx` にカードを追加し、`src/components/Header.tsx` の学習コンテンツサブメニューにも
  「クイズ・検定」を追加（既存3項目の並び順は変更せず末尾に追加）
- `learn/page.tsx` のカードグリッドは4項目になったため `sm:grid-cols-3` → `sm:grid-cols-2 lg:grid-cols-4` に調整
- クイズ一覧ページ（`/learn/quiz`）は「クイズ（練習モード）」「検定（検定モード）」の2セクションに分けて表示

### 2-4.2 練習モード UI

- `src/components/quiz/QuizQuestionCard.tsx`
  - 設問1問ずつ表示 → 選択肢（`allow_multiple` に応じてラジオ／チェックボックス出し分け）→「解答する」
  - 解答後: 正解の選択肢を緑、誤って選んだ選択肢を赤でハイライト、解説・参考リンクを表示、「次の問題へ」
  - 最終問題の後は「◯問中◯問正解しました」の完了画面を表示

### 2-4.3 検定モード UI

- `src/components/quiz/ExamRunner.tsx`
  - 全設問を一度に表示（正誤・解説は見せない）、「解答済み: X / Y問」の進捗表示
  - 提出時、未回答の設問があれば `window.confirm` で警告してから送信（プランの2-3.3仕様に合わせた確認ダイアログ）
  - 結果画面: 最終スコア（正答数／全問数）、`passing_score` 設定時は合否表示、設問ごとの正誤・正しい選択肢・解説の一覧

### 配置ページ

- `src/app/learn/quiz/page.tsx` — 一覧ページ（サーバーコンポーネント、`GET /api/quizzes` を使用）
- `src/app/learn/quiz/[slug]/page.tsx` — 受験ページ（サーバーコンポーネントで `GET /api/quizzes/:slug` を取得し、
  `quiz.mode` に応じて `QuizQuestionCard` / `ExamRunner` を出し分け）

### 2-4.4 未ログイン時の扱い

- 練習モード・検定モードとも未ログインで解答・受験可能（`useAuth()` の `user` が null でも動作）
- 結果画面（練習モードの完了画面／検定モードのスコア画面）で `user` が null の場合のみ、
  既存の `LoginPrompt` コンポーネント（LINE/Googleログイン導線）を表示

### 型定義

- `src/lib/types.ts` に Reader/公開向けの型を追加: `QuizListItem` / `QuizChoice` / `QuizQuestion` / `Quiz` /
  `QuizAnswerResult` / `QuizSubmitQuestionResult` / `QuizSubmitResult`（Admin向けの `AdminQuiz*` とは別型、
  Phase 1 の `Exam`/`AdminExam` の分離パターンを踏襲）

---

## プラン外で追加対応したこと

- **`GET /api/quizzes`（公開・一覧）を Step 2-3 に追記する形でバックエンドに追加**
  （`internal/handler/quiz.go` の `ListQuizzes`、`internal/server/server.go` にルート登録）。
  dev-plan-2-3 のエンドポイント一覧にはなかったが、本 Step 2-4.1 の「クイズ一覧ページ」に必須のため、
  `ListCourses`/`ListUsecases` と同じ公開一覧パターンで追加した。Go 側のテストも追加済み
  （`TestListQuizzes_ExcludesDraft`）。

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- マイページの受験履歴・進捗表示 — Step 2-5 で実装

---

## 検証

- `go build ./...` / `go vet ./...` / `gofmt -l .` — 成功（差分なし）
- `go test ./...` — 全パッケージ成功（`TestListQuizzes_ExcludesDraft` を含む）
- `npm run build`（Next.js）— 成功。`/learn/quiz`, `/learn/quiz/[slug]` を含む全ルートの型検証・ビルドが通過
- `npm test`（Vitest）— 22件成功（既存15件 + 新規: `QuizQuestionCard.test.tsx` 3件、`ExamRunner.test.tsx` 4件）
  - 練習モード: 解答→正誤・解説表示→次の問題へ→完了画面、未ログイン時の完了画面にログイン導線表示、
    未選択時は解答ボタンが無効
  - 検定モード: 全問提出→スコア・合否・設問ごとの振り返り、未回答時の確認ダイアログ表示／キャンセルで送信されないこと、
    未ログイン時の結果画面にログイン導線表示
- **実ブラウザでの動作確認**（ローカルで `go run ./cmd/server`（ポート8090）と `next dev`（ポート3001）を一時起動し、
  DB に検証用サンプルデータを直接投入して確認。確認後、サンプルデータ・一時プロセスとも削除・停止済み）:
  - `/learn/quiz` 一覧ページでクイズ／検定がセクション分けして表示されることを確認
  - 練習モード（複数回答設問）: 2つの正解選択肢をチェック→「正解です！」+ 解説 + 参考リンク表示 →
    「次の問題へ」→「1問中 1問正解しました」+ 未ログイン時のログイン導線を確認
  - 検定モード（設問2問）: 両方正解を選択して提出 →「2 / 2問正解」「合格（合格点: 60%）」+
    設問ごとの正誤・解説の振り返り + 未ログイン時のログイン導線を確認

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 練習モードのクイズに未ログインで解答し、その場で正誤・解説が確認できる | ✅ 実ブラウザで確認 |
| 検定モードの試験に全問解答して提出すると、最終スコア（・合否）と設問ごとの振り返りが表示される | ✅ 実ブラウザで確認 |
| 未ログイン時にログイン導線が適切に表示される | ✅ 実ブラウザで確認 |

## 次のステップ

- [dev-plan-2-5-frontend-mypage.md](../dev-plan-2-5-frontend-mypage.md)（マイページ「受験履歴・進捗」セクション）
