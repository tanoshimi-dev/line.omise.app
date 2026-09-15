# Step — クイズの解答モードをユーザー選択制にする（単発モード / 検定モード）

**対象:** Phase 2「クイズ・検定機能」の拡張（[dev-plan-2-0-overview.md](02/dev-plan-2-0-overview.md) 以降の設計を変更）
**依存:** Phase 2 全体（Step 2-1〜2-7）完了済み

---

## 背景・現状の課題

現状、`quizzes.mode`（`practice` | `exam`）は**クイズ 1 レコードにつき 1 つの固定値**で、
以下の各所でこの値が「解答方式そのものを固定するゲート」として使われている：

- `sys/02_backend/internal/handler/quiz.go` の `AnswerQuestion`（1問ずつ即採点）は
  `quiz.Mode != "practice"` なら 400 を返す
- 同じく `SubmitQuiz`（まとめて解答・最終スコア）は `quiz.Mode != "exam"` なら 400 を返す
- フロントエンド `sys/03_frontend/web/src/app/learn/quiz/[slug]/page.tsx:43` は
  `quiz.mode === 'exam' ? <ExamRunner> : <QuizQuestionCard>` で機械的にコンポーネントを振り分けている
- `/learn/quiz` 一覧ページはクイズを `mode` で「クイズ（1問ずつ解答・即採点）」「検定（まとめて解答・最終スコア）」の
  2 セクションに分けて表示している
- マイページの履歴・進捗 API（`GetPracticeHistory` / `ListAttempts` / `GetQuizProgress` /
  `GetMyQuizzesProgress`）も `quiz.Mode` で分岐し、練習履歴か検定履歴のどちらか一方の形状しか返さない

今回、**同じクイズ（同じ設問セット）に対して、ユーザーが解答開始時に「単発モード」（1問ずつ即採点）か
「検定モード」（まとめて解答・最後にまとめて採点）かを選べるようにする**。

### 決定済みの方針（ユーザー確認済み・前回案から修正）

- **admin はクイズの種別（練習／検定）を設定しない。** `quizzes.mode` は管理画面からも一覧の分類からも
  完全に廃止する。既存の「クイズ」レコードと「検定」レコードを統合する必要はない
  （両者はもともと別トピックの別レコードであり、内容をマージするわけではない）— 単に**どのレコードも
  種別を持たなくなり**、一覧は単一リストになる。
- **各ユーザーが、解答を始める前に毎回モードを選ぶ。** admin 設定によるデフォルト／おすすめ表示もしない
  （選択の初期値としても `mode` を参照しない）。
- **マイページの履歴・進捗は「クイズごとに両方のセクションを表示」** する（統合タイムラインにはしない）。
  そのクイズに単発解答の記録があれば単発履歴セクションを、検定受験の記録があれば検定履歴セクションを、
  データがある方だけ（あるいは両方）表示する。

データモデルの解答記録側（`user_quiz_answers.attempt_id IS NULL` = 単発解答、`user_quiz_attempts` 経由 = 検定受験）は
そのまま流用できるが、**`quizzes.mode` 列自体は不要になるため DB マイグレーションで削除する**
（`sys/02_backend/migrations/007_quiz.up.sql:6-7` で `NOT NULL DEFAULT 'practice' CHECK (mode IN ('practice','exam'))`
として定義されている列。`passing_score` とのクロスカラム CHECK 制約は無いため、削除は `mode` 列単体で完結する）。

---

## タスク

### 1. DB マイグレーション — `quizzes.mode` 列を削除

- [ ] 新規マイグレーション（例: `008_quiz_drop_mode.up.sql` / `.down.sql`）を追加し、
      `quizzes` テーブルから `mode` 列（`NOT NULL DEFAULT 'practice' CHECK (mode IN ('practice','exam'))`,
      `sys/02_backend/migrations/007_quiz.up.sql:6-7`）を削除する
  - `passing_score` は `mode` とのクロスカラム CHECK が無いことを確認済みのため、削除の影響を受けない
  - down マイグレーションでは列を `NOT NULL DEFAULT 'practice'` で復元する（既存データの実値までは復元不可）

### 2. バックエンド — Admin API: 種別（mode）設定を廃止、passing_score は独立項目に

`sys/02_backend/internal/handler/admin_quiz.go` / `internal/service/quiz.go` / `internal/repository/quiz.go`
（`Quiz.Mode` フィールド、`ValidateQuizMode`、Create/Update の request 構造体、[dev-plan-2-2-admin-api.md](02/dev-plan-2-2-admin-api.md) 参照）

