# Step 2-11 — 自分のアカウント削除（退会）

**対象:** ログイン済みユーザーが自分自身のアカウントと紐づく利用データを完全に削除できるようにする  
**依存:** Phase 1 の認証基盤（`dev-plan-04-auth.md`）、Phase 2 のクイズ履歴（`dev-plan-2-5-frontend-mypage.md`）

---

## ゴール

ログイン中のユーザーがマイページから、明示的な最終確認を経て自分のアカウントを完全に削除できるようにする。
削除後は、保存されたクイズ解答・検定受験履歴を含む**本人に関連するすべての現行データ**と全セッションを削除し、当該ブラウザの認証 Cookie も削除する。

## 方針・対象範囲

- API は `DELETE /auth/me` とし、既存の `GET /auth/me` と同じく `requireReader` を必須にする。
- 削除対象は認証済みセッションから得た本人の `users` レコードのみとする。リクエストの body、path、query でユーザー ID を受け取らず、他人や任意のアカウントを指定できないようにする。
- `users` をハードデリートする。バックアップ・復元用の論理削除レコード、削除済みユーザーのプロフィール、クイズ履歴は保持しない。
- 現行スキーマにあるユーザー関連データは、以下の `ON DELETE CASCADE` により漏れなく削除される。
  - `sessions`
  - `user_quiz_attempts`（検定の受験履歴）と、それに紐づく `user_quiz_answers`（検定中の全解答）
  - `user_quiz_answers`（単発練習モードの全解答履歴）
- 現行の Phase 2 スキーマには講座進捗テーブルは存在しない。将来、ユーザー所有データを追加するときは、アカウント削除で当該データも必ず消えるよう `users(id)` に対する `ON DELETE CASCADE` を必須の設計ルールとする。
- LINE / Google 側のアカウントや OAuth 連携許可を取り消す機能は対象外。再度同じプロバイダーでログインした場合は、新しい空のローカルアカウントが作成される。
- 管理者も自分のアカウントを削除できる。`ADMIN_EMAILS` に登録されたメールアドレスで再ログインした場合は、既存の昇格ルールに従い再び admin になる。
- 退会後の復元猶予・ソフトデリート・管理者による他ユーザー削除は対象外。

---

## タスク

### 2-11.1 バックエンド — ユーザー削除リポジトリ

`sys/02_backend/internal/repository/user.go`

- [ ] `DeleteByID(ctx context.Context, id int64) error` を追加する。
- [ ] `DELETE FROM users WHERE id = $1` を実行し、削除件数が 0 件なら既存の not-found エラーを返す。
- [ ] 個別のセッション削除やクイズ履歴削除をアプリケーション側で重複実装せず、外部キーの cascade に一任する。

### 2-11.2 バックエンド — 退会 API

`sys/02_backend/internal/handler/auth.go`

- [ ] `DeleteMe` ハンドラを追加する。
- [ ] `middleware.CurrentUser(c)` から本人の ID を取得し、`Users.DeleteByID` を呼ぶ。
- [ ] 成功時は `line_omise_session` Cookie を、既存 `Logout` と同じ属性（`Path=/`、`SameSite=Lax`、`HttpOnly`、環境に応じた `Secure`）で期限切れにする。
- [ ] 成功時は `204 No Content` を返す。セッション行も cascade で消えるため、他デバイスを含む全ログイン状態が無効になる。
- [ ] 想定外の DB エラーは 500 にし、ユーザー ID や内部エラー情報をレスポンスに含めない。

`sys/02_backend/internal/server/server.go`

- [ ] `authGroup.DELETE("/me", requireReader, authHandler.DeleteMe)` を登録する。
- [ ] 既存の CORS と SameSite Cookie の設定を維持し、認証付きの破壊的操作をクロスオリジンで許可しない。

### 2-11.3 フロントエンド — マイページの退会 UI

`sys/03_frontend/web/src/components/mypage/DeleteAccountSection.tsx`（新規）

