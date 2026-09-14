# Step 2-2 — クイズ／検定 管理 API（Admin CRUD）

**フェーズ:** Phase 2 — クイズ・検定機能
**依存:** Step 2-1（DB マイグレーション）、Phase 1 [dev-plan-04-auth.md](../01/dev-plan-04-auth.md)（`RequireAdmin`）

---

## ゴール

管理者が「クイズ（練習モード）」「検定（検定モード）」を、共通の問題・選択肢・解説データ構造で
作成・編集・削除できる Admin 専用 API を実装する。

---

## タスク

### 2-2.1 クイズ／検定本体

- [ ] `POST /api/admin/quizzes` — 作成（`title`, `slug`, `description`, `mode`, `passing_score`）
  - `mode = practice` の場合、`passing_score` を無視 or 指定されたら `400`
  - `mode = exam` の場合のみ `passing_score` を許可（NULL可＝合否判定なし、スコアのみ表示）
- [ ] `PUT /api/admin/quizzes/:id` — 更新
- [ ] `DELETE /api/admin/quizzes/:id` — 削除（`ON DELETE CASCADE` で設問・選択肢も削除）
- [ ] `GET /api/admin/quizzes` — 管理画面用一覧（`practice`/`exam` 両方、`published` 状態問わず）

### 2-2.2 設問・選択肢

- [ ] `POST /api/admin/quizzes/:quizId/questions` — 設問追加
  - `question_text`, `allow_multiple`, `explanation`, `reference_url`, `sort_order`
  - 選択肢一覧を同時に受け取り作成（`choice_text`, `is_correct`, `sort_order`）
- [ ] `PUT /api/admin/quiz-questions/:id` — 設問更新（選択肢の追加・更新・削除を含む）
- [ ] `DELETE /api/admin/quiz-questions/:id` — 設問削除

### 2-2.3 バリデーション

- [ ] 各設問に選択肢が2つ以上あること
- [ ] `is_correct = true` の選択肢が1つ以上存在すること
- [ ] `allow_multiple = false` の設問で `is_correct = true` の選択肢が2つ以上ある場合は `400`
- [ ] `slug` の重複チェック（`quizzes.slug` UNIQUE 制約違反時は `409`）

### 2-2.4 管理画面 UI（既存の管理画面基盤に追加）

- [ ] Phase 1 [dev-plan-11-frontend-admin.md](../01/dev-plan-11-frontend-admin.md) の管理画面に
      「クイズ／検定」管理メニューを追加
- [ ] クイズ／検定の一覧・作成・編集フォーム（`mode` 切り替え、`practice`/`exam` でフォーム項目を出し分け）
- [ ] 設問・選択肢のインライン編集 UI（正解フラグのチェックボックス、`allow_multiple` トグル）

---

## 成果物

- `internal/handler/admin_quiz.go`
- `internal/repository/quiz.go`（Admin 用 CRUD メソッド）
- 管理画面: `src/app/admin/quizzes/` 配下のページ・コンポーネント

## 完了条件

- Admin が練習モードのクイズと検定モードの試験（`passing_score` 付き）の両方を作成・編集・削除できる
- 不正なデータ（選択肢不足、正解フラグなし、`allow_multiple=false` での複数正解など）が `400` で弾かれる
- 管理画面から一連の操作が実際に行える
