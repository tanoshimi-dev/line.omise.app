# Step 07 — Next.js プロジェクト初期化・レイアウト・認証クライアント統合

**フェーズ:** Phase 1 — Frontend
**依存:** Step 04（認証 API が最低限動作していること）

---

## ゴール

`sys/03_frontend/web/` の現行 React + Vite ランディングページを置き換える Next.js
プロジェクトの土台（App Router、共通レイアウト、LINE Login / Google OAuth のログイン導線）
を構築する。

---

## タスク

### 7.1 プロジェクト初期化

- [ ] 既存の `sys/03_frontend/web/`（Vite）をどう扱うか決定:
  - 推奨: 新しい Next.js プロジェクトを同じディレクトリに再構築し、既存コンテンツ（`src/data/*`,
    `assets/`, `public/images/*` 等）は Step 08 で移植・再利用する
- [ ] Next.js (App Router) + TypeScript + Tailwind CSS で初期化
- [ ] `@/*` エイリアス設定を踏襲（`tsconfig.json` の `paths`）

### 7.2 ディレクトリ構成

```
sys/03_frontend/web/
├── src/
│   ├── app/
│   │   ├── (lp)/              # トップページ（Step 08）
│   │   ├── learn/             # /learn/ 配下（Step 09）
│   │   ├── usecase/           # /usecase/ 配下（Step 10）
│   │   ├── admin/             # 管理画面（Step 11、Admin 限定）
│   │   └── auth/              # ログインコールバックページ
│   ├── components/
│   ├── lib/
│   │   ├── api.ts             # バックエンド API クライアント
│   │   └── auth.ts            # ログイン状態管理ユーティリティ
│   └── types/
└── Dockerfile
```

### 7.3 共通レイアウト・ヘッダー/フッター

- [ ] `src/app/layout.tsx` — 全体レイアウト（既存 `Header` / `Footer` を移植）
- [ ] ハンバーガーメニューに「学習コンテンツ」（マーケティング講座 / 運用・設定 / AI活用事例）と
  「導入事例」を追加（メニュー内の並び順は README で保留中 — このステップで確定する）

### 7.4 認証クライアント統合

- [ ] ログインボタン（LINE / Google）→ バックエンドの `/auth/line/login`, `/auth/google/login` へ遷移
- [ ] コールバック後のセッション確認: `GET /auth/me` を叩いてログイン状態を取得
- [ ] ログイン状態を保持する仕組みを決定（React Context / サーバーコンポーネントで都度 `me` 取得 等）
- [ ] ログイン/ログアウト UI（ヘッダーに表示）

### 7.5 API クライアント

- [ ] `src/lib/api.ts` に `fetch` ベースの薄い API クライアントを実装
- [ ] `NEXT_PUBLIC_API_URL=https://api-line.omise.app` を環境変数化

---

## 成果物

- Next.js プロジェクト一式（上記ディレクトリ構成）
- 共通レイアウト・ログイン導線
- `src/lib/api.ts`

## 完了条件

- `npm run dev` で Next.js アプリが起動する
- LINE / Google でログインでき、ヘッダーにログイン状態が反映される
- `/learn/`, `/usecase/`, `/admin/` への空のルートが用意されている（中身は後続 Step）
