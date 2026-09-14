# Step 2-5 実装結果 — フロントエンド: マイページ「受験履歴・進捗」

**対応プラン:** [dev-plan-2-5-frontend-mypage.md](../dev-plan-2-5-frontend-mypage.md)

---

## 実施内容

バックエンド（Step 2-3）は変更なし。既存の `/api/me/quizzes/*` エンドポイントをそのまま利用。

### 2-5.1 練習モードの履歴表示

- `src/components/mypage/QuizHistoryPanel.tsx` — `GET /api/me/quizzes/:slug/history` を遅延取得（カード展開時のみ）し、
  設問・解答日時・正誤を一覧表示
- 正答率は `QuizProgressSection` 側で `GET /api/me/quizzes/progress` の `answered_count`/`correct_count` から算出して表示
  （「解答済み: X / Y問（正答率 Z%）」）

### 2-5.2 検定モードの受験履歴・進捗表示

- `src/components/mypage/ExamAttemptsPanel.tsx` — `GET /api/me/quizzes/:slug/attempts` を遅延取得し、
  受験日時・スコア・合否の一覧を表示。各行クリックで詳細モーダルを開く
- `src/components/mypage/AttemptDetailModal.tsx` — `GET /api/me/quizzes/:slug/attempts/:attemptId` を取得し、
  設問ごとの正誤・正しい選択肢・解説を振り返れるモーダル（プランの「受験詳細ページ／モーダル」のうちモーダルを選択 —
  一覧ページ内で完結させ、新規ルーティングを増やさないため）
- 進捗サマリー（受験回数・最高スコア・直近スコア・直近の合否）は `QuizProgressSection` のカードに表示

### 2-5.3 全体サマリー

- `src/components/mypage/QuizProgressSection.tsx` — `GET /api/me/quizzes/progress` を取得し、
  公開中の全クイズ／検定を `mode` に応じたカード（練習/検定）で一覧表示。セクション見出しに
  「一覧を見る」（`/learn/quiz` への導線）を常時表示し、未受験のクイズ／検定もカードのタイトルリンクから
  `/learn/quiz/:slug` へ遷移して挑戦できる
- `src/app/learn/me/page.tsx`（既存マイページ）の末尾に本セクションを追加

### 2-5.4 未ログイン時の扱い

- `QuizProgressSection` は `src/app/learn/me/page.tsx` の既存ログインゲート（`!user` は `LoginPrompt` を表示して
  return する分岐）の外側には出さず、`{!authLoading && user && <QuizProgressSection />}` の形で
  ログイン済みの場合のみレンダリングする。未ログイン時はページ全体が既存の `LoginPrompt` のみになり、
  クイズ・検定の履歴・進捗は一切表示されない

---

## プラン外で追加対応したこと

なし（バックエンド API は Step 2-3 のものをそのまま利用、追加・変更なし）。

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

なし。

---

## 検証

- `npm run build`（Next.js）— 成功。既存の `/learn/me` を含む全ルートの型検証・ビルドが通過
- `npm test`（Vitest）— 25件成功（既存22件 + 新規 `QuizProgressSection.test.tsx` 3件）
  - 公開クイズが0件の場合の空メッセージ表示
  - 練習モード: カードに解答済み数・正答率を表示 → 「解答履歴を見る」クリックで
    `GET /api/me/quizzes/:slug/history` を呼び出し、設問・正誤を表示
  - 検定モード: カードに受験回数・最高/直近スコア・直近合否を表示 → 「受験履歴を見る」クリックで
    `GET /api/me/quizzes/:slug/attempts` を呼び出し、受験行クリックで
    `GET /api/me/quizzes/:slug/attempts/:attemptId` を呼び出して詳細モーダルを表示、
    「閉じる」でモーダルが閉じることを確認
- **実ブラウザでの動作確認は未実施**: 本セクションはマイページ全体と同様ログイン必須で、
  実際の LINE/Google ログインでセッションを取得する必要がある（Step 2-2 の管理画面と同じ制約 — その際の
  結果ドキュメント参照）。今回は API 呼び出し・展開/モーダル開閉・表示内容の分岐（練習/検定・0件/複数件）を
  モック API に対する自動テストで確認する方針とした。Step 2-4（`/learn/quiz` の解答・受験自体）は
  ログイン不要のため実ブラウザで確認済みで、同じ `/api/me/quizzes/*` バックエンドは Step 2-3 の
  実 DB 統合テスト（ログイン済みユーザーでの解答保存・履歴取得を含む）で検証済み。

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| ログインユーザーがマイページで練習モードの解答履歴・正答率を確認できる | ✅ コンポーネントテストで確認（実ブラウザでのログイン確認は未実施、上記参照） |
| ログインユーザーがマイページで検定モードの受験履歴（スコア推移）・進捗・受験詳細を確認できる | ✅ 同上 |
| 未ログイン時は履歴・進捗が表示されず、ログイン導線が表示される | ✅ 既存のページ全体のログインゲートに準拠 |

## 次のステップ

- [dev-plan-2-6-test.md](../dev-plan-2-6-test.md)（Go / Next.js / E2E テスト）
