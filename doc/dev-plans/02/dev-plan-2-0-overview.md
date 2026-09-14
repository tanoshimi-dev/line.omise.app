# 開発計画 — クイズ・検定機能（Phase 2）

Phase 1（[../01/dev-plan-00-overview.md](../01/dev-plan-00-overview.md)）で構築した認証基盤・コンテンツ CMS API の上に、
LINE公式アカウント運用の知識を確認できる**クイズ（練習モード）・検定（試験モード）機能**を追加するための開発計画。

各 Step の実装は、`CLAUDE.md` の開発フロー（プラン → 実装指示 → 実装 + 結果ドキュメント作成 → git はユーザーが実施）に従う。

---

## スコープ（重要）

- 新規コンテンツ種別「クイズ（`quiz`）」を追加する。1レコードが `mode` で下記のどちらかを表す。
  - **練習モード（`practice`）**: 1問ずつ解答 → その場で正誤・解説を表示
  - **検定モード（`exam`）**: 全設問に解答 → 一括提出 → 最終スコア（正答数・正答率、任意で合否）を表示
  - 例: 「LINE運用クイズ」（practice）、「基礎LINE検定」「上級LINE検定」（exam）
- Phase 1 の `exams` / `exam_questions` / `exam_choices`（レッスン合否判定用、`lesson_id` 必須）とは**別モデル**。既存テーブル・API には手を加えない。
- 未ログインでも解答・受験は可能。**ログインユーザーのみ**解答・受験結果が保存され、マイページで受験履歴・進捗を確認できる。
- 会員管理・サロン予約・スイーツショップの3ミニアプリは引き続き対象外。

---

## 開発計画ファイル一覧

進捗ステータス最終更新: 2026-09-15（Step 2-2）

| # | ファイル | 内容 | 状態 |
|---|---|---|---|
| 2-1 | [dev-plan-2-1-db-migration.md](dev-plan-2-1-db-migration.md) | スキーマ設計・マイグレーション（`quizzes`/`quiz_questions`/`quiz_choices`/`user_quiz_attempts`/`user_quiz_answers`） | ✅ 完了（[結果](result/dev-plan-2-1-db-migration-result.md)）— サンプルデータのシード投入は Step 2-2 以降に先送り |
| 2-2 | [dev-plan-2-2-admin-api.md](dev-plan-2-2-admin-api.md) | クイズ／検定・設問・選択肢の管理 API（Admin CRUD） | ✅ 完了（[結果](result/dev-plan-2-2-admin-api-result.md)）— 管理画面での実ブラウザ確認（実ログイン）は未実施 |
| 2-3 | [dev-plan-2-3-answer-scoring-api.md](dev-plan-2-3-answer-scoring-api.md) | 解答・採点 API（練習モード即時採点／検定モード一括採点）・受験履歴・進捗 API | ⬜ 未着手 |
| 2-4 | [dev-plan-2-4-frontend-quiz-ui.md](dev-plan-2-4-frontend-quiz-ui.md) | `/learn/quiz/` 練習モード・検定モード受験 UI | ⬜ 未着手 |
| 2-5 | [dev-plan-2-5-frontend-mypage.md](dev-plan-2-5-frontend-mypage.md) | マイページ「受験履歴・進捗」セクション | ⬜ 未着手 |
| 2-6 | [dev-plan-2-6-test.md](dev-plan-2-6-test.md) | Go / Next.js / E2E テスト | ⬜ 未着手 |
| 2-7 | [dev-plan-2-7-deploy-production.md](dev-plan-2-7-deploy-production.md) | 本番マイグレーション適用・本番デプロイ | ⬜ 未着手 |

**次の Step:** 2-3（解答・採点 API／受験履歴・進捗 API）— Step 2-2（Admin API）完了。

## 依存関係

```
2-1 DB マイグレーション
        │
        ├──────────────┐
        ↓              ↓
2-2 Admin API    2-3 解答・採点/履歴/進捗 API
        │              │
        └──────┬───────┘
               ↓
   2-4 フロントエンド（クイズ/検定UI）
               │
               ↓
   2-5 フロントエンド（マイページ）
               │
               ↓
        2-6 テスト
               │
               ↓
   2-7 本番マイグレーション・デプロイ
```

## 未決定事項（各 Step 内で決定・記録する）

- クイズ／検定を `/learn/` 配下に置くか独立セクション（例: `/quiz/`, `/exam/`）にするか → Step 2-4 で決定
- 複数選択問題の採点基準（全選択肢一致で正解とするか、部分点を認めるか）→ 現時点では「完全一致のみ正解」を仮定（Step 2-3）
- クイズ／検定と既存レッスン・講座との紐付けの要否 → Step 2-2 で決定
- 検定モードの再受験ポリシー（無制限 or 回数制限・クールダウン）→ 現時点では無制限を仮定（Step 2-3）
- 検定モードで未回答の設問がある状態での提出を許可するか → Step 2-3 で決定

## テスト技術スタック

Phase 1 と同一（[../01/dev-plan-00-overview.md](../01/dev-plan-00-overview.md) 参照）。

| 対象 | ツール | コマンド |
|---|---|---|
| Go Unit / Integration | `go test` | `go test ./... -v` |
| Next.js Unit | Vitest + React Testing Library | `npm test` |
| E2E | Playwright | `npx playwright test` |

詳細は Step 2-6 で確定する。

## 変更履歴

- 初版: 単一ファイル `dev-plan-2-1-quiz-contents.md` として作成（クイズ練習モードのみ）
- 追記: 検定モード（`mode = exam`、一括提出・最終スコア表示）を追加
- 分割: 本オーバービューと Step 2-1〜2-7 に分割し、本番マイグレーション・デプロイを独立 Step として明示化
