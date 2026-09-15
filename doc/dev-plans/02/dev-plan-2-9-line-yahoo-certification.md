# 開発計画 2-9 — マーケティング講座・クイズ検定機能の廃止 と「LINEヤフー 認定資格勉強」の追加

`CLAUDE.md` の開発フロー（プラン → 実装指示 → 実装 + 結果ドキュメント作成 → git はユーザーが実施）に従う。本ステップは Phase 2（[dev-plan-2-0-overview.md](dev-plan-2-0-overview.md)）に追加する Step 2-9。

## 背景・決定事項

ユーザー指示: 学習コンテンツ（`/learn/`）ナビの「LINEマーケティング講座」「クイズ・検定」を削除し、「LINEヤフー　認定資格勉強」に置き換える。

確認の結果、スコープは以下の通り決定:

- **フル削除**: ナビだけでなく、`/learn/line-marketing`（講座・レッスン, Phase 1）と `/learn/quiz`（クイズ・検定, Phase 2）の配信ページ・管理画面・API・DB スキーマを含めて完全に削除する。
- **新設**: `/learn/quiz` の代わりに `/learn/line-yahoo-certification` を新設し、プレースホルダー（準備中）ページを配置する。コンテンツの中身は本ステップの対象外（将来の別ステップで決定）。
- Step 2-7（本番マイグレーション適用・本番デプロイ）は**未着手のまま不要になる**。クイズ機能は本番 DB に一度も適用されていないため、本ステップの新規マイグレーションが Step 2-1・2-7 を実質的に置き換える。[dev-plan-2-0-overview.md](dev-plan-2-0-overview.md) の該当行は本ステップ完了後に「キャンセル（2-9 で機能ごと削除）」に更新する。
- Phase 1 の講座・レッスン・試験（`courses` / `lessons` / `exams` / `exam_questions` / `exam_choices` / `user_lesson_progress` / `user_exam_results`）は本番 DB に適用済みの可能性が高い（Phase 1 で先行デプロイ済み）。削除は破壊的操作になるため、本番適用前に必ずユーザーに実施可否を確認する。

## スコープ外

- 会員管理・サロン予約・スイーツショップの3ミニアプリ。
- `/learn/line-operation`・`/learn/ai`（記事系, `articles`/`tags`/`usecases` テーブル）は無関係、変更しない。
- 「LINEヤフー 認定資格勉強」の実コンテンツ作成・データ投入（本ステップはプレースホルダーのみ）。

---

## 削除対象の棚卸し

### フロントエンド（`sys/03_frontend/web/src`）

**マーケティング講座（Phase 1: courses/lessons/exam）**
- `app/learn/line-marketing/page.tsx`, `app/learn/line-marketing/[lessonSlug]/page.tsx`
- `app/admin/courses/**`（`page.tsx`, `new/page.tsx`, `[id]/edit/page.tsx`, `[id]/lessons/new/page.tsx`）
- `app/admin/lessons/**`（`[id]/edit/page.tsx`, `[id]/exam/page.tsx`）
- `components/learn/CourseLessonList.tsx`
- `components/learn/LessonInteractive.tsx`, `LessonInteractive.test.tsx`
- `app/learn/me/page.tsx` のコース進捗表示部分（`progress.courses` を描画しているブロック、`GET /api/me/progress` 呼び出し）— マイページ自体は残すが、クイズ以外の進捗表示ロジックを削除・簡略化する

**クイズ・検定（Phase 2: quiz）**
- `app/learn/quiz/page.tsx`, `app/learn/quiz/[slug]/page.tsx`
- `app/admin/quizzes/**`（`page.tsx`, `new/page.tsx`, `[id]/edit/page.tsx`）
- `components/quiz/**`（`ExamRunner.tsx`+test, `QuizModeSelect.tsx`, `QuizQuestionCard.tsx`+test）
- `components/mypage/QuizHistoryPanel.tsx`, `QuizProgressSection.tsx`+test
- `app/learn/me/page.tsx` の `<QuizProgressSection />` 呼び出し

