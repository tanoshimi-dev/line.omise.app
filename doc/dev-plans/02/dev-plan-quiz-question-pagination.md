# Step — 検定モードも1問ずつ表示 + 両モード共通の前へ／次へナビゲーション

**対象:** [dev-plan-quiz-mode-selection.md](dev-plan-quiz-mode-selection.md) 実装済みのクイズ解答 UI の追加修正
**依存:** dev-plan-quiz-mode-selection（実装済み）

---

## 背景

現状:

- `QuizQuestionCard`（単発モード）は1問ずつ表示するが、「次の問題へ」しかなく、一度進むと前の設問には
  戻れない。
- `ExamRunner`（検定モード）は全設問を1ページにまとめて表示し、まとめて「提出する」方式（
  `sys/03_frontend/web/src/components/quiz/ExamRunner.tsx:69-106`）。

今回、**検定モードも単発モードと同じく1問ずつページ送りで表示し、両モードとも「前の問題へ」「次の問題へ」で
自由に行き来できるようにする**。バックエンドの API・レスポンス形状は変更しない
（検定モードは引き続き `POST /api/quizzes/:slug/submit` に全設問の解答をまとめて送信し、まとめて採点される
— 変わるのはフロントエンドの見せ方のみ）。

---

## 設計

### 検定モード（ExamRunner）

- 全設問を一度に描画するのをやめ、`index` state で現在の設問を1問だけ表示する（`QuizQuestionCard` と同じ
  レイアウト: 進捗表示「N / M問」+ 設問 + 選択肢）。
- 選択済みの回答はこれまで通り `answers: Record<string, string[]>` にため込み、まだ API へは送らない
  （採点は最後にまとめて `POST /api/quizzes/:slug/submit` を1回呼ぶ、という既存の挙動は維持）。
- ナビゲーション:
  - 「前の問題へ」: `index > 0` の間は常に表示・有効。未回答のまま戻れる。
  - 「次の問題へ」: 最後の設問以外で表示。回答の有無に関わらず進める（検定モードは未回答のまま提出できる
    既存仕様 = 未回答は不正解扱い、と一致させるため、次へ進むのに解答必須にはしない）。
  - 最後の設問でだけ「提出する」ボタンを表示し、クリックで既存の `handleSubmit`（未回答確認ダイアログ含む）
    を呼ぶ。

### 単発モード（QuizQuestionCard）

- 現状は `QuestionStep` が設問ごとにアンマウント/マウントされ、`result`（採点結果）はローカル state
  なので前の設問に戻ると採点結果が失われる。「前の問題へ」を成立させるため、親コンポーネント
  （`QuizQuestionCard`）に次の state を持たせる:
  - `selectedByQuestion: Record<string, string[]>` — 各設問で選択した選択肢
  - `resultByQuestion: Record<string, QuizAnswerResult>` — 各設問で API から返った採点結果（一度採点した
    設問は再送しない。バックエンドの `POST /api/quiz-questions/:id/answer` はモード非依存で何度でも
    呼べる仕様だが、UI 上は「戻って結果を見返す」動作にし、二重送信はしない）
- ナビゲーション:
  - 「前の問題へ」: `index > 0` の間は常に表示・有効。戻った設問がすでに採点済みなら、その結果
    （正誤・解説）をそのまま再表示する。
  - 「次の問題へ」: 現在の設問が採点済みの場合のみ表示（未回答のまま次に進むことはできない — 単発モードの
    既存仕様を維持）。最後の設問を採点し終えたら、ボタンラベルを「結果を見る」にして `finished` 画面へ。

### 両モード共通

- 進捗表示は常に「N / M問」形式で現在位置を示す。
- 「前の問題へ」は最初の設問（index 0）では非表示 or 無効化。

---

## タスク

- [ ] `sys/03_frontend/web/src/components/quiz/ExamRunner.tsx`: 全設問一括表示をやめ、`index` による
      1問ずつのページ送り + 前へ／次へボタンに変更。最後の設問でのみ「提出する」ボタンを表示。
- [ ] `sys/03_frontend/web/src/components/quiz/QuizQuestionCard.tsx`: 採点結果を親の state
      （`selectedByQuestion`/`resultByQuestion`）に引き上げ、「前の問題へ」ボタンを追加。
- [ ] 両コンポーネントの既存 Vitest（`ExamRunner.test.tsx`, `QuizQuestionCard.test.tsx`）を新しい
      ページ送り UI に合わせて更新し、前へ／次への行き来を検証するケースを追加。
- [ ] `sys/04_e2e/tests/quiz.spec.ts` の検定モードのシナリオ（全問1ページ表示・即座に「提出する」を
      クリックする現行の操作列）を、1問ずつ選択して次へ進む操作列に更新。

## 完了条件

- 検定モードで、設問が1問ずつ表示され、「前の問題へ」「次の問題へ」で自由に行き来できる。
- 単発モードで、採点済みの設問に「前の問題へ」で戻ると、選択した回答と正誤・解説がそのまま再表示される
  （API を再送しない）。
- 検定モードの最終提出（`POST /api/quizzes/:slug/submit` へのリクエスト内容・未回答確認ダイアログ）は
  従来どおり動作する。
- 既存の Vitest・Playwright テストを更新のうえ green。
