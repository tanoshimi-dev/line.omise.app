# Step — line-api ヘルスチェック追加（起動順序レースの解消）

**対象:** インフラ（Docker Compose）— 特定フェーズに属さない単発の不具合修正
**依存:** なし（[dev-plan-01-infra-docker.md](01/dev-plan-01-infra-docker.md) で構築した既存の compose 構成に対する修正）

---

## 背景・再現した事象

`docker compose up` 直後、`line-web` が以下のエラーを出力する（後続リクエストで自然に解消する場合が多いが、
起動直後の SSR/初回リクエストで発生しうる）:

```
line-web-1  | ⨯ TypeError: fetch failed
line-web-1  |   [cause]: Error: getaddrinfo EAI_AGAIN line-api
```

**原因:** `docker-compose.yml` / `docker-compose.prod.yml` の `line-web` の `depends_on: [line-api]` に
`condition` が指定されておらず、デフォルトの `service_started`（＝ `line-api` コンテナの**プロセスが起動しただけ**）
で `line-web` が起動してしまう。`line-api` はさらに `postgres` の `service_healthy` を待って起動するため、
`line-web` 起動時点では `line-api` のネットワーク上の名前解決やリッスンがまだ安定していないタイミングがあり、
Next.js の SSR が起動直後に `API_URL=http://line-api:8080`（`sys/03_frontend/web/.env:11`）へ即座に fetch する際に
DNS 解決が一時的に失敗する。

`postgres` に対してはすでに `healthcheck` + `condition: service_healthy` の組み合わせで同様のレースを回避済み
（`docker-compose.yml:29-31, 46-50`）。同じパターンを `line-api` にも適用する。

`line-api` には既に `GET /health` エンドポイントが実装済み（`sys/02_backend/internal/server` 内、DB 疎通確認込み。
[dev-plan-03-backend-base-result.md](01/result/dev-plan-03-backend-base-result.md) 参照）。
これを Compose の `healthcheck` から利用する。

`line-api` のランタイムイメージは `alpine:3.20`（`sys/02_backend/Dockerfile`）で、`curl` は入っていないが
busybox 版の `wget` が標準搭載されている。実際に稼働中のコンテナで確認済み:

```
$ docker exec line-api-1 wget --no-verbose --tries=1 --spider http://localhost:8080/health
Connecting to localhost:8080 ([::1]:8080)
remote file exists   # exit code 0
```

→ イメージ変更（`curl` 追加など）は不要で、`wget --spider` で healthcheck が組める。

---

## タスク

### 1. `line-api` に healthcheck を追加

- [ ] `sys/01_infra/docker-compose.yml` の `line-api` サービスに追加:
  ```yaml
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
    interval: 5s
    timeout: 3s
    retries: 5
  ```
- [ ] `sys/01_infra/docker-compose.prod.yml` の `line-api` サービスにも同内容を追加（開発・本番で挙動を揃える）

### 2. `line-web` の `depends_on` を healthcheck 待ちに変更

- [ ] `sys/01_infra/docker-compose.yml` の `line-web`:
  ```yaml
  depends_on:
    line-api:
      condition: service_healthy
  ```
- [ ] `sys/01_infra/docker-compose.prod.yml` の `line-web` にも同様に適用

---

## 成果物

- `sys/01_infra/docker-compose.yml`（`line-api` に `healthcheck`、`line-web` の `depends_on` 修正）
- `sys/01_infra/docker-compose.prod.yml`（同上）

## 完了条件

- `docker compose up` 実行後、`docker compose ps` で `line-api` が `healthy` になってから `line-web` が起動する
- `line-web` 起動直後のログに `EAI_AGAIN line-api` が出力されない（複数回の再起動で再現しないことを確認）
- 既存の `postgres` の healthcheck パターンと構成が揃っている
