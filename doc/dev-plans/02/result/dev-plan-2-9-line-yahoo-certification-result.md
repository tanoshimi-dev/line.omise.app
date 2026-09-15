# Step 2-9 実装結果 — マーケティング講座の廃止 と クイズ・検定の「LINEヤフー 認定資格勉強」への再ブランディング

**対応プラン:** [dev-plan-2-9-line-yahoo-certification.md](../dev-plan-2-9-line-yahoo-certification.md)
（作成時のファイル名は `dev-plan-2-8-...`。ユーザー側で `dev-plan-2-9-...` にリネームされた状態で実装指示を受けたため、
本結果ドキュメントおよび overview の対応も Step 2-9 として記録する。プラン本文中の内部タスク番号は
作成時のまま `8.1`〜`8.9` を使用している。）

---

## ⚠️ プラン方針からの重大な訂正（ユーザー指摘によるスコープ修正）

プラン本体は「マーケティング講座・クイズ検定機能の両方をフル削除し、`/learn/line-yahoo-certification`に
プレースホルダーを新設する」という内容で、下記「実施内容」8.1〜8.9 はその通りに一度実装・完了した。

その後ユーザーから「quiz/exam disappeared. I want quiz/exam feature kept as LINEヤフー認定資格勉強」との
指摘があり、**クイズ・検定（Phase 2）機能を削除する判断は誤りだった**ことが判明。プラン作成時のヒアリング
（クイズ機能を含めて「フル削除」と回答をもらった）が、実際の意図（クイズ・検定の機能自体は維持し、公開ラベル
とURLだけを「LINEヤフー認定資格勉強」に変更したい）と異なっていた。

このため、8.1〜8.9 完了後に追加で以下の是正作業を実施し、**最終的な状態は以下の通り**:

- **マーケティング講座（Phase 1: `courses`/`lessons`/`exams`/`exam_questions`/`exam_choices`/
  `user_lesson_progress`/`user_exam_results`、および関連コード）→ 削除のまま維持**
  （ユーザーはこちらについては指摘していないため）
- **クイズ・検定（Phase 2: `quizzes`/`quiz_questions`/`quiz_choices`/`user_quiz_attempts`/
  `user_quiz_answers`、および関連コード）→ 全面復元**。ただし:
  - 公開URLを `/learn/quiz` → `/learn/line-yahoo-certification` に変更（プレースホルダーは削除し、
    実際のクイズ一覧・受験UIをこのURLに移設）
  - 公開ラベルを「クイズ・検定」→「LINEヤフー　認定資格勉強」に変更（`/learn` トップページカード、
    Header ナビ、マイページのセクション見出し、管理画面ナビ、管理画面見出し）
  - バックエンドAPIパス（`/api/quizzes` 等）・DBテーブル名・Go/TSの内部識別子（`Quiz`, `quiz` 等）は
    変更していない（実装詳細であり、ユーザー要望はURL/ラベルの話と判断）
  - マイグレーション `009_drop_quiz.up/down.sql` は撤回（削除）。ローカルDBは一旦ドロップされていた
    `quizzes` 系テーブルを手動で再作成し復元。`010_drop_course_exam.*`（マーケティング講座側）はそのまま維持

是正作業の詳細は本ドキュメント末尾の「## 是正作業（クイズ・検定機能の復元）」セクションを参照。

---

## 実施内容（初回実装 — 事後に一部訂正）

プラン記載のタスク 8.1〜8.9 を全て実施。ほぼプラン通りだが、棚卸し時点で見つけていなかった追加の
削除対象・修正対象がいくつかあった（下記「プランからの差分」参照）。

### 8.1 DB マイグレーション

- `sys/02_backend/migrations/009_drop_quiz.up.sql` / `.down.sql` を追加（`quizzes` 系5テーブルを DROP、
  down は `008_quiz_drop_mode` 適用後の状態— `mode` カラムなし — で再作成）
- `sys/02_backend/migrations/010_drop_course_exam.up.sql` / `.down.sql` を追加（`courses`/`lessons`/`exams`系
  /`user_lesson_progress`/`user_exam_results` を DROP、down は `005_exam_lesson_unique` 適用後の状態
  — `exams.lesson_id` UNIQUE 制約あり — で再作成）
