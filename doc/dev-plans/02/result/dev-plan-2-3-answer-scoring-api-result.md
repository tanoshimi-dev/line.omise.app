# Step 2-3 実装結果 — 解答・採点 API／受験履歴・進捗 API

**対応プラン:** [dev-plan-2-3-answer-scoring-api.md](../dev-plan-2-3-answer-scoring-api.md)

---

## 実施内容

### 2-3.1 公開取得

- `GET /api/quizzes/:slug`（`QuizHandler.GetQuiz`、ログイン不要）
  - `QuizRepository.GetPublishedBySlug` で未公開クイズは `ErrNotFound`（404）
  - レスポンスは `mode`・`passing_score` を含むが、`is_correct`・`explanation`・`reference_url` は一切含めない

### 2-3.2 練習モード（`mode = practice`）

- `POST /api/quiz-questions/:id/answer`（`QuizHandler.AnswerQuestion`）
  - 対象設問が属するクイズが `exam` モードなら `400`
  - 未公開クイズの設問は `404`（`quiz.Published` を確認）
  - `choice_ids` が空なら `400`
  - 採点は `service.GradeQuizQuestion`（選択セットの完全一致のみ正解、`allow_multiple=false` で複数選択は `400`、無関係の choice_id は `400`）
  - ログイン中のみ `user_quiz_answers`（`attempt_id = NULL`）に保存

### 2-3.3 検定モード（`mode = exam`）

- `POST /api/quizzes/:slug/submit`（`QuizHandler.SubmitQuiz`）
  - 対象クイズが `practice` モードなら `400`
  - 採点は `service.GradeQuizExam`（設問ごとに `GradeQuizQuestion` を適用し集計）
  - **未回答の設問は不正解扱い**（プランの決定事項どおり、`400` にはしない）
  - `score` は正答数（生の件数）、`passed` は `passing_score`（0〜100 のパーセンテージとして解釈）に対して
    `round(score/total*100) >= passing_score` で判定 — `passing_score` 未設定なら `passed` は `null`
  - ログイン中のみ `user_quiz_attempts` を1件作成し、各設問の解答を `attempt_id` 付きで `user_quiz_answers` に保存
  - レスポンスに設問ごとの正誤・正しい選択肢・解説を含め、提出後の振り返りに使う

### 2-3.4 受験履歴・進捗

- `GET /api/me/quizzes/:slug/history` — 練習モードの解答履歴（`attempt_id IS NULL` の行、新しい順）。`exam` モードのクイズに対しては `400`
- `GET /api/me/quizzes/:slug/attempts` — 検定モードの受験履歴一覧。`practice` モードのクイズに対しては `400`
- `GET /api/me/quizzes/:slug/attempts/:attemptId` — 1回分の受験詳細（設問ごとの正誤・解説）。
  `GetOwnAttempt(attemptID, userID)` で所有者チェックをクエリ自体に組み込み、他ユーザーの attempt は
  「存在しない」場合と同じ `404` を返す（存在の有無を漏らさない）
- `GET /api/me/quizzes/:slug/progress` — モード別の進捗（practice: 解答済み設問数・正答率〈**設問ごとの最新解答**を基準に集計〉、
  exam: 受験回数・最高得点・直近得点・直近合否）
- `GET /api/me/quizzes/progress` — 全公開クイズ／検定のサマリー（`ProgressHandler.GetMyProgress` と同様、未受験でも一覧に含む）

### 採点ロジック（`internal/service/quiz_scoring.go`）

- `GradeQuizQuestion` — 選択肢セットの完全一致判定（部分点なし）。空選択は「不正解」（エラーではない）として扱い、
  検定モードの一括採点が「未回答は不正解」を実現できるようにした
- `GradeQuizExam` — 全設問を `GradeQuizQuestion` で採点し、`score`/`total_questions`/`passed`（`passing_score` 設定時のみ）を算出

### リポジトリ・ルーティング

- `internal/repository/quiz.go` に追加: `ListPublished`、`GetPublishedBySlug`、`GetQuestionByID`、`ListChoicesForQuestion`、
  `UserQuizAttempt`/`UserQuizAnswer` 型、`SaveAnswer`、`CreateAttempt`、`ListPracticeHistory`、`ListAttempts`、
  `GetOwnAttempt`、`ListAnswersForAttempt`
- `internal/middleware/auth.go` に `OptionalUser` を追加 — 有効なセッションがあれば `CurrentUser` で取得できるように
  コンテキストへセットするが、未ログインでも `401` にせず処理を継続させる（練習/検定の「誰でも解答可・ログイン時のみ保存」を実現）
- `internal/server/server.go` に `/api/quizzes/:slug`（公開）、`/api/quiz-questions/:id/answer`・`/api/quizzes/:slug/submit`
  （`optionalUser`）、`/api/me/quizzes/*`（`requireReader`）を配線

---

## プラン外で追加対応したこと

なし。

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- フロントエンド UI（練習/検定の受験画面、マイページの履歴・進捗表示） — Step 2-4・2-5 で実装
- 実ブラウザでの動作確認 — 本 Step は API のみのため、Step 2-4（フロントエンド実装）で UI と合わせて確認する

---

## 検証

- `go build ./...` / `go vet ./...` — 成功
- `gofmt -l .` — 差分なし
- `go test ./...` — 全パッケージ成功
  - `internal/service`: 新規 `quiz_scoring_test.go` 10件
    （単一/複数選択の正誤判定、未回答=不正解、`allow_multiple=false` での複数選択拒否、
    無関係choice_idの拒否、複数選択の完全一致要求、検定一括採点のスコア/合否、未回答問題の扱い、
    存在しない設問参照の拒否、合格点の境界値）
  - `internal/handler`: 新規 `quiz_test.go` 13件
    - 公開GETが `is_correct`/`explanation`/`reference_url` を含まないこと、未公開クイズが404
    - 練習モード: 匿名解答は結果を返すが保存されない（`user_quiz_answers` 0件を確認）、
      ログイン済み解答は保存されマイページ履歴から取得できる、検定モード設問への誤アクセスは400
    - 検定モード: 練習モードクイズへのsubmitは400、匿名submitは結果を返すが保存されない
      （`user_quiz_attempts` 0件を確認、`attempt_id` は `null`）、ログイン済みsubmitは
      attempt保存・`/attempts` 一覧・`/attempts/:id` 詳細から取得できる、他ユーザーのattemptは404
    - 進捗: 練習モードは最新解答基準で集計されること、検定モードは受験回数・最高/直近スコア・合否、
      全体サマリーは未受験のクイズも一覧に含むこと
    - 未ログインでの履歴取得は401
  - 既存テスト（Phase 1・Step 2-1・2-2 分）に regression なし

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 練習モード: 未ログインでも解答し、その場で正誤と解説が確認できる | ✅ |
| 検定モード: 全設問提出後に最終スコア（・合否）が返り、設問ごとの正誤・解説を振り返れる | ✅ |
| ログインユーザーの解答・受験結果が保存され、履歴・進捗 API から取得できる | ✅ |
| 正解フラグ・解説が解答前／提出前のレスポンスに含まれない | ✅ |
| 他ユーザーの履歴・進捗・受験詳細にアクセスできない | ✅ |

## 次のステップ

- [dev-plan-2-4-frontend-quiz-ui.md](../dev-plan-2-4-frontend-quiz-ui.md)（`/learn/quiz/` 練習モード・検定モード受験 UI）
