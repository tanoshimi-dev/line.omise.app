# Step — 「LINEヤフー　認定資格勉強」ラベルを「LINEヤフー認定資格」に短縮

**対象:** ナビ・学習コンテンツ表示まわりの文言修正 — 特定フェーズに属さない単発の修正
**依存:** [dev-plan-2-9-line-yahoo-certification.md](02/dev-plan-2-9-line-yahoo-certification.md)（このラベル自体を導入した Step）

---

## 背景

ユーザー指示: 「Change MENU label LINEヤフー　認定資格勉強 to LINEヤフー認定資格」
（全角スペース入り「LINEヤフー　認定資格勉強」→ 「LINEヤフー認定資格」、スペースなし・「勉強」を落とした短い表記）

指示は「MENU label」（ナビゲーションメニュー項目）を指しているが、このフレーズは
`dev-plan-2-9-line-yahoo-certification` で意図的にナビ・学習コンテンツ一覧カード・ページ見出し・
マイページ・管理画面の全箇所で統一されている。ナビだけ変えてページ見出しを変えないと、
メニューをクリックした先の表示と文言が食い違うため、**同じ文字列が出ている箇所は全て統一して変更する**。

---

## 変更対象（`grep -rn 'LINEヤフー　?認定資格勉強'` で洗い出し済み、全8ファイル）

| ファイル | 種別 |
|---|---|
| `sys/03_frontend/web/src/components/Header.tsx` | デスクトップ/モバイルナビの `learnLinks` ラベル（本体） |
| `sys/03_frontend/web/src/app/learn/page.tsx` | `/learn` トップページのカードタイトル・`metadata.description` |
| `sys/03_frontend/web/src/app/admin/layout.tsx` | 管理画面サイドナビのラベル |
| `sys/03_frontend/web/src/app/admin/quizzes/page.tsx` | 管理画面の一覧見出し `<h1>` |
| `sys/03_frontend/web/src/components/mypage/QuizProgressSection.tsx` | マイページのセクション見出し `<h2>` |
| `sys/03_frontend/web/src/app/learn/line-yahoo-certification/page.tsx` | 公開一覧ページの `metadata.title`・`<h1>` |
| `sys/03_frontend/web/src/app/learn/line-yahoo-certification/[slug]/page.tsx` | 「〜一覧へ戻る」リンク文言 |
| コメント各所（`Header.tsx`, `learn/page.tsx`, `learn/me/page.tsx`, `learn/me/page.test.tsx`） | コード上のコメントのみ — 動作に影響なし、文言変更に合わせて更新するかは任意 |

`description`/`metadata.description` など「LINEヤフー認定資格の取得に向けて…」のような文中で
自然文として使われている箇所（`learn/page.tsx` のカード説明文、`line-yahoo-certification/page.tsx` の
本文中の説明文）は、ラベルそのものではないため対象外（現状の「LINEヤフー認定資格の取得に向けて」という
言い回しのまま変更不要）。

## タスク

- [ ] 上表のラベル文字列（`LINEヤフー　認定資格勉強`）を全て `LINEヤフー認定資格` に置換
- [ ] `sys/03_frontend/web/src/app/learn/me/page.test.tsx` のアサーション文字列
      （現状 `/LINEヤフー\s*認定資格勉強/` の正規表現マッチャー）も新ラベルに更新
- [ ] `sys/04_e2e/tests/public-content.spec.ts`・`quiz.spec.ts` の `getByRole('link'/'heading', { name: 'LINEヤフー　認定資格勉強' })` 系アサーションを新ラベルに更新
- [ ] コード上のコメント（Header.tsx 等）も新ラベルに合わせて更新（任意だが表記統一のため実施）

## 完了条件

- `npm run build` / `npm test` / `go test ./...` / `npx playwright test` が全て成功
- 実ブラウザで、デスクトップ/モバイルナビ・`/learn` トップページ・`/learn/line-yahoo-certification`・
  マイページ・管理画面ナビ・管理画面一覧見出しの全てで「LINEヤフー認定資格」に統一されていることを確認