**ナビ・一覧の更新（削除ではなく編集）**
- `components/Header.tsx`: `learnLinks` から `line-marketing` と `quiz` を削除し、新規リンク（`/learn/line-yahoo-certification`, ラベル「LINEヤフー　認定資格勉強」）を追加
- `app/learn/page.tsx`: `categories` 配列を同様に更新、`metadata.description` の文言も見直す
- `app/admin/layout.tsx`: `navLinks` から `/admin/courses`（「講座・レッスン」）と `/admin/quizzes`（「クイズ・検定」）を削除

**新設**
- `app/learn/line-yahoo-certification/page.tsx`: 準備中プレースホルダーページ（タイトル「LINEヤフー 認定資格勉強」、"近日公開" 等の文言、`metadata` 設定）

### バックエンド（`sys/02_backend`）

**マーケティング講座・試験・進捗**
- `internal/handler/course.go`+test, `internal/repository/course.go`+test
- `internal/handler/exam.go`+test, `internal/service/exam.go`+test, `internal/repository/exam.go`+test
- `internal/handler/progress.go`, `internal/repository/progress.go`+test

**クイズ**
- `internal/handler/quiz.go`+test, `internal/handler/admin_quiz.go`+test
- `internal/service/quiz.go`, `internal/service/quiz_scoring.go`+test
- `internal/repository/quiz.go`+test

**ルーティング**
- `internal/server/server.go`: 上記ハンドラの生成・route 登録をすべて削除
  - 公開: `GET /courses`, `/courses/:slug`, `/courses/:slug/lessons/:lessonSlug`, `/quizzes`, `/quizzes/:slug`
  - 管理: `/admin/courses*`, `/admin/lessons*`, `/admin/quizzes*`, `/admin/quiz-questions*`
  - Reader: `/lessons/:lessonId/exam*`, `/lessons/:lessonId/complete`, `/courses/:slug/progress`, `/me/progress`, `/quiz-questions/:id/answer`, `/quizzes/:slug/submit`, `/me/quizzes/**`

### DB マイグレーション

新規マイグレーションを追加（既存の `007`/`008` ファイルは削除しない — 履歴として残し、down 方向で取り消す）:

- `009_drop_quiz.up.sql` / `.down.sql`: `user_quiz_answers`, `user_quiz_attempts`, `quiz_choices`, `quiz_questions`, `quizzes` を DROP（down で再作成、`008`/`007` の CREATE 文を復元）
- `010_drop_course_exam.up.sql` / `.down.sql`: `user_exam_results`, `user_lesson_progress`, `exam_choices`, `exam_questions`, `exams`, `lessons`, `courses` を DROP（down で再作成、`002`/`003`/`005` の該当 CREATE/ALTER 文を復元）
  - ⚠️ 本番 DB にデータが入っている場合、この DROP は破壊的。本番適用前に必ずユーザー確認を取る。
- `migrations/seed.sql`: `courses`/`lessons` への INSERT 文（line 9-16）を削除

### テスト

- Go: 上記ハンドラ・サービス・リポジトリの `*_test.go` を削除。`repository/progress_test.go` も対象。
- Vitest: `LessonInteractive.test.tsx`, `QuizProgressSection.test.tsx`, `ExamRunner.test.tsx`, `QuizQuestionCard.test.tsx`, `app/learn/me/page.test.tsx`（コース進捗部分の期待値を更新 or 削除）
- Playwright (`sys/04_e2e/tests`): `quiz.spec.ts` を削除。`admin-content.spec.ts`, `auth-and-progress.spec.ts`, `public-content.spec.ts`, `tests/helpers/db.ts` から講座・レッスン・クイズ関連のシナリオ/ヘルパーを削除し、必要なら `line-yahoo-certification` プレースホルダーの簡単な疎通テストを追加

### ドキュメント

