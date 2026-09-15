# Step — クイズの解答モードをユーザー選択制にする（単発モード / 検定モード）— 実施結果

計画: [dev-plan-quiz-mode-selection.md](dev-plan-quiz-mode-selection.md)

---

## 実施内容

計画のタスク1〜8を計画どおりに実施。主な変更点は以下の通り。

### 1. DBマイグレーション

- `sys/02_backend/migrations/008_quiz_drop_mode.up.sql` / `.down.sql` を追加し、`quizzes.mode` 列
  （`NOT NULL DEFAULT 'practice' CHECK (mode IN ('practice','exam'))`）を削除。

### 2〜4. バックエンド

- `repository.Quiz` から `Mode` フィールドを削除し、`Create`/`Update`/各 SELECT から `mode` 列を除去。
- `service.ValidateQuizMode` / `ValidatePassingScore` / `ErrInvalidQuizMode` / `ErrPassingScoreNotAllowed`
  を削除（呼び出し元が無くなったため）。`passing_score` はどのクイズにも自由に設定できる独立項目に。
- `service.ErrQuizModeMismatch`（未使用センチネル）を削除。
- `AnswerQuestion`（単発解答）・`SubmitQuiz`（一括提出）双方のモードゲート（400エラー）を削除 —
  どのクイズもどちらの方式でも解答できるようになった。
- `GetPracticeHistory` / `ListAttempts` のモード分岐を削除し、データがあれば（無ければ空配列を）常に返す形に。
- `quizProgress`（`GetQuizProgress` / `GetMyQuizzesProgress` の内部関数）を統合し、単発系フィールド
  （`answered_count`/`correct_count`）と検定系フィールド（`attempt_count`/`best_score`/`latest_score`/
  `latest_passed`/`passing_score`）を常に両方含むレスポンスに変更。

### 5〜7. フロントエンド

- Admin 画面（`admin/quizzes/new`, `[id]/edit`, 一覧）から種別（mode）の select を削除し、
  「合格点を設定する」チェックボックス + 任意の数値入力に置き換え。一覧の「種別」列は「合格点」列に変更。
- 公開一覧 `learn/quiz/page.tsx` を「クイズ」「検定」の2セクションから単一リストに統合。
- `learn/quiz/[slug]/page.tsx` から `quiz.mode` によるコンポーネント振り分けを削除し、新規
  `components/quiz/QuizModeSelect.tsx`（クライアントコンポーネント）を追加。解答開始前に必ず
  「単発モード」「検定モード」を選ばせ、選択後に同じ `quiz` データを `QuizQuestionCard`／`ExamRunner`
  に渡す。admin 側のデフォルト値は存在しないため初期状態はどちらも未選択。
- `QuizQuestionCard`／`ExamRunner` 自体は元々 `quiz.mode` を参照していなかったため、ロジック変更なし
  （doc コメントのみ更新）。

### 8. マイページ

- `QuizProgressSection.tsx` の `PracticeProgressCard`/`ExamProgressCard`（`quiz.mode` で二択）を
  単一の `QuizProgressCard` に統合。単発モードの進捗＋履歴パネルと検定モードの進捗＋受験履歴パネルを、
  それぞれデータがある場合に独立して展開できる形で、同じカード内に両方表示する。
- `QuizHistoryPanel`／`ExamAttemptsPanel`／`AttemptDetailModal` は元々 slug 単位で独立して動作しており
  変更不要。

### 型定義

- `sys/03_frontend/web/src/lib/types.ts`: `QuizMode` 型、`Quiz`/`QuizListItem`/`AdminQuiz` の `mode`
  フィールドを削除。`QuizPracticeProgress`/`QuizExamProgress`/`QuizProgressSummary`（union）を、
  単発系・検定系フィールドを両方持つ単一の `QuizProgressSummary` インターフェースに統合。

### テスト更新

- Go: `admin_quiz_test.go`／`quiz_test.go`／`repository/quiz_test.go` の `mode` パラメータ・アサーションを
  更新。モードゲートを検証していた `TestAnswerQuestion_ExamModeQuestionRejected` /
  `TestSubmitQuiz_PracticeModeQuizRejected` は、逆に「passing_score の有無に関わらずどちらの方式でも
  解答・提出できる」ことを確認する `TestAnswerQuestion_WorksEvenWithPassingScoreSet` /
  `TestSubmitQuiz_WorksEvenWithoutPassingScore` に置き換え。`TestAdminCreateQuiz_InvalidModeRejected`
  （mode バリデーション自体のテスト）は概念が無くなったため削除。
- Vitest: `QuizProgressSection.test.tsx` を統合後のレスポンス形状に合わせて更新し、単発・検定の両方の
  セクションが同じクイズに同時に表示されることを確認する新規テストを追加。`ExamRunner.test.tsx` /
  `QuizQuestionCard.test.tsx` からモックデータの `mode` フィールドを削除。
- Playwright（`sys/04_e2e/tests/quiz.spec.ts`, `helpers/db.ts`）: `seedQuiz(mode, ...)` を
  `seedQuiz(options?)` に変更（DB へ `mode` 列を書き込まなくなったため）。各シナリオに
  `単発モード`／`検定モード` ボタンのクリックを追加し、テスト名・コメントの「practice-mode」表現も
  「single-mode」に更新。

## 検証

- `go build ./...` / `go vet ./...` / `go test ./...`（`sys/02_backend`）— 全てパス
  （testcontainers-go が起動する使い捨て Postgres に対して新マイグレーション008を含めて実行）。
- `npm run build`（`sys/03_frontend/web`）— 型チェック含め成功。
- `npm run test`（Vitest, 同上）— 7ファイル28件全てパス。
- `npx playwright test --list`（`sys/04_e2e`）— 構文・import エラーなく全21件が一覧表示されることを確認。

## 計画からの逸脱・未実施事項

- **Playwright E2E は実行していない。** `playwright.config.ts` は「docker compose up 済みのスタックに
  対して実行する」前提（webServer 定義なし）だが、現在起動中の `01_infra` の Docker コンテナは今回の
  コード変更前のイメージのままで、ユーザーの作業中セッションでもあるため、無断でのイメージ再ビルド／
  コンテナ再起動は行わなかった。`docker compose up -d --build` でイメージを再ビルドしてから
  `npm test`（`sys/04_e2e`）を実行することを推奨。
- それ以外は計画どおりに実施し、逸脱はなし。
