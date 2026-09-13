# Step 01 実装結果 — インフラ・Docker Compose 構築

**対応プラン:** [dev-plan-01-infra-docker.md](dev-plan-01-infra-docker.md)

---

## 実施内容

### 1.1 / 1.2 Docker Compose 定義

- `sys/01_infra/docker-compose.yml`（開発用）を作成
  - `line-web`（`../03_frontend/web` をビルド、`3000:3000` 公開）
  - `line-api`（`../02_backend` をビルド、`8080:8080` 公開）
  - `line-omise-internal` ネットワーク
- `sys/01_infra/docker-compose.prod.yml`（本番用）を作成
  - `line-web` / `line-api` に Traefik ラベルを付与（`Host` ルール、`entrypoints=https`、
    `tls.certresolver=cloudflare`）— `hannari.dev/cloudflare/log` の
    `sys/infra/docker-compose.prod.yml` と同じパターン
  - 外部ネットワーク `traefik-network`（共有 Traefik）+ 内部ネットワーク `line-omise-internal`
- 両ファイルとも `docker compose config` で構文検証済み（`.env` を一時的に用意して確認、確認後削除）

### 1.4 環境変数テンプレート

- `sys/02_backend/.env.example` を作成（`PORT`, `DATABASE_URL`（仮）, `CORS_ALLOWED_ORIGINS`。
  Step 04 で追加される LINE Login / Google OAuth 関連変数はコメントアウトでプレビューのみ記載）
- `sys/03_frontend/web/.env.example` を作成（`NEXT_PUBLIC_API_URL`）

---

## プラン外で追加対応したこと

- ルート `.gitignore` と `sys/03_frontend/web/.gitignore` の `.env.*` ルールが
  `.env.example` も誤って無視する状態だったため、両ファイルに `!.env.example` を追加した
  （`git check-ignore` で新規作成した2つの `.env.example` が無視されていたのを確認して修正）。

---

## プラン未実施（意図的にスキップ）

- **1.1 の DB サービス定義**: `dev-plan-02-database.md` で DB エンジンを決定してから追加する。
  現時点の `docker-compose.yml` / `docker-compose.prod.yml` に DB サービスは含まれていない
  （コメントで明記済み）。
- **1.3 Cloudflare DNS 設定**: Cloudflare 管理画面での操作が必要な手動作業のため未実施。
  - `line.omise.app` の A レコードは既存の可能性があるため要確認
  - `api-line.omise.app` の A レコード（Proxied: ON）は新規追加が必要
  - SSL/TLS モードを Full (Strict) に設定
- `hannari.dev/cloudflare/log/README.md` の「Adding a New Service」節（本プランが参照している
  オンボーディング手順）は、以前の編集内容が現在ファイルから失われている状態を確認済み
  （前回のやり取りで報告済み）。Step 13（本番デプロイ）までに復元 or 再作成が必要。

---

## 既知の制約

- `docker-compose.yml` の `line-web` / `line-api` はそれぞれ `Dockerfile` を参照するが、
  実体は未作成（`line-api` の Dockerfile は Step 03、`line-web` の Dockerfile は Step 07 で追加）。
  そのため現時点で `docker compose up -d` を実行しても **ビルドは失敗する**
  （`docker compose config` によるYAML構文検証は成功済み）。

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| 開発環境で `docker compose up -d` により `line-web` / `line-api` が起動する | △ 未達— Dockerfile が Step 03/07 で追加されるまで起動不可（YAML自体は検証済み） |
| 本番想定のラベル構成が共有基盤の手順と整合する | ✅ Traefik ラベルパターンを踏襲 |
| Cloudflare DNS レコードが確認できる | ❌ 未実施（手動作業、ユーザー側で対応） |

## 次のステップ

- `dev-plan-02-database.md`（DB エンジン決定・スキーマ）
- Cloudflare DNS の登録（ユーザー側での対応が必要)