- [ ] マイページ最下部に、他の操作と視覚的に区別した「退会」セクションを配置する。
- [ ] 削除される内容（アカウント、ログイン状態、保存済みクイズ解答・受験履歴）と、元に戻せないこと、再ログイン時には履歴が復元されないことを日本語で明示する。
- [ ] 削除ボタンは即時 API 呼び出しにせず、確認ダイアログを表示する。最終確認ではユーザーに指定語句（例: `退会する`）を入力させ、完全一致するまで実行ボタンを無効にする。
- [ ] 確認後に `DELETE /auth/me` を一度だけ呼び出し、送信中は重複送信を防ぐ。
- [ ] 成功時は認証コンテキストのユーザー状態を `null` にし、完全なページ遷移でトップページへ移動する。API が Cookie を消せなかった場合にも、画面側にはログイン済み表示を残さない。
- [ ] 失敗時は退会していないことが分かるエラーメッセージを表示し、再試行できるようにする。

`sys/03_frontend/web/src/app/learn/me/page.tsx`

- [ ] ログイン済み表示の `QuizProgressSection` の後に `DeleteAccountSection` を追加する。
- [ ] 未ログイン時には退会 UI を表示しない。

`sys/03_frontend/web/src/lib/auth.tsx`

- [ ] 退会成功後に auth state を安全に破棄するため、必要に応じて `clearUser` / `deleteAccount` のような小さな公開操作を追加する。`logout` の既存の全ページ遷移方針を踏襲する。

### 2-11.4 テスト

バックエンド（`sys/02_backend/internal/handler/auth_test.go`、`internal/repository/user_test.go`）

- [ ] 未認証の `DELETE /auth/me` が 401 になることを確認する。
- [ ] 認証済みユーザーの削除が 204 になることを確認する。
- [ ] 削除後、削除に使用した Cookie で `/auth/me` にアクセスすると 401 になることを確認する。
- [ ] 同ユーザーの `users`、`sessions`、`user_quiz_attempts`、`user_quiz_answers` が削除されることを、テストデータを作成して確認する。
- [ ] テストでは単発練習の解答と検定受験（その配下の複数解答）の両方を作成し、退会後に残存行が 0 件であることを確認する。
- [ ] 別ユーザーのレコード、セッション、履歴は残ることを確認する。

フロントエンド（`sys/03_frontend/web/src/components/mypage/DeleteAccountSection.test.tsx` など）

- [ ] 退会説明と確認ダイアログが表示されることを確認する。
- [ ] 確認語句が一致しない間は実行できないことを確認する。
- [ ] 成功時に DELETE を 1 回だけ呼び、トップへ遷移することを確認する。
- [ ] API 失敗時はエラーを表示し、画面をログアウト状態へ遷移させないことを確認する。

### 2-11.5 検証

- [ ] `sys/02_backend` で `go test ./... -v` を実行する。
- [ ] `sys/03_frontend/web` で `npm test` と `npm run build` を実行する。
- [ ] ローカル環境で reader アカウントを用い、退会確認、削除後のトップ遷移、再ログイン時に空の履歴になることを確認する。

---

## 成果物

- `sys/02_backend/internal/repository/user.go` — 本人アカウント用の削除メソッド
- `sys/02_backend/internal/handler/auth.go` — `DELETE /auth/me`
- `sys/02_backend/internal/server/server.go` — 認証必須ルート
- `sys/03_frontend/web/src/components/mypage/DeleteAccountSection.tsx` — 二段階確認付き退会 UI
- 関連する Go / Vitest テスト

## 完了条件

- ログイン中のユーザーだけが、自分のアカウントを明示的な最終確認後に削除できる。
- API はクライアント指定の ID を受け付けず、他ユーザーを削除できない。
- 削除によって本人のすべてのセッションと、練習・検定を含む現行のユーザー所有データが残さず消え、他デバイスのセッションも無効になる。
- 退会後に UI がログイン済み状態を保持せず、再ログインした場合に旧データが復元されない。
- 既存テストと追加テストがすべて成功する。

## 実装時の確認事項

- 本番環境の DB に、将来追加されたユーザー参照テーブルで `ON DELETE RESTRICT` / `NO ACTION` がないことを、マイグレーション適用前に確認する。
- プライバシーポリシーまたは利用規約に保管義務のある監査ログが後から追加される場合は、アカウント削除前に法務・運用方針を確認し、このハードデリート方針と整合させる。