- `migrations/seed.sql` から `courses`/`lessons` への INSERT を削除
- ローカル DB（docker-compose の `postgres` サービス、`localhost:5432`）で検証:
  - 適用前バージョンは `8`（クイズ機能適用済み）
  - `migrate up` → version `10`、成功
  - このプロジェクトの `cmd/migrate` は `up`/`down` のみでステップ数指定ができない仕様のため、
    `down` は全マイグレーションを巻き戻す形で実行（`user`/`session` 等も含め全テーブルが消える）。
    実行後ただちに `up` で再適用し、`seed.sql` を再投入してローカル DB を復元。
    → 結果として `009`/`010` の `up.sql`・`down.sql` は両方とも実行され、エラーなく成功したことを確認済み。
  - ⚠️ 本番 DB へは未適用（プラン記載の未決定事項1の通り、本ステップの範囲外）。

### 8.2 バックエンド

以下を削除:
- `internal/handler/{course,exam,progress,quiz,admin_quiz}.go` と対応する `*_test.go`
- `internal/service/{exam,quiz,quiz_scoring}.go` と対応する `*_test.go`
- `internal/repository/{course,exam,progress,quiz}.go` と対応する `*_test.go`

`internal/server/server.go` から該当ハンドラの生成・route登録を全て削除（`courseRepo`/`examRepo`/
`progressRepo`/`quizRepo`、`courseHandler`/`examHandler`/`progressHandler`/`adminQuizHandler`/
`quizHandler`、および `/courses*`・`/quizzes*`・`/admin/courses*`・`/admin/lessons*`・`/admin/quizzes*`・
`/lessons/*`・`/me/progress`・`/me/quizzes/*` 系の全ルート）。`optionalUser` ミドルウェアはクイズ回答
route専用だったため呼び出し自体を削除（未使用変数エラーになるのを確認しながら整理）。

`go build ./...` / `go vet ./...` / `go test ./...` 全て成功。

### 8.3〜8.6 フロントエンド

- 削除: `app/learn/line-marketing/**`, `app/learn/quiz/**`, `app/admin/courses/**`,
  `app/admin/lessons/**`, `app/admin/quizzes/**`, `components/learn/CourseLessonList.tsx`,
  `components/learn/LessonInteractive.tsx`(+test), `components/quiz/**`,
  `components/mypage/QuizHistoryPanel.tsx`, `components/mypage/QuizProgressSection.tsx`(+test)
- `Header.tsx` の `learnLinks`、`app/learn/page.tsx` の `categories`、`app/admin/layout.tsx` の
  `navLinks` を更新（`line-marketing`/`quiz` 削除、`line-yahoo-certification` 追加）
- 新規: `app/learn/line-yahoo-certification/page.tsx`（プレースホルダーページ）
- `app/learn/me/page.tsx` を簡略化: コース進捗・クイズ進捗の取得/表示ロジックを削除し、ログイン有無だけで
  出し分ける空状態シェルに縮小（未決定事項3への対応: ページ自体は残す方針を採用）

`npm run build`（型チェック含む）成功。ルート一覧に `/learn/line-marketing`・`/learn/quiz`・
`/admin/courses`・`/admin/lessons`・`/admin/quizzes` が存在せず、`/learn/line-yahoo-certification` が
静的ページとして生成されることを確認。

### 8.7 テスト更新

- `sys/04_e2e/tests/quiz.spec.ts` を削除
- `auth-and-progress.spec.ts` から「Reader progress」describe ブロック（レッスン完了・マイページ進捗）を削除
- `public-content.spec.ts` からコース/レッスン関連の2テストを削除、トップページのテストを新ナビ
  （運用・設定／AI活用事例／LINEヤフー認定資格勉強）に合わせて更新
- `helpers/db.ts` から未使用になった `seedQuiz`/`TestQuiz` を削除
- `app/learn/me/page.test.tsx` を簡略化後のページに合わせて書き換え（ログイン有無の2ケースのみ）
- Vitest: 3ファイル・13件全て成功
- Playwright: 13件全て成功（docker イメージを再ビルドし、実際に line-web/line-api コンテナを
  差し替えた上で実行）

### 8.8 ドキュメント

- `README.md` の Content Site Map テーブルから `line-marketing` 行を削除し
  `line-yahoo-certification` 行を追加。Nav order の説明文を更新
- `doc/dev-plans/02/dev-plan-2-0-overview.md`: Step 2-7 を「キャンセル」、Step 2-9 を「完了」で追加、
  「次の Step」を更新
- `public/sitemap.xml`（静的ファイル）・`doc/seo-update-2026-02-16.md` に `line-marketing`/`quiz` への
  参照なし（確認のみ、変更不要）

### 8.9 動作確認

- `go build`/`go vet`/`go test` (`sys/02_backend`): 成功
- `npm run build`/`npm test` (`sys/03_frontend/web`): 成功
- `npx playwright test` (`sys/04_e2e`, docker-compose の実スタックに対して): 13件成功
- 実ブラウザ: `/learn` トップページに新カードが表示されること、
  `/learn/line-yahoo-certification` がプレースホルダー表示されることを確認（スクリーンショット確認）
