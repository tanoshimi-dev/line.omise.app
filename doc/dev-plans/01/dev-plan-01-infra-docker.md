# Step 01 — インフラ・Docker Compose 構築

**フェーズ:** Phase 1 — Foundation
**依存:** なし（最初に着手）

---

## ゴール

開発・本番環境の Docker Compose 構成を整備し、`hannari.dev/cloudflare/log` と同様に
**共有 Traefik（VPS 上で稼働中、外部ネットワーク `traefik-network`）に相乗りする形**で
`line.omise.app` / `api-line.omise.app` をルーティングできるようにする。
Traefik 自体はこのリポジトリでは起動しない。

---

## タスク

### 1.1 Docker Compose 定義（開発用）

- [ ] `sys/01_infra/docker-compose.yml` を作成
- [ ] 以下のサービスを定義:

| サービス | イメージ | 内部ポート | 役割 |
|---|---|---|---|
| line-web | custom (Node) | 3000 | Next.js |
| line-api | custom (Go) | 8080 | Gin API |
| （DB） | Step 02 で決定 | - | RDB |

- [ ] ボリューム定義（DB データ）
- [ ] ネットワーク定義（internal 用の内部ネットワーク）

### 1.2 Docker Compose 定義（本番用）

- [ ] `sys/01_infra/docker-compose.prod.yml` を作成
- [ ] `line-web` / `line-api` に Traefik ラベルを付与（`hannari.dev/cloudflare/log` の
  `sys/infra/docker-compose.prod.yml` と同じパターン）:

```yaml
services:
  line-web:
    networks:
      - line-omise-web
      - line-omise-internal
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.line-web.rule=Host(`line.omise.app`)"
      - "traefik.http.routers.line-web.entrypoints=https"
      - "traefik.http.routers.line-web.tls=true"
      - "traefik.http.routers.line-web.tls.certresolver=cloudflare"
      - "traefik.http.services.line-web.loadbalancer.server.port=3000"

  line-api:
    networks:
      - line-omise-web
      - line-omise-internal
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.line-api.rule=Host(`api-line.omise.app`)"
      - "traefik.http.routers.line-api.entrypoints=https"
      - "traefik.http.routers.line-api.tls=true"
      - "traefik.http.routers.line-api.tls.certresolver=cloudflare"
      - "traefik.http.services.line-api.loadbalancer.server.port=8080"

networks:
  line-omise-web:
    external: true
    name: traefik-network
  line-omise-internal:
    name: line-omise-internal
```

- [ ] `restart: unless-stopped` を全サービスに設定
- [ ] ポート公開は行わない（Traefik が共有ネットワーク経由でアクセス）

### 1.3 Cloudflare DNS 設定

- [ ] `line.omise.app` → VPS IP（A レコード, Proxied: ON）※既存の可能性あり、要確認
- [ ] `api-line.omise.app` → VPS IP（A レコード, Proxied: ON）新規追加
- [ ] SSL/TLS モード: Full (Strict)

### 1.4 環境変数テンプレート

- [ ] `sys/02_backend/.env.example` を作成
- [ ] `sys/03_frontend/web/.env.example` を作成
- [ ] `.gitignore` に `.env` を追加（既存の `sys/secrets/*` 除外と重複しないか確認）

---

## 成果物

- `sys/01_infra/docker-compose.yml`
- `sys/01_infra/docker-compose.prod.yml`
- `sys/02_backend/.env.example`
- `sys/03_frontend/web/.env.example`

## 完了条件

- 開発環境で `docker compose up -d` により `line-web` / `line-api` が起動する
- 本番想定のラベル構成で `hannari.dev/cloudflare/log` 側の README（「Adding a New Service」節）の手順と整合する
- Cloudflare DNS レコードが確認できる
