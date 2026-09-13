# Step 07 実装結果 — Next.js プロジェクト初期化・レイアウト・認証クライアント統合

**対応プラン:** [dev-plan-07-frontend-base.md](dev-plan-07-frontend-base.md)

---

## 実施内容

### 7.1 プロジェクト初期化

- プラン推奨どおり、既存 Vite プロジェクトを同じディレクトリ (`sys/03_frontend/web/`) で
  Next.js (App Router) に置き換え
- 削除したファイル（Vite 固有・Next.js と共存不可）:
  `vite.config.ts`, `index.html`, `tsconfig.json`, `tsconfig.app.json`, `tsconfig.node.json`,
  `package.json`, `package-lock.json`, `node_modules/`, `dist/`, `src/main.tsx`, `src/vite-env.d.ts`
  — すべて Git 管理下にあるため、必要なら履歴から復元可能
- **温存したファイル**（Step 08 で再利用予定、フレームワーク非依存のためそのまま残置）:
  `src/App.tsx`, `src/components/*`（Header/Footer は本 Step で更新、他は未着手）, `src/data/*.ts`,
  `src/index.css`, `assets/`, `public/`, `doc/`, `design-pattern/`
- Next.js 16.3.5 / React 19.3.0 / Tailwind CSS 4.3.3（`@tailwindcss/postcss`）で構築
- `@/*` エイリアスを `tsconfig.json` に踏襲

### 7.2 ディレクトリ構成

プラン記載の構成で作成（`(lp)/`, `learn/`, `usecase/`, `admin/`, `auth/` の各ルート、
`components/`, `lib/`, 型は今のところ各ファイル内で定義のため `types/` は未作成 — 必要になった時点で追加）

### 7.3 共通レイアウト・ヘッダー/フッター

- `src/app/layout.tsx` — 全体レイアウト（`AuthProvider` でラップ）
- `Header.tsx` / `Footer.tsx` を Next.js 用に更新（`<a href="#...">` → `next/link`、既存の
  LP内アンカーリンクは Step 08 でトップページ本文が復元されるまで `/#contact` 等の形で維持）
- **決定事項（プランで保留されていたナビ順序）**: 学習コンテンツの並び順を
  マーケティング講座 → 運用・設定 → AI活用事例 に決定し実装。README.md の該当の保留事項を更新済み

### 7.4 認証クライアント統合

- ログインボタン（LINE / Google）→ `NEXT_PUBLIC_API_URL` 配下の `/auth/line/login`,
  `/auth/google/login` への直接リンク
- **決定事項（プランで保留されていたセッション保持方式）**: React Context（`AuthProvider` /
  `useAuth`、`src/lib/auth.tsx`）を採用。マウント時に `GET /auth/me` を呼び、成功すればログイン状態、
  失敗（404 含む）は未ログイン扱いにフォールバック
- `src/app/auth/callback/page.tsx` — バックエンドのログインコールバック後に戻ってくる想定の
  ページ。`/auth/me` を再取得してトップへリダイレクト
- ヘッダーにログイン/ログアウト UI を実装（未ログイン時: LINE/Googleログインボタン、
  ログイン時: 表示名 + ログアウトボタン）

### 7.5 API クライアント

- `src/lib/api.ts` — `fetch` ベース、`credentials: 'include'` 付き、`NEXT_PUBLIC_API_URL`
  （デフォルト `http://localhost:8080`）を参照

---

## プラン外で追加対応したこと

- `sys/03_frontend/web/.gitignore` に `.next` を追加（Next.js のビルド出力を除外）
- `next.config.ts` に `output: 'standalone'` を設定（Dockerfile を小さく保つため。プランに
  明記はなかったが Dockerfile 作成に伴う自然な選択）
- `next dev` / `next build` が自動生成した `AGENTS.md` / `sys/03_frontend/web/CLAUDE.md`
  （`@AGENTS.md` を import するだけの内容）はそのまま残置。Next.js 側のツールが管理するファイルで
  削除しても再生成されるため、コミットする方針とした（このバージョンの Next.js が学習データと
  異なる可能性がある旨の注意書きが入っている）

## プラン未実施（意図的にスキップ / 後続 Step 待ち）

- `/learn/`, `/usecase/`, `/admin` の実コンテンツ（それぞれ Step 09〜11）— 現状はプレースホルダーページ
- Admin アクセス制御（`role=admin` チェック）— Step 11
- 実際のログイン成功フロー — バックエンドの `/auth/line/login` 等（Step 04）が存在しないため、
  現時点でログインボタンを押すと 404 になる（Step 03 の DB スタブと同じ「依存 Step 未着手」の状態）

---

## 検証

- `npm run build` — 成功（`/`, `/learn`, `/usecase`, `/admin`, `/auth/callback` すべて静的生成）
- `npm run dev` → `curl` で `/`, `/learn`, `/usecase` が `200`
- `docker build` → 単体コンテナ起動 → `curl` で `/`, `/learn` が `200`
- `docker compose up -d --build`（`line-web` + `line-api` 両方）→
  `line-web:/` `200`、`line-web:/learn` `200`、`line-api:/health` `200`、
  `docker compose ps` で両サービス `Up` を確認
- 検証用に作成したコンテナ・イメージはすべて後片付け済み（`docker compose down` + `docker rmi`）

---

## 完了条件との対比

| 完了条件（プラン記載） | 状態 |
|---|---|
| `npm run dev` で Next.js アプリが起動する | ✅ |
| LINE / Google でログインでき、ヘッダーにログイン状態が反映される | △ UI・導線は実装済みだが、Step 04（バックエンド認証）待ちのため実際のログインは未検証 |
| `/learn/`, `/usecase/`, `/admin/` への空のルートが用意されている | ✅ |

## 次のステップ

- `dev-plan-08-frontend-lp.md`（既存 LP コンテンツの移行）
- `dev-plan-04-auth.md`（実装すればログインフローが最後まで通る）
