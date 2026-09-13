# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development workflow

This repo follows a strict plan-then-implement flow. **Do not write or change any code or configuration until explicitly told to implement it** — even if the request seems small or obvious.

1. **Plan**: When asked to create or update a dev plan, write/update `doc/dev-plans/dev-plan-<slug>.md`. This step is planning only — no code or config changes.
2. **Implement**: Only once explicitly told to implement, carry out the plan, then write a result document to `doc/dev-plans/dev-plan-<slug>-result.md` summarizing what was done (and any deviation from the plan).
3. **Git**: Never run `git add`, `git commit`, or `git push`. The user performs all git operations themselves.

## What this is

`line.omise.app`, はんなりdev's LINE mini-app platform: three LINE mini-apps (membership management, salon reservation, sweets shop) authenticated via LINE Login and Google OAuth.

**Pivot in progress**: the repo currently holds only a static marketing landing page (see Architecture below — React + Vite, no auth, no backend). It is transitioning into the real application: Next.js frontend + Go Gin backend, with LINE Login / Google OAuth authentication. See `README.md` for the target architecture. That target is *not yet implemented* — everything in the Architecture section below still describes what actually exists in this repo today. New work toward the pivot goes through the dev-plan flow above, same as any other change.

The demo apps' current (pre-pivot) sibling repos, referenced by the existing landing page content:

- `line-apps-demo/membership`
- `line-apps-demo/salon-reservation`
- `line-apps-demo/sweets-shop`

## Repo layout

The repo is structured for a future multi-part system, but only the frontend exists today:

```
sys/
  01_infra/       (empty — reserved)
  02_backend/     (empty — reserved)
  03_frontend/web (the actual app — everything below lives here)
  04_e2e/         (empty — reserved)
  secrets/        (empty — gitignored, local secrets only)
```

All commands below are run from `sys/03_frontend/web`.

## Commands

```bash
npm install       # install deps
npm run dev       # start dev server on http://localhost:3000
npm run build     # tsc -b (type-check via project references) then vite build -> dist/
npm run preview   # serve the built dist/ locally
```

There is no lint script and no test suite configured in this project.

## Architecture

- **Stack**: React 19 + TypeScript + Vite 7, styled with Tailwind CSS v4 (via `@tailwindcss/vite`, configured in `src/index.css` using `@theme`, not a `tailwind.config.js`).
- **Path alias**: `@/*` → `src/*` (set in both `vite.config.ts` and `tsconfig.app.json` — keep these two in sync if the alias changes).
- **TypeScript project references**: `tsconfig.json` has no compiler options itself; it references `tsconfig.app.json` (the `src/` app code) and `tsconfig.node.json` (Vite config). `npm run build` type-checks via `tsc -b` before `vite build`.
- **Content is data-driven, not hardcoded in JSX**: page copy lives in `src/data/siteContent.ts` and the three demo-app cards' content lives in `src/data/demoApps.ts` (name, description, accent color, features, tech badges, screenshot/movie paths). Adding or editing a demo app is a data change in `demoApps.ts`, not a new component. Adding/editing page copy is a change in `siteContent.ts`.
- **Page composition**: `src/App.tsx` is a flat list of section components (`Header`, `HeroSection`, `DemoAppsSection`, `FeaturesSection`, `WhyUsSection`, `ContactSection`, `ProfileSection`, `Footer`), each in `src/components/`, each reading from the two data files above rather than taking props from `App.tsx`.
- **Demo card media behavior**: `DemoAppCard.tsx` shows a static screenshot by default and swaps to an autoplaying muted `<video>` on hover (see the ref-based play/pause in that component) — follow this pattern for any new hover-media behavior rather than introducing a new approach.
- **LINE OA link**: the LINE official account URL used by CTAs is a single value, `siteContent.lineOaUrl` — update it there, not per-component.
- **SEO**: metadata (title, description, OGP, Twitter Card, JSON-LD, GA4 tag) is hardcoded directly in `index.html`, not generated from `siteContent.ts` — when SEO copy changes, both `index.html` and `siteContent.ts` need updating for consistency. `public/robots.txt` and `public/sitemap.xml` are static files. See `doc/seo-update-2026-02-16.md` for the SEO history/rationale and its verification checklist (build, check `dist/index.html`/`robots.txt`/`sitemap.xml`, validate structured data and OGP) before shipping further SEO changes.
