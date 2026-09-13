# Step 11 実装結果 — 管理画面（コンテンツ作成・更新、Admin 限定）

**対応プラン:** [dev-plan-11-frontend-admin.md](../dev-plan-11-frontend-admin.md)

---

## 実施内容

### バックエンド追加（実装に必須だった判断）

管理画面が下書きを含む全コンテンツを一覧・編集するには、既存の Admin API（POST/PUT/DELETE のみ）
だけでは足りず、GET 系の管理者用エンドポイントが1つも無かった（公開用 GET は
`status=published` のみを返すため、下書きの編集フォームを事前入力できない）。以下を追加:

- `GET /api/admin/courses` / `GET /api/admin/courses/:id`（全ステータス、レッスン同梱）
- `GET /api/admin/lessons/:id`
- `GET /api/admin/articles` / `GET /api/admin/articles/:id`（`category` クエリは任意）
- `GET /api/admin/usecases` / `GET /api/admin/usecases/:id`
- `GET /api/admin/lessons/:id/exam`（Admin向け — `is_correct` を含む。Reader向けの
  `GET /api/lessons/:lessonId/exam` とは別レスポンス）

`CourseRepository.ListAll` / `ArticleRepository.ListAll` / `UsecaseRepository.ListAll` を
リポジトリ層に追加（既存の `ListPublished` とは別に、ステータス条件なしで取得）。
ルーティングで `/admin/lessons/:id` と `/admin/lessons/:id/exam` の両方に `:id` を使うよう統一
（既存の `/admin/lessons/:id`（PUT/DELETE）と異なるワイルドカード名を同じ位置に登録すると
Gin のルーターが衝突しうるため）。

### 11.1 アクセス制御

- **決定**: `/admin` 配下は `src/app/admin/layout.tsx`（Server Component）で一括ガード。
  `src/lib/serverAuth.ts`（新規）が `next/headers` の `cookies()` からセッション Cookie を読み、
  バックエンドの `GET /auth/me` にそのまま転送して `role` を確認する
  - **dev-plan-09 の方針からの意図的な例外**: dev-plan-09-frontend-learn は
    「ユーザー固有データはブラウザ側フェッチに限定」という境界を敷いたが、
    本タスクはプランが明示的に「サーバーコンポーネント側で確認する仕組み」を要求しているため、
    Cookie をサーバー側に転送する例外を設けた（クライアント側チェックでは
    保護対象コンテンツのレンダリング自体を止められないため、この目的には使えない）
  - 未ログイン・Reader ともに `/` へリダイレクト（全ページのヘッダーに
    LINE/Google ログインボタンが常設されているため、専用の「ログインへ」ページを
    別途用意しなくても事実上「ログインへ」の導線になっている）

### 11.2〜11.4 講座・記事・事例管理

- `/admin`（ダッシュボード）、`/admin/courses`（一覧/新規/編集）、
  `/admin/courses/:id/lessons/new`、`/admin/lessons/:id/edit`、
  `/admin/articles`（一覧/新規/編集、タグ追加）、`/admin/usecases`（一覧/新規/編集）
- 共通コンポーネント化（プラン記載の指示に対応）:
  - `src/components/admin/FormField.tsx` — `TextField`/`TextAreaField`/`NumberField`/
    `StatusSelect`/`SubmitButton`
  - `src/components/admin/DeleteButton.tsx` — 確認ダイアログ付き削除ボタン
    （一覧ページでは `onDeleted` コールバックでローカル state を更新、編集ページでは
    `redirectTo` で遷移 — 一覧ページはクライアント側 `useEffect` でデータ取得しているため
    `router.refresh()` では再取得されないことに注意して設計）
  - `src/components/admin/UsecaseForm.tsx` — 新規/編集で共有するフォーム本体
- Markdown エディタは指示通りシンプルな `<textarea>`（リッチエディタは任意扱いのため未実装）
- タグ付け UI: 追加のみ対応。dev-plan-05 の Admin API がタグの付与（POST）のみを提供し、
  削除（detach）エンドポイントが元々存在しないため、UI もそれに合わせた
  （新しい API を追加はしていない — Step05 時点のスコープ外の機能を今回追加するのは
  見送り、既存 API の実能力に忠実な UI とした）

### 11.5 試験管理

