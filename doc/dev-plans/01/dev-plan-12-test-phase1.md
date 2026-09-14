# Step 12 — Phase 1 テスト

**フェーズ:** Test 1
**依存:** Step 03〜11 全て完了

---

## ゴール

バックエンド（Go）・フロントエンド（Next.js）・E2E のテスト基盤を整え、
Phase 1 の主要フローに対するテストを実装する。

---

## タスク

### 12.1 Go テスト基盤

- [ ] `internal/testutil/` にテストヘルパーを整備（DB セットアップ、認証済みリクエストのモック等）
- [ ] Repository 層のテスト（DB を使う統合テスト — `testcontainers-go` の採用を検討）
- [ ] Service 層のテスト（採点ロジック、公開判定ロジック等）
- [ ] Handler 層のテスト（認証ミドルウェアの 401/403 挙動を含む）
- [ ] カバレッジ計測: `go test ./... -coverprofile=coverage.out`

### 12.2 Next.js テスト

- [ ] Vitest + React Testing Library の導入
- [ ] コンポーネント単体テスト（試験受験フォーム、進捗表示、記事一覧のタグ絞り込み等）
- [ ] API クライアント（`src/lib/api.ts`）のモックテスト

### 12.3 E2E テスト（Playwright）

- [ ] LINE / Google ログイン〜ログアウトのフロー（テスト用アカウント or モックプロバイダー方針を決定）
- [ ] `/learn/line-marketing/` の講座受講 → レッスン完了 → 試験受験 → 進捗反映の一連の流れ
- [ ] `/learn/line-operation/`, `/learn/ai/` の記事一覧・タグ絞り込み・詳細表示
- [ ] `/usecase/` の一覧・詳細表示
- [ ] Admin ログインでのコンテンツ作成 → 公開 → 公開ページでの表示確認

### 12.4 回帰確認

- [ ] LP（Step 08 移行分）の表示・SEO メタタグの確認
- [ ] 未ログイン状態でのアクセス制御（保護エンドポイント・管理画面）の確認

---

## 成果物

- Go / Next.js / E2E の各テストコード
- カバレッジレポート

## 完了条件

- `go test ./...` が全て成功する
- `npm test`（Vitest）が全て成功する
- `npx playwright test` の主要シナリオが成功する
