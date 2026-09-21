# Web — Sticky Footer Layout

**対象:** `sys/03_frontend/web/src/app/layout.tsx` の共通ページレイアウト
**依存:** なし

---

## 背景・原因

短い `/learn` ページでは、`body` に `min-h-screen` があるだけで、その子要素が縦方向の
flex layout になっていない。したがって `main` はコンテンツ高だけで終わり、Footer は残りの
viewport 高を埋めず、その直後に余白が見える。

## 方針

- `body` を viewport 最小高の縦 flex container にする。
- `main` に `flex-1` を付け、Header と Footer 以外の残りの高さいっぱいに伸ばす。
- Footer のコンポーネントと各ページ個別の余白・コンテンツ量には変更を加えない。

## 成果物

- `sys/03_frontend/web/src/app/layout.tsx`

## 完了条件

- `/learn` のような短いページで Footer 下端が viewport 下端と一致する。
- コンテンツが viewport より高いページでは、通常どおり Footer がコンテンツの後に続く。
- Header の固定表示用の `main` 上部余白は維持される。