- `curl` で `/learn/quiz`・`/learn/line-marketing` が404になることを確認

---

## プランからの差分（棚卸し漏れ・追加対応）

プラン本文の棚卸しで見落としていた箇所。実装中に `grep`/ビルドエラーで発見し、都度対応した:

1. **`internal/testutil/db.go`**: Go の統合テストが testcontainers 上で実マイグレーションを流すため、
   ハードコードされたテーブル truncate リスト（`dbTables`）に `quiz_*`/`exam_*`/`course`/`lesson` 系が
   残っていて全リポジトリテストが失敗した。リストを articles/tags/usecases/sessions/users のみに更新。
2. **`app/admin/page.tsx`**: ダッシュボードの `sections` 配列に `/admin/courses`（「講座・レッスン」）への
   カードリンクがあり、プランの棚卸しに含まれていなかった。削除。
3. **`components/mypage/AttemptDetailModal.tsx` / `ExamAttemptsPanel.tsx`**: `QuizProgressSection`/
   `QuizHistoryPanel` の削除で孤立したクイズ専用コンポーネント（棚卸しに含まれていなかった）。削除。
4. **`lib/types.ts`**: `Course`/`Lesson`/`Exam*`/`CourseProgress*`/`MyProgress*`/`AdminQuiz*`/`Quiz*`/
   `QuizAttempt*`/`MyQuizzesProgress` 型が丸ごと不要になった。`Tag`/`Article`/`Usecase` のみ残す形に縮小。
5. **`app/sitemap.ts`**: 動的サイトマップが `/api/courses/line-marketing` を取得してレッスンURLを
   列挙していた。取得コードを削除し、静的エントリを `line-yahoo-certification` に置き換え。
6. **`.next` ビルドキャッシュ**: 削除済みルートの型定義がキャッシュに残り `npm run build` が失敗した。
   `.next` を削除してクリーンビルドすることで解消（リポジトリ状態には影響しない一時ファイル）。
7. **`lib/api.test.ts`**: 汎用APIクライアントのテストが例示パスとして `/api/courses` 等を使っていた。
   機能的な影響はないが `/api/articles` 系に差し替え。
8. **docker イメージの再ビルド**: ローカルで起動中だった `line-web`/`line-api` コンテナが変更前の
   ビルド済みイメージだったため、E2E 実行前に `docker compose build line-web line-api` と再起動が必要だった。

## 未決定事項への対応

1. **本番 DB への影響**: 本ステップでは本番へのマイグレーション適用は行っていない（ローカル DB のみ）。
   本番適用は別途ユーザー判断・別ステップとする。
2. プレースホルダー文言: 「LINEヤフー認定資格の取得に向けた学習コンテンツを準備中です。公開まで
   今しばらくお待ちください。」という汎用的な文言で実装（要調整であれば別途修正可）。
3. マイページ: 初回実装ではページ自体を維持し空状態メッセージのみに縮小したが、是正作業で
   `QuizProgressSection`（クイズ・検定＝LINEヤフー認定資格勉強の進捗）を復元。コース進捗（マーケティング
   講座側）は引き続き非表示。
4. 管理画面ダッシュボード: `/admin/page.tsx` に courses への参照があったため削除済み（上記差分3参照）。
   quizzes への参照はダッシュボードに元々なかったため追加不要。

---

## 是正作業（クイズ・検定機能の復元）

ユーザー指摘「No, quiz/exam disapeared. I want quiz/exam feature kept as LINEヤフー認定資格勉強」を受けて、
上記8.1〜8.9で削除したクイズ・検定（Phase 2）関連ファイルを git から復元し、公開URL/ラベルのみ変更した。
マーケティング講座（Phase 1）の削除はそのまま維持。

### DB

- `009_drop_quiz.up.sql`/`.down.sql` を削除（このマイグレーション自体を撤回）
- ローカル DB に対し `009_drop_quiz.down.sql`（削除前に用意していた内容）を直接実行し、
  `quizzes`/`quiz_questions`/`quiz_choices`/`user_quiz_attempts`/`user_quiz_answers` を手動で再作成
- `migrate version` は `10`（`010_drop_course_exam` 適用済み）のまま変化なし。`009` 番の欠番を挟んで
  `010` が存在する形になるが、golang-migrate はバージョン番号の欠番を許容するため、真っさらな DB に対して
  `migrate up` を実行しても 001〜008 → 010（`009`はファイルごと存在しない）の順で適用され、
  「クイズテーブルは残る・コース/レッスン/試験テーブルは無い」という最終形と一致することを確認済み
