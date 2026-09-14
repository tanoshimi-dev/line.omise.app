# Step 2-6 実装結果 — テスト（Go / Next.js / E2E）

**対応プラン:** [dev-plan-2-6-test.md](../dev-plan-2-6-test.md)

---

## 実施内容

プラン記載の2-6.1・2-6.2の大部分は、実装を進めながら Step 2-2〜2-5 の時点で既に用意済みだった
（各ステップの結果ドキュメント参照）。本 Step で新たに追加したのは以下:

### 2-6.1 Go テスト — 既存確認のみ（追加なし）

Step 2-2（`internal/handler/admin_quiz_test.go`）・Step 2-3（`internal/service/quiz_scoring_test.go`、
`internal/handler/quiz_test.go`、`internal/repository/quiz_test.go`）で、プラン記載の項目
（バリデーション・練習/検定モードの採点ロジック・他ユーザーの履歴/進捗/受験詳細への認可・
正解フラグ/解説の非公開）は全てカバー済みであることを確認。追加テストは不要と判断。

### 2-6.2 Next.js（Vitest）— 1件追加

- Step 2-4（`QuizQuestionCard.test.tsx`／`ExamRunner.test.tsx`）・Step 2-5
  （`QuizProgressSection.test.tsx`）で大部分をカバー済み。
- ただし「マイページの…未ログイン時の表示切り替え」は `QuizProgressSection` 自体には
  ログイン判定がなく、親ページ（`src/app/learn/me/page.tsx`）側でセクションごと
  出し分けている（dev-plan-2-5 2-5.4 の設計）ため、`QuizProgressSection.test.tsx` だけでは
  この分岐を検証できていなかった。本 Step で **`src/app/learn/me/page.test.tsx`** を新規作成し、
  未ログイン時は「クイズ・検定」セクション自体がレンダリングされない（`api.get` も呼ばれない）こと、
  ログイン時は講座進捗とクイズ・検定セクションの両方が表示されることを検証。

### 2-6.3 Playwright E2E — 新規作成

- `sys/04_e2e/tests/helpers/db.ts` に `seedQuiz`（公開済みクイズ＋設問1問＋選択肢2つを直接 INSERT する
  テストヘルパー、既存の `seedUser` と同じ方針）を追加
- `sys/04_e2e/tests/quiz.spec.ts` を新規作成（4テスト）:
  1. 未ログインで練習モードのクイズに解答 → 正誤・解説が表示される
  2. ログインして検定モードの試験に解答・提出 → 最終スコア・合否が表示される
  3. ログイン後、マイページで練習モードの解答履歴（展開）・検定モードの受験履歴（展開→受験詳細モーダル）が
     表示される
  4. Admin が管理画面からクイズ・設問・選択肢を作成し公開 → フロントエンドから解答できる
     （`admin-content.spec.ts` と同じ「Admin UI で作成 → 公開 → 公開ページで確認」パターン）

---

## プラン外で追加対応したこと

- ローカル検証のため `docker compose up -d --build line-api line-web` で Phase 2（Step 2-1〜2-6）の
  コードを含む本番相当イメージを再ビルド・再起動した（テスト実行環境の準備。ソースコード変更ではない）
- 検証中、開発用 DB に `migrations/seed.sql` が未投入（コース・記事・事例のテーブルが空）であることが判明。
  これは本 Step より前からの環境状態であり Phase 2 の変更によるものではないが、既存 Playwright テスト
  （`auth-and-progress.spec.ts`・`public-content.spec.ts`）がこれに依存しているため、
  `psql < migrations/seed.sql` で投入して確認環境を整えた（ソースコード変更ではない）

## プラン未実施

なし。

---

## 検証

### Go（`sys/02_backend`、`go test ./... -v -count=1`）

- **106件成功**（うち quiz 関連 38件、Phase 1 分 68件）。regression なし。
- `go build ./...` / `go vet ./...` / `gofmt -l .` — 成功・差分なし

### Next.js / Vitest（`sys/03_frontend/web`）

- **27件成功**（Phase 1 分15件 + Phase 2 分12件〈`QuizQuestionCard` 3、`ExamRunner` 4、
  `QuizProgressSection` 3、本 Step 追加の `learn/me/page.test.tsx` 2〉）
- `npm run build`（`next build`）— 成功、全ルートの型検証・ビルドが通過

### Playwright E2E（`sys/04_e2e`、`npx playwright test`）

- **21件成功**（Phase 1 分17件 + 本 Step追加の `quiz.spec.ts` 4件）。Phase 1 の既存17件に regression なし
- 実行環境: `docker compose up -d --build` で再ビルドした line-api/line-web + 稼働中の postgres

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 上記テストが全て成功する | ✅ Go 106件・Vitest 27件・Playwright 21件、全て成功 |
| Phase 1 の既存テスト（Go 49件・Vitest 15件・Playwright 17件）がデグレしていない | ✅ Phase 1 分のテストは全て成功（Go の総数は Phase 2 分の追加により106件に増加、Phase 1 側の個別テストに失敗なし） |

## 次のステップ

- [dev-plan-2-7-deploy-production.md](../dev-plan-2-7-deploy-production.md)（本番マイグレーション適用・本番デプロイ）
