# Step — クイズ解説（explanation）の改行を表示に反映する

**対象:** クイズ機能の表示バグ修正
**依存:** なし

---

## 背景

管理画面の「解説」textarea で改行を入れて保存しても、解答画面・受験結果・マイページの受験詳細では
すべて1行の地の文として表示されてしまい、読みにくい（スクリーンショットで確認: `・` 区切りの箇条書きが
改行なしで連続表示されている）。原因は、`explanation` を素の `<p>{explanation}</p>` で描画しており、
HTML のデフォルトの空白折りたたみにより改行（`\n`）が無視されるため。

## タスク

- [ ] 以下4箇所の `explanation` 表示 `<p>` に `whitespace-pre-line`（連続スペースは畳むが改行は保持）を
      追加し、管理画面で入力した改行がそのまま表示されるようにする:
  - `sys/03_frontend/web/src/components/quiz/QuizQuestionCard.tsx:168`（単発モード解答結果）
  - `sys/03_frontend/web/src/components/quiz/ExamRunner.tsx:167`（検定モード結果の設問別レビュー）
  - `sys/03_frontend/web/src/components/mypage/AttemptDetailModal.tsx:56`（マイページの受験詳細モーダル）
  - `sys/03_frontend/web/src/app/admin/quizzes/[id]/edit/page.tsx:200`（管理画面の設問一覧プレビュー）

## 完了条件

- 管理画面で解説に改行を入れて保存すると、単発モード解答時・検定モード結果・マイページ受験詳細・
  管理画面のプレビューのいずれでも、入力した改行どおりに表示される。
