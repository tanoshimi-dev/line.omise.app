# Step 2-3 — 解答・採点 API／受験履歴・進捗 API

**フェーズ:** Phase 2 — クイズ・検定機能
**依存:** Step 2-1（DB マイグレーション）、Step 2-2（Admin API でクイズ/検定が作成できること）

---

## ゴール

- 練習モード（`mode = practice`）: 1問ずつ解答し、その場で正誤・解説を返す。
- 検定モード（`mode = exam`）: 全設問の解答を一括提出し、最終スコア（正答数・正答率、任意で合否）を返す。
- ログインユーザーの解答・受験結果を保存し、受験履歴・進捗を取得する API を実装する。

---

## タスク

### 2-3.1 公開取得（誰でも利用可）

- [ ] `GET /api/quizzes/:slug` — クイズ／検定と設問一覧を取得
  - **正解フラグ・解説は含めない**レスポンス（カンニング防止）
  - `mode` をレスポンスに含め、フロントエンドが練習/検定 UI を出し分けられるようにする

### 2-3.2 練習モード（`mode = practice`）— 1問ずつ即時採点

- [ ] `POST /api/quiz-questions/:id/answer` — 選択した選択肢ID（複数可）を送信
  - 採点をバックエンドで実施し、`is_correct` / `explanation` / `reference_url` / 正しい選択肢を返す
  - ログイン中は `user_quiz_answers`（`attempt_id = NULL`）に保存、未ログインなら保存せず結果のみ返す
  - 対象の設問が `mode = exam` のクイズに属する場合は `400`（練習エンドポイントの誤用を防止）

### 2-3.3 検定モード（`mode = exam`）— 一括提出・最終スコア

- [ ] `POST /api/quizzes/:slug/submit` — 全設問の解答をまとめて送信
  - リクエスト例: `{ "started_at": "...", "answers": [{ "question_id": 1, "choice_ids": [2] }, ...] }`
  - 未回答の設問がある場合の扱いを決定し実装（不正解扱い、または `400`）
  - 全問採点し `score`（正答数）・`total_questions`・`passed`（`passing_score` 設定時のみ）を算出
  - ログイン中は `user_quiz_attempts` を1件作成し、各設問の解答を `attempt_id` 付きで `user_quiz_answers` に保存
  - 未ログインは保存せず、結果（スコア・設問ごとの正誤・解説）のみ返す
  - レスポンスに設問ごとの正誤・正しい選択肢・解説を含め、提出後の振り返り表示に使う
  - 対象クイズが `mode = practice` の場合は `400`

### 2-3.4 受験履歴・進捗（ログイン必須）

- [ ] `GET /api/me/quizzes/:slug/history` — 練習モードの解答履歴（設問ごとの正誤・解答日時）
- [ ] `GET /api/me/quizzes/:slug/attempts` — 検定モードの受験履歴一覧（受験日時・スコア・合否）
- [ ] `GET /api/me/quizzes/:slug/attempts/:attemptId` — 検定モードの1回分の受験詳細（設問ごとの正誤・解説）
- [ ] `GET /api/me/quizzes/:slug/progress` — 当該クイズ/検定の進捗
  - practice: 解答済み設問数・正答率
  - exam: 受験回数・最高スコア・直近スコア・合否
- [ ] `GET /api/me/quizzes/progress` — 全クイズ／検定の進捗サマリー（マイページ一覧用）

### 2-3.5 不正防止・整合性

- [ ] 正解フラグ・解説は解答前／提出前のレスポンスに含めない
- [ ] 存在しない `quiz`/`question`/`choice`/`attempt` への参照は `404`
- [ ] 他ユーザーの履歴・進捗・受験詳細は取得できない（`user_id` はセッションから取得、リクエストパラメータで渡させない）
- [ ] `allow_multiple = false` の設問で複数選択肢が送られた場合は `400`
- [ ] 複数選択問題の採点基準を決定・実装（現時点の仮定: 選択セットの完全一致のみ正解）

---

## 成果物

- `internal/handler/quiz.go`（公開 API・解答・提出・履歴・進捗ハンドラ）
- `internal/repository/quiz.go`（`user_quiz_answers` / `user_quiz_attempts` 関連クエリ）
- `internal/service/quiz_scoring.go`（練習モード即時採点・検定モード一括採点の共通ロジック）

## 完了条件

- 練習モード: 未ログインでも解答し、その場で正誤と解説が確認できる
- 検定モード: 全設問提出後に最終スコア（・合否）が返り、設問ごとの正誤・解説を振り返れる
- ログインユーザーの解答・受験結果が保存され、履歴・進捗 API から取得できる
- 正解フラグ・解説が解答前／提出前のレスポンスに含まれない
- 他ユーザーの履歴・進捗・受験詳細にアクセスできない