- [ ] `repository.Quiz` 構造体から `Mode` フィールドを削除し、SELECT/INSERT/UPDATE 文から `mode` 列を除去
- [ ] Admin の Create/Update リクエスト構造体・ハンドラから `mode` パラメータを削除
- [ ] `ValidateQuizMode` を削除（呼び出し元が無くなるため）
- [ ] `passing_score` の入力を「exam モードのクイズのみ」という制約から解放し、どのクイズにも
      任意で合格点を設定できるようにする（検定モードで解答された際に合否判定を出せるようにするため）

### 3. バックエンド — 解答エンドポイントのモードゲート除去

`sys/02_backend/internal/handler/quiz.go`

- [ ] `AnswerQuestion`: `if quiz.Mode != "practice" { 400 }` のチェックを削除し、どのクイズの設問でも
      1問ずつ即採点で解答できるようにする
- [ ] `SubmitQuiz`: `if quiz.Mode != "exam" { 400 }` のチェックを削除し、どのクイズでもまとめて解答・
      最終スコア算出ができるようにする
- [ ] `service/quiz_scoring.go` の `ErrQuizModeMismatch`（現状どこからも呼ばれていないゲート用センチネル）が
      他で未使用であることを確認し、不要であれば削除
- [ ] `GradeQuizExam` が `passing_score = NULL`（合格点未設定のクイズ）でも安全に動作すること
      （`Passed` は `nil` を返し、フロントは合否バッジを出さずスコアのみ表示）を確認・必要ならテスト追加

### 4. バックエンド — 履歴・進捗 API をクイズごと両方の形状で返す

`sys/02_backend/internal/handler/quiz.go`（`GetPracticeHistory` / `ListAttempts` / `GetQuizProgress` /
`GetMyQuizzesProgress`, 現状 `quiz.Mode` で分岐しているロジック一式）

- [ ] `GET /me/quizzes/:slug/history`（単発履歴）: 分岐条件を撤廃し、そのクイズに
      `attempt_id IS NULL` の解答記録があれば返す（無ければ空配列）
- [ ] `GET /me/quizzes/:slug/attempts`（検定受験一覧）: 同様に分岐条件を撤廃し、
      そのクイズの `user_quiz_attempts` があれば返す（無ければ空配列）
- [ ] `GET /me/quizzes/:slug/progress`・`GET /me/quizzes/progress`: レスポンス形状を
      「単発系フィールド（`answered_count`/`correct_count`）」と
      「検定系フィールド（`attempt_count`/`best_score`/`latest_score`/`latest_passed`/`passing_score`）」
      の両方を含む形に統合（データが無い方は `0`/`null`）。既存の mode 分岐によるレスポンス形状の
      切り替えを廃止
- [ ] 上記レスポンス形状変更に伴い、`sys/03_frontend/web/src/lib/types.ts` の
      `QuizProgress` 相当の型を更新

### 5. フロントエンド — Admin 画面から種別（mode）UI を削除

`sys/03_frontend/web/src/app/admin/quizzes/new/page.tsx` / `[id]/edit/page.tsx` / `page.tsx`

- [ ] `new/page.tsx`・`edit/page.tsx` の `mode` の `useState`・select 入力・送信ペイロードへの
      `mode` フィールドを削除
- [ ] `passing_score` 入力欄の表示条件だった `mode === 'exam'` を削除し、常に（任意項目として）表示する
- [ ] 一覧 `page.tsx:49` の `{quiz.mode === 'exam' ? '検定' : 'クイズ'}` 列を削除
      （種別という概念自体が無くなるため）
- [ ] `sys/03_frontend/web/src/lib/types.ts` の `AdminQuiz`/`QuizMode` 関連の型から `mode` を削除

### 6. フロントエンド — 公開クイズ一覧を単一リストに統合

`sys/03_frontend/web/src/app/learn/quiz/page.tsx`

- [ ] `data.quizzes.filter((q) => q.mode === 'practice')` / `'exam'` による2セクション分割を廃止し、
      単一のクイズ一覧として表示する（見出し文言「クイズ（1問ずつ解答・即採点）」「検定（まとめて解答・
      最終スコア）」も統合・削除）

### 7. フロントエンド — クイズ詳細ページにモード選択 UI を追加

