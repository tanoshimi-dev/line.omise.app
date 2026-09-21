# Docker Compose — Web Hot Reload: Result

## 実施内容

- `sys/03_frontend/web/Dockerfile` に、開発依存関係を含む `dev` stage を追加した。
- `sys/01_infra/docker-compose.override.yml` を追加した。通常の
  `docker compose up` ではこの override が自動適用され、`line-web` は次の設定で起動する。
  - ホストの Web ソースを `/app` に bind mount
  - `node_modules` と `.next` は named volume として保持
  - `next dev --hostname 0.0.0.0 --webpack` を実行
  - `WATCHPACK_POLLING=true` により Windows の bind mount 越しでも変更を検知
- `docker-compose.prod.yml` と production 用 Docker stages は変更していない。

## 検証

- `docker compose config --no-interpolate` で override のマージ結果を確認した。
- `docker compose build line-web` が成功した。
- `docker compose up -d line-web` 後、以下を確認した。
  ```
  ▲ Next.js 16.3.5 (webpack)
  ✓ Ready
  ```
- `http://localhost:3000/` は HTTP 200 を返した。
- bind mount 上のホームページソースを一時的に変更してリクエストし、開発サーバーが継続して
  ページを提供することを確認した。一時的な変更は検証直後に取り除いた。

## 利用方法

```powershell
cd sys/01_infra
docker compose up --build
```

Web の `src/` を編集すると Fast Refresh（対象外の変更では再コンパイル）が動作する。
この変更は Web frontend の hot reload のみを対象とし、Go API の自動再起動は追加していない。