- `/admin/lessons/:id/exam` — 試験が無ければ作成フォーム、あれば設問一覧+追加/編集/削除。
  設問フォーム（`QuestionForm`、ページ内ローカルコンポーネント）は追加・編集の両方で共有し、
  選択肢の動的な追加・削除、正解チェックボックスに対応
  - **既知の制約**: 試験自体（タイトル・合格点）の更新・削除 API が dev-plan-06 の時点で
    そもそも存在しない（当時のプランが設問の作成/更新/削除のみを要求していたため）。
    本 Step でも新設せず、既存 API の範囲内で UI を実装した

---

## プラン外で追加対応したこと

- **ログアウト時のバグ発見・修正**: `/admin` 配下は Server Component（layout.tsx）でのみ
  ガードしているため、`/admin/*` を閲覧中にヘッダーの「ログアウト」を押すと
  クライアント側の認証状態はクリアされるが、画面遷移が発生せず管理画面のUIがそのまま
  表示され続けてしまうことが分かった（サーバー側ガードはナビゲーション時にしか再評価されない）。
  `src/lib/auth.tsx` の `logout()` を修正し、ログアウト後に `window.location.href = '/'`
  でフルリロードするようにした（`router.push` ではブラウザの戻る操作でクライアント側の
  キャッシュされた管理画面が再表示されるリスクが残るため、あえてハードナビゲーションを選択）
- `src/lib/api.ts` に `put` / `delete` メソッドを追加（既存は `get`/`post` のみだった）
- 管理画面の `<title>` に `robots: { index: false, follow: false }` を設定
  （検索エンジンにインデックスさせない）

## プラン未実施

なし。11.1〜11.5 すべて実装済み。

---

## 検証

- `go build ./...` / `go vet ./...` / `gofmt -l .`（差分なし）— バックエンド成功
- `npm run build` — 成功。`/admin/*` は `layout.tsx` が `cookies()`（dynamic API）を使うため
  全ルートが正しく動的レンダリング（`ƒ`）と判定されることを確認
- 新規 Admin GET エンドポイントを `curl` + 署名付きテストセッションで確認:
  `GET /api/admin/courses`, `/api/admin/courses/:id`, `/api/admin/articles`,
  `/api/admin/usecases`, `/api/admin/lessons/:id/exam`（試験未作成時は `404`）—
  いずれも期待通りのレスポンス
- **アクセス制御の確認方法について**: `/admin` 配下は全ページが認証必須のため、
  dev-plan-09/10 のように「未ログインの見た目だけブラウザで確認、認証ロジックは
  API レベルで確認」という切り分けが本 Step には使えなかった。ユーザーに実際に
  Google アカウント（`keijimitaki@gmail.com`）を `ADMIN_EMAILS` に追加した上で
  再ログインしてもらい、実際の Admin セッションで `/admin` にアクセスできることを
  **ユーザー自身に確認してもらった**（本セッションの Chrome 拡張タブは別のブラウザ
  プロファイル文脈のため Cookie を共有できず、Claude 側から直接操作しての確認はできなかった）
- Claude 側では未ログイン状態での `/admin` アクセスが `/` へリダイレクトされることを
  ブラウザで確認（認証不要な部分の動作確認として）
- 検証で作成したテストセッション（`admin-test-11` 等）はすべて DB から削除済み。
  `ADMIN_EMAILS` へのユーザー本人のメールアドレス追加は実運用上必要な変更のため保持

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| Admin ユーザーが講座・レッスン・記事・事例・試験を作成・編集・削除できる | ✅（ユーザー自身が `/admin` へのアクセスを確認。個々の CRUD 操作は API レベルで確認済みのエンドポイントを UI から呼び出す実装） |
| Reader / 未ログインユーザーは `/admin` にアクセスできない | ✅（未ログインをブラウザで確認。Reader は `role` チェックのロジックが未ログインと同一分岐のため同様に機能する） |
| 下書き（draft）コンテンツが公開ページに表示されないことを確認できる | ✅（dev-plan-05/06/09/10 で既に確認済みの公開APIの絞り込みロジックは本Stepで変更していない） |

**Phase 1 — Frontend（Step 07〜11）完了。**

## 次のステップ

- `dev-plan-12-test-phase1.md`（Go / Next.js / E2E テスト）
- `dev-plan-13-deploy-phase1.md`（本番デプロイ）
- 任意の改善候補（未実施、必要になれば別 Step で）: タグの detach API、試験自体の更新/削除 API、
  Markdown プレビュー