- `internal/testutil/db.go` の `dbTables`（Goテストの truncate 対象）に `quiz_*` 系を復元

### バックエンド

- git checkout で復元: `handler/{quiz,admin_quiz}.go`+test, `service/{quiz,quiz_scoring}.go`+test,
  `repository/quiz.go`+test
- `internal/server/server.go`: `quizRepo`/`adminQuizHandler`/`quizHandler`/`optionalUser` と
  `/quizzes*`・`/admin/quizzes*`・`/admin/quiz-questions*`・`/quiz-questions/*`・`/me/quizzes/*` 系
  ルートを復元（course/exam/progress 側は復元せず削除のまま）
- `handler/admin_quiz_test.go` が `course_test.go`（削除したまま）で定義されていた package-level ヘルパー
  `mustJSON` に依存していたため、`admin_quiz_test.go` 内に `mustJSON` を移設して解決
- `go build`/`go vet`/`go test ./...`: 全て成功

### フロントエンド

- git checkout で復元: `app/admin/quizzes/**`, `app/learn/quiz/**`（後述の通り移設）,
  `components/quiz/**`, `components/mypage/{QuizHistoryPanel,QuizProgressSection,
  AttemptDetailModal,ExamAttemptsPanel}.tsx`(+test)
- `lib/types.ts`: `AdminQuiz*`/`Quiz*`/`QuizAttempt*`/`MyQuizzesProgress` 型を復元（`Course`/`Lesson`/
  `Exam*`/`CourseProgress*`/`MyProgress*` は復元せず削除のまま）
- ルート移設: `app/learn/quiz/page.tsx` と `app/learn/quiz/[slug]/page.tsx` を
  `app/learn/line-yahoo-certification/` 配下へ移動（初回実装で作成したプレースホルダーは置き換え）。
  ページ内のタイトル・見出し・戻るリンクを「クイズ・検定」→「LINEヤフー　認定資格勉強」・
  `/learn/quiz` → `/learn/line-yahoo-certification` に変更
- `components/mypage/QuizProgressSection.tsx`: 見出しとリンク先を「LINEヤフー　認定資格勉強」/
  `/learn/line-yahoo-certification` に変更
- `app/admin/layout.tsx`: 管理画面ナビに「LINEヤフー　認定資格勉強」（`/admin/quizzes`）を復元
- `app/admin/quizzes/page.tsx`: 見出しを「LINEヤフー　認定資格勉強」に変更
- `app/learn/me/page.tsx`: `QuizProgressSection` を復元表示（コース進捗は非表示のまま）
- `Header.tsx`／`app/learn/page.tsx`: コメントを実態に合わせて更新（「削除した」ではなく
  「同じ機能を再ブランディングした」旨に修正）。ラベル・リンク先は初回実装時点で既に
  `/learn/line-yahoo-certification` / 「LINEヤフー　認定資格勉強」になっていたため変更不要
- `app/learn/me/page.test.tsx`: クイズセクション表示を検証する内容に書き換え。実装中に判明した
  既知の問題として、全角スペース（U+3000）を含む文字列を `screen.getByText('...')` に直接渡すと
  Testing Library のデフォルト正規化（`\s+` で U+3000 も半角スペースに潰す）によって一致しなくなるため、
  `/LINEヤフー\s*認定資格勉強/` のような正規表現マッチャーに変更して回避

`npm run build` 成功（`/learn/line-yahoo-certification`・`/learn/line-yahoo-certification/[slug]`・
`/admin/quizzes`・`/admin/quizzes/[id]/edit`・`/admin/quizzes/new` がルート一覧に復元されていることを確認）。
Vitest: 6ファイル・28件全て成功。

### E2E・実ブラウザ確認

- `sys/04_e2e/tests/quiz.spec.ts` を git checkout で復元し、`/learn/quiz/` → `/learn/line-yahoo-certification/`
  へのURL置換と、マイページ見出しのアサーションを「クイズ・検定」→「LINEヤフー　認定資格勉強」に変更
- `helpers/db.ts` を git checkout で復元（`seedQuiz`/`TestQuiz` が戻る）
- docker イメージを再ビルド・再起動（`docker compose build line-web line-api` → `up -d`）
- `npx playwright test`: 17件全て成功（quiz.spec.ts の4件を含む — 未ログイン解答、ログイン検定モード提出、
  マイページ解答・受験履歴、管理画面でのクイズ作成→公開→公開ページで解答、が実際のスタックに対して動作確認済み）
- 実ブラウザで `/learn/line-yahoo-certification` を確認: タイトル・説明文が正しく表示され、
  `curl`で `/api/quizzes` が `{"quizzes":[]}` を返す（=バックエンドAPIも生きている）ことを確認
