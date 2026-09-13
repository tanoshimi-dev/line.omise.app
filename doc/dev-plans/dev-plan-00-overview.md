# 開発計画 — line.omise.app

`E:\dev\vs_code\products\hannari.dev\cloudflare\log\sys` (`doc/dev-plan/`) の構成を参考に、
line.omise.app を静的ランディングページから実アプリ（Next.js フロントエンド + Go Gin バックエンド、
LINE Login / Google OAuth 認証）へ移行するための開発計画。

各 Step の実装は、`CLAUDE.md` の開発フロー（プラン → 実装指示 → 実装 + 結果ドキュメント作成 → git はユーザーが実施）に従う。

---

## スコープ（重要）

- **会員管理・サロン予約・スイーツショップの3ミニアプリは対象外。** 既に別ドメインで稼働中のため、
  line.omise.app 側では実装しない（LP 上では実績・事例として紹介するのみ）。
- **line.omise.app の主要コンテンツは以下の3つ:**
  1. **LP**（トップページ — 既存 React+Vite ランディングページの Next.js への移行）
  2. **`/learn/`**（学習コンテンツ — LINEマーケティング講座・LINE運用/設定記事・生成AI活用事例）
  3. **`/usecase/`**（導入店舗インタビュー）
- **バックエンド（Go Gin）の役割 = コンテンツ CMS API**
  - 管理者（Admin）: コンテンツ（講座・レッスン・記事・事例・試験問題）の作成・更新
  - 一般ユーザー（Reader）: LINE Login / Google OAuth でログインし、コンテンツ閲覧・学習コンテンツの
    試験（exam）受験・学習進捗（progress）の保存ができる

---

## フェーズ構成

| フェーズ | 内容 | 目安 |
|---|---|---|
| Phase 1 — Foundation | インフラ・DB・バックエンド基盤・認証（Admin/Reader） | Step 01〜04 |
| Phase 1 — CMS API | コンテンツ CRUD・試験/進捗 API | Step 05〜06 |
| Phase 1 — Frontend | Next.js 基盤・LP移行・learn・usecase・管理画面 | Step 07〜11 |
| Test 1 | Phase 1 テスト（Go / Next.js / E2E） | Step 12 |
| Deploy 1 | 本番 VPS デプロイ（既存 Traefik/Cloudflare 共有基盤に相乗り） | Step 13 |

## 開発計画ファイル一覧

進捗ステータス最終更新: 2026-09-13（Step 08）

| # | ファイル | 内容 | 状態 |
|---|---|---|---|
| 01 | [dev-plan-01-infra-docker.md](dev-plan-01-infra-docker.md) | Docker Compose・共有 Traefik への相乗り・Cloudflare DNS | ✅ 完了（[結果](result/dev-plan-01-infra-docker-result.md)）— Cloudflare DNS 登録のみユーザー側対応待ち |
| 02 | [dev-plan-02-database.md](dev-plan-02-database.md) | DB 選定・スキーマ（ユーザー・コンテンツ・試験・進捗）・マイグレーション | ✅ 完了（[結果](result/dev-plan-02-database-result.md)）— PostgreSQL + golang-migrate、`docker-compose.yml` に `postgres`/`adminer` サービス追加済み |
| 03 | [dev-plan-03-backend-base.md](dev-plan-03-backend-base.md) | Go プロジェクト初期化・Gin セットアップ | ✅ 完了（[結果](result/dev-plan-03-backend-base-result.md)）— `/health` の DB 到達性チェックは Step 02 で実ドライバ（pgx）に置き換え済み |
| 04 | [dev-plan-04-auth.md](dev-plan-04-auth.md) | LINE Login / Google OAuth 統合・Admin/Reader ロール | ✅ 完了（[結果](result/dev-plan-04-auth-result.md)）— LINE / Google 両方の実ログインをユーザーが確認済み |
| 05 | [dev-plan-05-content-api.md](dev-plan-05-content-api.md) | コンテンツ CMS API（講座・レッスン・記事・事例、Admin CRUD + 公開 GET） | ✅ 完了（[結果](result/dev-plan-05-content-api-result.md)）— 実装中に `RequireAdmin` の認可バイパスの重大バグを発見・修正 |
| 06 | [dev-plan-06-exam-progress-api.md](dev-plan-06-exam-progress-api.md) | 試験（exam）API・学習進捗（progress）保存 API | ✅ 完了（[結果](result/dev-plan-06-exam-progress-api-result.md)） |
| 07 | [dev-plan-07-frontend-base.md](dev-plan-07-frontend-base.md) | Next.js プロジェクト初期化・レイアウト・認証クライアント統合 | ✅ 完了（[結果](result/dev-plan-07-frontend-base-result.md)）— ログイン導線は Step 04 の実ログイン確認で動作確認済み |
| 08 | [dev-plan-08-frontend-lp.md](dev-plan-08-frontend-lp.md) | 既存 LP（React+Vite）の Next.js への移行 | ✅ 完了（[結果](result/dev-plan-08-frontend-lp-result.md)）— `ProfileSection` のハイドレーションバグを修正、`og-image.png` は未作成のまま |
| 09 | [dev-plan-09-frontend-learn.md](dev-plan-09-frontend-learn.md) | `/learn/` 講座・記事一覧/詳細・試験・進捗表示 UI | ⬜ 未着手 |
| 10 | [dev-plan-10-frontend-usecase.md](dev-plan-10-frontend-usecase.md) | `/usecase/` 一覧・詳細 UI | ⬜ 未着手 |
| 11 | [dev-plan-11-frontend-admin.md](dev-plan-11-frontend-admin.md) | 管理画面（コンテンツ作成・更新、Admin 限定） | ⬜ 未着手 |
| 12 | [dev-plan-12-test-phase1.md](dev-plan-12-test-phase1.md) | Go テスト・Next.js テスト（Vitest）・Playwright E2E | ⬜ 未着手 |
| 13 | [dev-plan-13-deploy-phase1.md](dev-plan-13-deploy-phase1.md) | 本番 VPS デプロイ・Traefik/Cloudflare 登録・稼働確認 | ⬜ 未着手 |

