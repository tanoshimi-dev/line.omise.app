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

| # | ファイル | 内容 |
|---|---|---|
| 01 | [dev-plan-01-infra-docker.md](dev-plan-01-infra-docker.md) | Docker Compose・共有 Traefik への相乗り・Cloudflare DNS |
| 02 | [dev-plan-02-database.md](dev-plan-02-database.md) | DB 選定・スキーマ（ユーザー・コンテンツ・試験・進捗）・マイグレーション |
| 03 | [dev-plan-03-backend-base.md](dev-plan-03-backend-base.md) | Go プロジェクト初期化・Gin セットアップ |
| 04 | [dev-plan-04-auth.md](dev-plan-04-auth.md) | LINE Login / Google OAuth 統合・Admin/Reader ロール |
| 05 | [dev-plan-05-content-api.md](dev-plan-05-content-api.md) | コンテンツ CMS API（講座・レッスン・記事・事例、Admin CRUD + 公開 GET） |
| 06 | [dev-plan-06-exam-progress-api.md](dev-plan-06-exam-progress-api.md) | 試験（exam）API・学習進捗（progress）保存 API |
| 07 | [dev-plan-07-frontend-base.md](dev-plan-07-frontend-base.md) | Next.js プロジェクト初期化・レイアウト・認証クライアント統合 |
| 08 | [dev-plan-08-frontend-lp.md](dev-plan-08-frontend-lp.md) | 既存 LP（React+Vite）の Next.js への移行 |
| 09 | [dev-plan-09-frontend-learn.md](dev-plan-09-frontend-learn.md) | `/learn/` 講座・記事一覧/詳細・試験・進捗表示 UI |
| 10 | [dev-plan-10-frontend-usecase.md](dev-plan-10-frontend-usecase.md) | `/usecase/` 一覧・詳細 UI |
| 11 | [dev-plan-11-frontend-admin.md](dev-plan-11-frontend-admin.md) | 管理画面（コンテンツ作成・更新、Admin 限定） |
| 12 | [dev-plan-12-test-phase1.md](dev-plan-12-test-phase1.md) | Go テスト・Next.js テスト（Vitest）・Playwright E2E |
| 13 | [dev-plan-13-deploy-phase1.md](dev-plan-13-deploy-phase1.md) | 本番 VPS デプロイ・Traefik/Cloudflare 登録・稼働確認 |

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

- DB エンジン（Step 02 で決定 — 候補: PostgreSQL、hannari.dev/log と揃える）
- セッション方式（Cookie セッション vs JWT、Step 04 で決定）
- Admin 権限の付与方法（許可リスト方式 or DB フラグ手動設定、Step 04 で決定）
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
