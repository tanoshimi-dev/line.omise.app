# Step — 学習コンテンツメニューの並び替え・中央寄せ、管理画面ダッシュボードの並び替え

**対象:** `/learn` トップページのカードメニュー、ヘッダーのモバイルメニュー、管理画面ダッシュボード — 表示順・レイアウトの調整
**依存:** [dev-plan-2-9-line-yahoo-certification.md](02/dev-plan-2-9-line-yahoo-certification.md)、[dev-plan-menu-label-shorten.md](dev-plan-menu-label-shorten.md)

---

## ユーザー指示

- learn メニューの並び順を「LINEヤフー認定資格、LINE運用・設定、生成AI活用事例」に変更
- learn メニューのレイアウトを中央寄せに変更（現状は左寄せ）
- 管理画面ダッシュボードのメニューも「LINEヤフー認定資格、記事、導入事例」の順に変更

## 対象箇所の特定

指示文中の「LINE運用・設定」「生成AI活用事例」という表記は、ヘッダーの短縮ラベル（「運用・設定」
「AI活用事例」）ではなく `sys/03_frontend/web/src/app/learn/page.tsx` の `categories` 配列のタイトルと
一致する。よって **「learn メニュー」= `/learn` トップページのカードグリッド**と判断する。

- **並び順:** `categories`（`learn/page.tsx`）と、同じ順序で運用されている
  `learnLinks`（`Header.tsx` のモバイルメニュー「学習コンテンツ」サブリスト）の両方を変更する
  （2つの配列は「同じ並び順にする」と既存コードコメントで明記されており、これまでも常に揃えてきたため）。
- **中央寄せ:** 現状 `/learn` のグリッドは `grid-cols-1 sm:grid-cols-2 lg:grid-cols-4`。項目が3つしか
  ないため、`lg`（4カラム）幅では右側に空きカラムができ、カード群が左寄りに見える。
  **`lg:grid-cols-4` → `lg:grid-cols-3` に変更**し、3項目がちょうど3カラムを均等に埋める形にする
  （カード自体は現状通りグリッドセルいっぱいに広がる `rounded-2xl` カードのまま、外枠 `max-w-5xl mx-auto`
  は変更しない）。
  - ヘッダーのモバイルメニュー側（`学習コンテンツ` サブリスト、`block pl-2` の縦並びリンク）は
    「左寄せ／中央寄せ」という指摘の対象外と判断（縦リストのテキスト整列は今回変更しない）。
    もし実際にはこちらも中央寄せにしたい場合は別途指摘してください。

## 変更内容

### 1. `/learn` トップページの並び替え・グリッド変更

`sys/03_frontend/web/src/app/learn/page.tsx`

- [ ] `categories` 配列の並びを次の順に変更:
  1. `/learn/line-yahoo-certification`（LINEヤフー認定資格）
  2. `/learn/line-operation`（LINE運用・設定）
  3. `/learn/ai`（生成AI活用事例）
- [ ] グリッドのクラスを `grid-cols-1 sm:grid-cols-2 lg:grid-cols-4` → `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3` に変更
- [ ] 配列直前のコメント（並び順の由来を説明している部分）を新しい順序に合わせて更新

### 2. ヘッダーのモバイルメニュー並び替え

`sys/03_frontend/web/src/components/Header.tsx`

- [ ] `learnLinks` 配列を `learn/page.tsx` の `categories` と同じ順序（LINEヤフー認定資格 → 運用・設定 → AI活用事例）に変更
- [ ] 配列直前のコメントを新しい順序に合わせて更新

### 3. 管理画面ダッシュボードの並び替え・カード追加

`sys/03_frontend/web/src/app/admin/page.tsx`

現状 `sections` に「LINEヤフー認定資格」（`/admin/quizzes`）へのカードが無い
（dev-plan-2-9 の時点でダッシュボードには元々無かったため復元対象外としていた）。
今回の指示で明示的に追加が必要になったため新設する。

- [ ] `sections` の先頭に `{ href: '/admin/quizzes', title: 'LINEヤフー認定資格', description: '...' }` を追加
      （description文言は「クイズ・検定（LINEヤフー認定資格）の作成・設問の管理」のような形を想定）
- [ ] 結果として `sections` は「LINEヤフー認定資格 → 記事 → 導入事例」の順になる
- [ ] グリッドは既存の `grid-cols-1 sm:grid-cols-3`（3カラム）のままで3項目にちょうど収まるため変更不要

### 4. 管理画面サイドナビの並び替え（指示の明示対象外だが一貫性のため）

`sys/03_frontend/web/src/app/admin/layout.tsx`

- [ ] `navLinks` を「ダッシュボード → LINEヤフー認定資格 → 記事 → 導入事例」の順に変更
      （ダッシュボードへのリンクは常に先頭のまま。ダッシュボードのカード順と揃える）
- [ ] これは指示に明記されていない拡張のため、不要であれば実装時に伝えてください（その場合はサイドナビの
      現状の並び「ダッシュボード → 記事 → 導入事例 → LINEヤフー認定資格」を維持する）

---

## 成果物

- `sys/03_frontend/web/src/app/learn/page.tsx`
- `sys/03_frontend/web/src/components/Header.tsx`
- `sys/03_frontend/web/src/app/admin/page.tsx`
- `sys/03_frontend/web/src/app/admin/layout.tsx`

## 完了条件

- `npm run build` / `npm test` / `npx playwright test` が全て成功
  （`sys/04_e2e/tests/public-content.spec.ts` の「top page links to all categories」テストは
  リンクの存在チェックのみで順序は見ていないため影響なし。ただし並び順を検証するテストが必要か
  実装時に判断する）
- 実ブラウザで `/learn` を確認: カードが「LINEヤフー認定資格 → LINE運用・設定 → 生成AI活用事例」の順、
  かつ3カード目まで均等にグリッドを埋めていて右側に空きが無いことを確認
- 実ブラウザでモバイル幅のヘッダーメニューを確認: サブリストが同じ新しい順序になっていることを確認
- 実ブラウザで `/admin` を確認: ダッシュボードカードが「LINEヤフー認定資格 → 記事 → 導入事例」の順で
  3枚とも表示されることを確認
- 実ブラウザで管理画面サイドナビを確認: 「ダッシュボード → LINEヤフー認定資格 → 記事 → 導入事例」の
  順になっていることを確認（4節参照、不要と言われた場合はスキップ）