`sys/03_frontend/web/src/app/learn/quiz/[slug]/page.tsx`

- [ ] `quiz.mode === 'exam' ? <ExamRunner> : <QuizQuestionCard>` の機械的な振り分けを廃止
- [ ] 新規クライアントコンポーネント（例: `QuizModeSelect`）を追加し、解答開始前に必ず
      「単発モード（1問ずつ解答・即採点）」「検定モード（まとめて解答・最終スコア）」を選ばせる
      （admin 側のデフォルト値は存在しないため、初期状態はどちらも選択されていない状態にする）
- [ ] 選択結果に応じて同じ `quiz` オブジェクト（設問データ）を `QuizQuestionCard` または `ExamRunner` に渡す
- [ ] 「合格点: N%」表示（現状 `quiz.mode === 'exam' && quiz.passing_score != null` の条件）を
      `quiz.passing_score != null`（かつ検定モード選択時）に変更 — 合格点が設定されているクイズであれば
      検定モードで合否表示されるようにする

`sys/03_frontend/web/src/components/quiz/QuizQuestionCard.tsx` / `ExamRunner.tsx`

- [ ] 各コンポーネントが `quiz.mode` を前提にした分岐を内部に持っていないか確認し、あれば
      呼び出し元から渡された選択モードベースに揃える

### 8. フロントエンド — マイページ履歴・進捗を両セクション表示に

[dev-plan-2-5-frontend-mypage.md](02/dev-plan-2-5-frontend-mypage.md) で実装したマイページの
クイズ履歴・進捗コンポーネント（`quiz.mode` で練習履歴 or 検定履歴のどちらか一方を出し分けている箇所）を、
「データがある方（または両方）を両方表示」に変更する。対象コンポーネントは実装時に
`GetPracticeHistory`/`ListAttempts`/`GetQuizProgress` を呼んでいる箇所から特定する。

- [ ] クイズごとに、単発履歴セクション（データがあれば）と検定受験履歴セクション（データがあれば）を
      両方表示できるようにする
- [ ] 進捗表示（Step 4 で統合したレスポンス形状）に合わせて、単発系・検定系の指標を両方表示する

---

## 成果物

- `sys/02_backend/migrations/008_quiz_drop_mode.up.sql` / `.down.sql`
- `sys/02_backend/internal/repository/quiz.go`（`Mode` フィールド除去）
- `sys/02_backend/internal/handler/quiz.go` / `admin_quiz.go`（モードゲート除去、履歴・進捗レスポンス統合、
  Admin リクエストから `mode` 除去）
- `sys/02_backend/internal/service/quiz.go` / `quiz_scoring.go`（`ValidateQuizMode` 削除、passing_score 制約緩和、
  不要センチネル整理）
- `sys/03_frontend/web/src/app/admin/quizzes/new/page.tsx` / `[id]/edit/page.tsx` / `page.tsx`
  （種別 UI・列の削除）
- `sys/03_frontend/web/src/app/learn/quiz/page.tsx`（単一リスト化）
- `sys/03_frontend/web/src/app/learn/quiz/[slug]/page.tsx` ＋ 新規モード選択コンポーネント
- `sys/03_frontend/web/src/components/quiz/QuizQuestionCard.tsx` / `ExamRunner.tsx`（必要に応じて）
- マイページのクイズ履歴・進捗コンポーネント
- `sys/03_frontend/web/src/lib/types.ts`（`Quiz`/`QuizListItem`/`AdminQuiz` から `mode` 除去、進捗レスポンス型の更新）

## 完了条件

- admin 画面にクイズの種別（練習／検定）を設定する項目が存在しない
- どのクイズを開いても、解答開始前に単発モード／検定モードを選択でき、両方で解答を開始できる
- 単発モードは 1 問ずつ即座に正誤・解説を表示し、検定モードは全問解答後にまとめてスコア・合否
  （`passing_score` 設定時）を表示する — 挙動は既存の `QuizQuestionCard`／`ExamRunner` のロジックを流用
- マイページで、あるクイズについて単発モードの解答履歴と検定モードの受験履歴が両方記録されている場合、
  両方のセクションが表示される
- 他ユーザーの履歴・進捗にアクセスできないこと等、[dev-plan-2-6-test.md](02/dev-plan-2-6-test.md) で
  確認済みの認可要件がデグレしていないこと（既存テストの再実行 + 新規分岐のテスト追加）
