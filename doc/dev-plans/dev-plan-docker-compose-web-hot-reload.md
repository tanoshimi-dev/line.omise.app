# Docker Compose — Web Hot Reload

**対象:** ローカル開発用 Docker Compose の `line-web` サービス
**依存:** なし

---

## 背景・原因

既存の `sys/01_infra/docker-compose.yml` は、`line-web` を production 用の
standalone イメージとしてビルドし、`node server.js` で起動する。ホストの
ソースコードをコンテナにマウントしておらず、`next dev` も起動しないため、
`docker compose up` 後の編集はコンテナに届かず Fast Refresh も起きない。

Windows ホストから Linux コンテナへの bind mount では、ファイル変更イベントが
通知されない環境がある。そのため、開発用サービスでは polling を明示的に有効化する。

## 方針

Docker Compose の標準的な `docker-compose.override.yml` を追加する。これにより
`sys/01_infra` での通常の `docker compose up` は開発設定を自動的にマージする一方、
`docker-compose.prod.yml` は変更しない。

### 1. 開発用の web Docker stage を追加

- `sys/03_frontend/web/Dockerfile` に `dev` stage を追加する。
- `npm ci` 済みの開発依存関係を含む `node_modules` を持たせる。
- production 用 `build` / `run` stage は変更しない。

### 2. Compose override を追加

- `line-web` の build target を `dev` にする。
- Web ソースを `/app` に bind mount する。
- `/app/node_modules` と `/app/.next` を named volume にして、Linux 用依存関係と
  Next の開発キャッシュをホスト側のファイルで上書きしないようにする。
- `next dev --hostname 0.0.0.0 --webpack` を起動する。Webpack mode と
  `WATCHPACK_POLLING=true` により、bind mount 越しでも変更を検知する。

## 成果物

- `sys/03_frontend/web/Dockerfile`
- `sys/01_infra/docker-compose.override.yml`

## 完了条件

- `cd sys/01_infra; docker compose up --build` で `line-web` が Next.js development
  server として起動する。
- ホスト上の `sys/03_frontend/web/src/` の変更がコンテナで検知され、ブラウザで
  Fast Refresh または再コンパイルが行われる。
- `docker compose -f docker-compose.prod.yml up --build` は production 用 stage を
  引き続き使用する。
