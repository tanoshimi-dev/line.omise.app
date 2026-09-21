# Web — Sticky Footer Layout: Result

## 実施内容

`sys/03_frontend/web/src/app/layout.tsx` の共通レイアウトを次のように更新した。

- `body` を `flex min-h-screen flex-col` にして、画面の高さを最小高とする縦 flex container にした。
- `main` に `flex-1` を追加して、Header と Footer の間の余った高さを埋めるようにした。

このため、コンテンツが短い `/learn` でも Footer 下端が viewport 下端に配置される。
コンテンツが長いページでは、Footer は従来どおりコンテンツの後ろへ続く。

## 検証

- `npm run build` が成功した（Next.js compilation、TypeScript、全18 route の static generation）。
- `git diff --check` が成功した。
- 実行環境に利用可能な browser surface がなかったため、修正後のブラウザ screenshot は取得できなかった。
