# line.omise.app

はんなりdev's LINE mini-app platform, running on [line.omise.app](https://line.omise.app). Ships three LINE mini-apps and authenticates users via **LINE Login** and **Google OAuth**.

> **Status**: this repo is transitioning from a static marketing landing page (React + Vite) into the real application (Next.js frontend + Go Gin backend). Nothing below describing the new architecture is implemented yet — it is tracked and built via dev plans in [`doc/dev-plans/`](doc/dev-plans) per the workflow in [`CLAUDE.md`](CLAUDE.md).

---

## Mini-Apps

| App | Description |
| --- | --- |
| 会員管理アプリ (Membership) | Digital membership card, points, and rank tiers, run entirely inside LINE. |
| サロン予約アプリ (Salon Reservation) | 4-step booking flow with staff scheduling and LINE notifications. |
| スイーツショップアプリ (Sweets Shop) | Product gallery, points QR code, and reviews for a sweets shop. |

---

## Content Site Map (`/learn/`, `/usecase/`)

URL structure for SEO-driven content marketing (how-to articles, client case studies) — supports the SEO goals in `sys/03_frontend/web/doc/seo-update-2026-02-16.md`. `/learn/` implemented in `dev-plan-09-frontend-learn`; `/usecase/` not yet implemented (`dev-plan-10-frontend-usecase`).

| Path | Format | Contents |
| --- | --- | --- |
| `/learn/` | — | Learning content top |
| `/learn/line-operation/` | Article list | LINE運用・設定 — practical how-to (rich menu setup, official account initial setup, Messaging API config, reply-mode usage, etc.) |
| `/learn/line-operation/{article-slug}/` | — | e.g. `/learn/line-operation/rich-menu-setup/` |
| `/learn/ai/` | Article list | 生成AI活用事例 |
| `/learn/ai/{article}/` | — | Individual article |
| `/learn/line-yahoo-certification/` | Quiz/exam list | LINEヤフー　認定資格勉強 — same クイズ・検定 engine from `dev-plan-2-4-frontend-quiz-ui`, rebranded/moved here in `dev-plan-2-9-line-yahoo-certification` (マーケティング講座 was removed instead) |
| `/learn/line-yahoo-certification/{quiz-slug}/` | — | Individual quiz/exam |
| `/usecase/` | — | 導入店舗インタビュー (client store interviews) |
| `/usecase/{client}/` | — | Individual store interview |

`line-operation` and `ai` are flat, on-demand article lists sharing one pattern. If `line-operation` mixes setup articles (account creation, rich menu, reply settings) and operation articles (broadcast tips, step campaigns, friend-add tactics), keep the URL flat (`/learn/line-operation/{slug}/`) and filter by tag (`?tag=rich-menu` or `/learn/line-operation/tag/rich-menu/`) instead of adding directory levels.

Category name decided (`dev-plan-09-frontend-learn`): **`line-operation`** — already committed to by the `articles.category` CHECK constraint since `dev-plan-02-database` and the `/api/articles` handlers since `dev-plan-05-content-api`, so it's the value in actual use rather than a still-open choice.

Nav order for 学習コンテンツ's sections was decided in `dev-plan-07-frontend-base`: originally マーケティング講座 → 運用・設定 → AI活用事例 → クイズ・検定 (`dev-plan-2-4-frontend-quiz-ui`). In `dev-plan-2-9-line-yahoo-certification`, マーケティング講座 (courses/lessons, Phase 1) was removed entirely; クイズ・検定 (Phase 2) was kept — same backend (`quizzes`/`quiz_questions`/`quiz_choices`/... tables, `/api/quizzes` etc.) and same UI components, just moved from `/learn/quiz` to `/learn/line-yahoo-certification` and relabeled — giving the current order 運用・設定 → AI活用事例 → LINEヤフー認定資格勉強, see `sys/03_frontend/web/src/components/Header.tsx`.

---

## Architecture

```
                  +-----------------------+
                  |     Cloudflare CDN    |
                  |   (DNS + Edge Cache)  |
                  +-----------+-----------+
                              |
                  +-----------+-----------+
                  |      Traefik v3       |
                  |  (Reverse Proxy, TLS) |
                  +-----------+-----------+
                      |               |
              +-------+-----+   +-----+------+
              | Next.js     |   |  Go (Gin)  |
              | (Web Front) |   |   API      |
              +-------------+   +------------+
```

Traefik and Cloudflare are shared across VPS projects, not started by this repo — see [`hannari.dev/cloudflare/log`](https://github.com/tanoshimi-dev) README for how a new service joins the shared `traefik-network`. Planned domains: `line.omise.app` (frontend) and `api-line.omise.app` (backend).

### Authentication Flow

```
Client (Web) --> LINE Login / Google OAuth --> ID Token / Access Token
    --> API Request: Authorization: Bearer <token>
    --> Go (Gin) Backend: verifies token with the provider
    --> users table upsert (provider + provider user ID as key)
```

Providers: LINE Login, Google OAuth.

---

## Tech Stack

| Layer        | Technology       | Notes                                  |
| ------------ | ---------------- | --------------------------------------- |
| Web Frontend | Next.js          | Replaces the current React + Vite site |
| Backend      | Go / Gin         |                                         |
| Auth         | LINE Login, Google OAuth |                                 |
| CDN          | Cloudflare       | DNS + Edge Cache                       |
| Proxy        | Traefik v3       | Auto TLS, routing, shared across VPS   |
| Container    | Docker Compose   |                                         |

Everything else (database, caching, hosting details for each mini-app) is undecided and will be settled in `doc/dev-plans/` before implementation.

---

## Repo Layout

```
doc/
├── architecture-decision-record/
└── dev-plans/                 # Plan-then-implement docs (see CLAUDE.md)

sys/
├── 01_infra/                  # Reserved
├── 02_backend/                # Go (Gin) API — not yet implemented
├── 03_frontend/web/           # Currently the React + Vite landing page;
│                               # migrating to Next.js
├── 04_e2e/                    # Reserved
└── secrets/                   # Local secrets only, gitignored
```

---

## Development Workflow

This repo follows a strict plan-then-implement flow — see [`CLAUDE.md`](CLAUDE.md) for the full rule:

1. Write/update a plan in `doc/dev-plans/dev-plan-<slug>.md`.
2. Only once told to implement, build it and record the outcome in `doc/dev-plans/dev-plan-<slug>-result.md`.
3. Git operations (`add`/`commit`/`push`) are done by the repo owner, not Claude.

---

## License

Private