- `README.md`: 「Content Site Map」テーブルの `line-marketing` 行を `line-yahoo-certification` に置き換え、Nav order の説明文を更新
- `doc/dev-plans/02/dev-plan-2-0-overview.md`: Step 2-8 を表に追加、Step 2-7 を「キャンセル（2-8 で機能ごと削除）」に更新
- `sys/03_frontend/web/doc/seo-update-2026-02-16.md` に `line-marketing`/`quiz` の URL 記載があれば確認し、影響があれば追記（新規 URL の sitemap 反映は別ステップでも可）
- `public/sitemap.xml`: `line-marketing`/`quiz` 系 URL が列挙されていれば `line-yahoo-certification` に置き換え

---

## タスク一覧

| # | 内容 |
|---|---|
| 8.1 | DB マイグレーション `009_drop_quiz.*`, `010_drop_course_exam.*` 追加、`seed.sql` 更新。ローカル DB で `up`/`down` 往復を確認 |
| 8.2 | バックエンド: quiz/course/exam/progress のハンドラ・サービス・リポジトリ・テストを削除、`server.go` の route 登録を削除。`go build ./...` / `go vet ./...` / `go test ./...` が通ることを確認 |
| 8.3 | フロントエンド: `/learn/line-marketing`, `/learn/quiz`, `/admin/courses`, `/admin/lessons`, `/admin/quizzes` の各ページ・関連コンポーネント・テストを削除 |
| 8.4 | フロントエンド: `Header.tsx`, `app/learn/page.tsx`, `app/admin/layout.tsx` のナビ/一覧を更新（削除 + 新規リンク追加） |
| 8.5 | フロントエンド: `app/learn/line-yahoo-certification/page.tsx` プレースホルダーページを新規作成 |
| 8.6 | フロントエンド: `app/learn/me/page.tsx` からコース進捗・クイズ進捗の表示を削除し、マイページの残す内容を整理 |
| 8.7 | E2E/Vitest: 該当テストの削除・更新（`quiz.spec.ts` 削除、他 spec からシナリオ削除、`db.ts` ヘルパー整理） |
| 8.8 | ドキュメント更新: `README.md`, `dev-plan-2-0-overview.md`, 必要なら SEO doc / sitemap |
| 8.9 | 動作確認: `npm run build`（型チェック含む）、`go test ./...`、Vitest、Playwright、実ブラウザでナビ・`/learn`・`/learn/line-yahoo-certification`・管理画面を確認 |

## 依存関係

8.1 → 8.2 → (8.3, 8.4, 8.5, 8.6 は並行可) → 8.7 → 8.8 → 8.9

## 未決定事項（実装前にユーザー確認が必要）

1. **本番 DB への影響**: Phase 1（講座・レッスン・試験）は本番適用済みの可能性が高い。`010_drop_course_exam` を本番に適用してよいか、既存の受講・進捗データを失ってよいか、実装前に要確認。
2. `/learn/line-yahoo-certification` プレースホルダーページの文言・公開時期の目安表示の要否。
3. `app/learn/me/page.tsx`（マイページ）はコース進捗・クイズ進捗を両方失うと表示するものがほぼ無くなる。ページ自体を維持する（空状態のメッセージのみ）か、ナビからも一旦外すか。
4. 管理画面ダッシュボード（`/admin`）に courses/quizzes への参照や件数表示があれば別途確認要（未調査）。

## 検証チェックリスト（実装後）

- [ ] `npm run build`（`sys/03_frontend/web`）が成功する
- [ ] `go build ./...` / `go vet ./...` / `go test ./...`（`sys/02_backend`）が成功する
- [ ] Vitest: `npm test` が成功する
- [ ] Playwright: `npx playwright test` が成功する
- [ ] ローカル DB でマイグレーション `up` → `down` → `up` が通る
- [ ] 実ブラウザ: デスクトップ/モバイルナビに「LINEヤフー　認定資格勉強」が表示され、`/learn/line-marketing`・`/learn/quiz`・`/admin/courses`・`/admin/lessons`・`/admin/quizzes` が404になる
- [ ] `/learn/line-yahoo-certification` がプレースホルダー表示される
- [ ] 管理画面ナビから「講座・レッスン」「クイズ・検定」が消えている