**次の Step:** 09（`/learn/` UI）・10（`/usecase/` UI）— LP移行（Step08）で確立した
「ページごとに metadata を持つ」パターンと、Step05/06 の `/api/*` を消費して実装する。
残る TODO: `public/images/og-image.png` の作成（ユーザー側のデザイン作業）。

## 依存関係

```
01 インフラ ─→ 02 DB ─→ 03 バックエンド基盤 ─→ 04 認証(Admin/Reader)
                                                      │
                                ┌─────────────────────┤
                                ↓                     ↓
                        05 コンテンツAPI       06 試験/進捗API
                                │                     │
                                └──────────┬──────────┘
                                           ↓
                        07 フロントエンド基盤
                                │
                    ┌───────────┼───────────┬──────────────┐
                    ↓           ↓           ↓              ↓
              08 LP移行   09 learn UI  10 usecase UI  11 管理画面
                    │           │           │              │
                    └───────────┴───────────┴──────────────┘
                                           ↓
Test 1                          07〜11 完了 ─→ 12 Phase1テスト
                                           ↓
Deploy 1                              12 完了 ─→ 13 本番デプロイ
```

## 未決定事項（各 Step 内で決定・記録する）

- ~~DB エンジン（Step 02 で決定 — 候補: PostgreSQL、hannari.dev/log と揃える）~~ → **決定済み: PostgreSQL + `golang-migrate`**（[dev-plan-02-database.md](dev-plan-02-database.md) 2.1）
- ~~セッション方式（Cookie セッション vs JWT、Step 04 で決定）~~ → **決定済み: HMAC署名付き Cookie + DB セッション（`sessions` テーブル）**（[dev-plan-04-auth-result.md](result/dev-plan-04-auth-result.md) 4.3）
- ~~Admin 権限の付与方法（許可リスト方式 or DB フラグ手動設定、Step 04 で決定）~~ → **決定済み: `ADMIN_EMAILS` 許可リストで自動昇格（降格はしない）**（[dev-plan-04-auth-result.md](result/dev-plan-04-auth-result.md) 4.5）
- コンテンツのデータモデル詳細（講座/レッスン/記事/事例/試験、Step 05〜06 で決定）
- ホスティング詳細（VPS 上のリソース配分、Step 13 で決定）

## テスト技術スタック（想定）

| 対象 | ツール | コマンド |
|---|---|---|
| Go Unit / Integration | `go test` (+ `testcontainers-go` を検討) | `go test ./... -v` |
| Go Coverage | `go tool cover` | `go test ./... -coverprofile=coverage.out` |
| Next.js Unit | Vitest + React Testing Library | `npm test` |
| E2E | Playwright | `npx playwright test` |

詳細は Step 12 で確定する。
