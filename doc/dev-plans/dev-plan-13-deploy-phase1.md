# Step 13 — 本番 VPS デプロイ

**フェーズ:** Deploy 1
**依存:** Step 12（テスト完了）

---

## ゴール

Phase 1 の成果物を、`hannari.dev/cloudflare/log` と同じ VPS 上の共有 Traefik / Cloudflare
基盤に相乗りする形で本番公開する。

---

## タスク

### 13.1 本番用 Docker Compose 調整

- [ ] `sys/01_infra/docker-compose.prod.yml` の最終確認（Step 01 で作成したラベル構成）
  - リソース制限、ログドライバー設定、`restart: unless-stopped`
- [ ] 本番用環境変数ファイル配置（`sys/02_backend/.env`, `sys/03_frontend/web/.env`）
- [ ] LINE Login / Google OAuth の本番コールバック URL・許可ドメインを設定

### 13.2 Traefik / Cloudflare 登録

- [ ] `hannari.dev/cloudflare/log` README「Adding a New Service」の手順に従い、
  `line-web` / `line-api` を共有 `traefik-network` に接続
- [ ] Cloudflare DNS: `line.omise.app`（既存の可能性、要確認）・`api-line.omise.app`（新規）を
  Proxied: ON で登録
- [ ] SSL/TLS モード: Full (Strict)

### 13.3 DB 本番設定

- [ ] 本番パスワード設定
- [ ] マイグレーション実行
- [ ] バックアップ設定（`pg_dump` を cron で日次実行 等）

### 13.4 デプロイ手順書

- [ ] `sys/01_infra/deploy.sh` を作成（`git pull` → build → migrate → `up -d` → ヘルスチェック）
- [ ] ロールバック手順を文書化

### 13.5 動作確認チェックリスト

- [ ] `https://line.omise.app` で LP が表示される
- [ ] `https://api-line.omise.app/health` が `200 OK` を返す
- [ ] LINE / Google ログインが本番ドメインで動作する
- [ ] `/learn/`, `/usecase/` が正しく表示される
- [ ] Admin が本番で講座・記事・事例を作成できる
- [ ] Reader が試験受験・進捗保存できる
- [ ] SSL 証明書の有効性確認
- [ ] `doc/seo-update-2026-02-16.md` の検証チェックリストを本番でも再実施

### 13.6 監視・ログ

- [ ] `docker compose logs -f line-api line-web` での確認方法を文書化
- [ ] ディスク容量・SSL 証明書有効期限の監視方針

---

## 成果物

- `sys/01_infra/docker-compose.prod.yml`（確定版）
- `sys/01_infra/deploy.sh`
- 本番環境変数ファイル群（Git 管理外）
- デプロイ手順書

## 完了条件

- `line.omise.app` / `api-line.omise.app` が本番で稼働している
- LINE Login / Google OAuth が本番で動作する
- Admin によるコンテンツ管理・Reader による学習/進捗保存が本番で確認できる
- バックアップが日次で自動実行されている
