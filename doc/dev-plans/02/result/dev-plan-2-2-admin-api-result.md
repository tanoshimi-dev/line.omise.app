# Step 2-2 実装結果 — クイズ／検定 管理 API（Admin CRUD）

**対応プラン:** [dev-plan-2-2-admin-api.md](../dev-plan-2-2-admin-api.md)

---

## 実施内容

### 2-2.1 クイズ／検定本体

- `internal/repository/quiz.go` — `Quiz` / `QuizQuestion` / `QuizChoice` 型と `QuizRepository`
  （`ListAll` / `GetByID` / `Create` / `Update` / `Delete` / `ListQuestions` / `ListChoicesForQuiz` /
  `CreateQuestion` / `UpdateQuestion` / `DeleteQuestion`）
  - `PassingScore` は `*int`（NULL 許容）。既存の `description`/`reference_url` のような
    「COALESCE で空文字に丸める」パターンとは異なり、NULL＝「合否判定なし」という意味を持つため
    ポインタ型のまま往復させる方針にした（`repository/quiz_test.go` で NULL/値どちらの
    ラウンドトリップも検証済み）
- `internal/handler/admin_quiz.go` — `AdminQuizHandler`
  - `GET /api/admin/quizzes`、`POST /api/admin/quizzes`、`PUT /api/admin/quizzes/:id`、`DELETE /api/admin/quizzes/:id`
  - プラン本文に明記はなかったが、編集フォームのプリフィル（2-2.4 完了条件「編集できる」）に必要なため
    `GET /api/admin/quizzes/:id`（設問・選択肢を埋め込んだ詳細取得。`AdminGetCourse`/`AdminGetExamByLesson` と同じ形）を追加

### 2-2.2 設問・選択肢

- `POST /api/admin/quizzes/:quizId/questions`、`PUT /api/admin/quiz-questions/:id`、`DELETE /api/admin/quiz-questions/:id`
- `UpdateQuestion` は既存の `exam_questions` と同じ「選択肢は全置換」方式（PUT = リソース全体の置き換え）

### 2-2.3 バリデーション

`internal/service/quiz.go` に集約:
- `ValidateQuizMode` — `mode` が `practice`/`exam` 以外なら `400`
- `ValidatePassingScore` — `practice` モードで `passing_score` が指定されていたら `400`
- `ValidateQuestionChoices` — 選択肢2つ未満／正解0件／`allow_multiple=false` で正解2件以上、をそれぞれ `400`
- slug 重複は既存の `service.ErrDuplicateSlug`/`AsDuplicateSlug`（Postgres unique violation 判定、制約名を問わない既存実装）をそのまま再利用 → `409`

### 2-2.4 管理画面 UI

- `src/app/admin/quizzes/page.tsx` — 一覧（タイトル・スラッグ・種別〈クイズ/検定〉・公開状態・削除）
- `src/app/admin/quizzes/new/page.tsx` — 新規作成フォーム（`mode` 切り替えで `passing_score` 欄の表示を出し分け）
- `src/app/admin/quizzes/[id]/edit/page.tsx` — 編集フォーム＋設問管理
  （`lessons/[id]/exam/page.tsx` の構成を踏襲し、`allow_multiple` トグル・`explanation`・`reference_url` の入力を追加）
- `src/app/admin/layout.tsx` のナビゲーションに「クイズ・検定」を追加
- `src/lib/types.ts` に `QuizMode` / `AdminQuizChoice` / `AdminQuizQuestion` / `AdminQuiz` を追加
  （Phase 1 の `Exam`/`AdminExam` とは別型として定義し、混同を防止）

### ルーティング・テスト基盤

- `internal/server/server.go` — `quizRepo` / `adminQuizHandler` を配線し、`/api/admin/quizzes*` と
  `/api/admin/quiz-questions/:id` を `requireAdmin` 配下に登録
- `internal/testutil/db.go` — `TruncateAll` の対象テーブルに `quizzes` 系5テーブルを追加

---

## プラン外で追加対応したこと

- `GET /api/admin/quizzes/:id`（前述、2-2.4 の編集 UI に必須なためプランへの自然な補完として追加）

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- 実ブラウザでの管理画面の動作確認 — 本番同等の確認には LINE/Google の実ログインで Admin セッションを
  取得する必要があり（Phase 1 Step 11 の際も同様にユーザー自身の Google アカウントで確認した実績あり）、
  本 Step では代わりに以下で検証: (a) Admin/Reader/未ログインの認可を含むハンドラテスト12件、
  (b) `npm run build` によるフロントエンドの型検証（新規ページ含め成功）。実ブラウザでの操作確認が必要な場合は
  別途 Admin アカウントでの確認をお願いしたい。
- `next lint`（`npm run lint`）— 本プロジェクトには ESLint の設定ファイル（`eslint.config.*`）が存在せず、
  `package.json` の `lint` スクリプト自体が現状のプロジェクトで動作しない状態だった（本 Step 開始前からの
  既存事象、今回の変更が原因ではない）。修正は本 Step のスコープ外と判断し着手していない。

---

## 検証

- `go build ./...` / `go vet ./...` — 成功
- `gofmt -l .` — 差分なし
- `go test ./...` — 全パッケージ成功
  - `internal/repository`: 新規 `quiz_test.go` 9件（重複slug、`passing_score` の NULL/値ラウンドトリップ、
    設問・選択肢の作成/更新/削除とカスケード）含め全成功
  - `internal/handler`: 新規 `admin_quiz_test.go` 12件（未ログイン401、Reader403、Admin作成成功、
    `mode`/`passing_score` バリデーション、選択肢バリデーション3種、詳細取得での `is_correct` 混入確認、
    削除カスケード確認）含め全成功
  - 既存テスト（Phase 1 分）に regression なし
- フロントエンド: `npm run build` 成功（`/admin/quizzes`, `/admin/quizzes/new`, `/admin/quizzes/[id]/edit` を
  含む全ルートの型検証・ビルドが通過）、`npm test`（Vitest）15件成功（regression なし）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| Admin が練習モードのクイズと検定モードの試験（`passing_score` 付き）の両方を作成・編集・削除できる | ✅ ハンドラテストで確認（実ブラウザでの確認は未実施、上記参照） |
| 不正なデータ（選択肢不足、正解フラグなし、`allow_multiple=false` での複数正解など）が `400` で弾かれる | ✅ |
| 管理画面から一連の操作が実際に行える | ✅ 画面は実装・型検証済み。実ブラウザでの Admin ログインを伴う動作確認は未実施 |

## 次のステップ

- [dev-plan-2-3-answer-scoring-api.md](../dev-plan-2-3-answer-scoring-api.md)（解答・採点 API／受験履歴・進捗 API）
