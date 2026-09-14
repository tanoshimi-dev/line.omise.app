# Step 06 — 試験（exam）API・学習進捗（progress）API

**フェーズ:** Phase 1 — CMS API
**依存:** Step 05（コンテンツ API）

---

## ゴール

レッスンに紐づく試験（選択式）の作成・受験機能と、ログインユーザーの学習進捗
（レッスン完了・試験結果）を保存・取得する API を実装する。

---

## タスク

### 6.1 試験の作成・管理（Admin）

- [ ] `POST /api/admin/lessons/:lessonId/exam` — 試験作成（タイトル・合格点）
- [ ] `POST /api/admin/exams/:examId/questions` — 設問追加（設問文・選択肢・正解フラグ）
- [ ] `PUT /api/admin/questions/:id` — 設問更新
- [ ] `DELETE /api/admin/questions/:id` — 設問削除

### 6.2 試験の受験（Reader、ログイン必須）

- [ ] `GET /api/lessons/:lessonId/exam` — 試験取得（**正解フラグは含めない**レスポンス）
- [ ] `POST /api/lessons/:lessonId/exam/submit` — 解答送信
  - リクエスト: `{ "answers": [{ "question_id": 1, "choice_id": 3 }, ...] }`
  - 採点はバックエンドで実施し、`score` と `passed`（合格点以上か）を返す
  - `user_exam_results` に結果を保存

### 6.3 学習進捗（Reader、ログイン必須）

- [ ] `POST /api/lessons/:lessonId/complete` — レッスンを完了済みとしてマーク
  - `user_lesson_progress` に upsert
- [ ] `GET /api/courses/:slug/progress` — ログインユーザーの当該講座の進捗取得
  - 完了済みレッスン数 / 全レッスン数、各レッスンの完了状況、試験結果を含む
- [ ] `GET /api/me/progress` — ログインユーザーの全講座の進捗サマリー（任意、マイページ用）

### 6.4 不正防止・整合性

- [ ] 同じユーザーが同じ試験を再受験した場合の扱いを決定（最新結果で上書き / 履歴を残す）
- [ ] 存在しない `lesson_id` / `question_id` へのリクエストは `404`
- [ ] 他人の進捗は取得できないことを保証（`user_id` はセッションから取得、リクエストパラメータで渡させない）

---

## 成果物

- `internal/handler/exam.go`, `progress.go`
- `internal/repository/exam.go`, `progress.go`
- `internal/service/exam.go`（採点ロジック）

## 完了条件

- Admin が試験・設問を作成できる
- Reader が試験を受験し、採点結果が返る
- レッスン完了マークと進捗取得が動作する
- 正解フラグが受験前のレスポンスに含まれない（カンニング防止）
- 他ユーザーの進捗にアクセスできない
