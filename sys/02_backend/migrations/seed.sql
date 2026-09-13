-- Development-only sample data. Not run automatically by golang-migrate
-- (it is not a versioned migration) — apply manually via `make migrate-seed`.
-- Safe to re-run: each insert is guarded with ON CONFLICT DO NOTHING.

INSERT INTO users (provider, provider_user_id, email, display_name, role)
VALUES ('google', 'dev-admin', 'admin@example.com', 'Dev Admin', 'admin')
ON CONFLICT (provider, provider_user_id) DO NOTHING;

INSERT INTO courses (slug, title, description, sort_order, status)
VALUES ('line-marketing', 'LINEマーケティング講座', '基礎から学ぶLINEマーケティング', 1, 'published')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO lessons (course_id, slug, title, body, sort_order, status)
SELECT id, 'intro', 'はじめに', 'このレッスンではLINEマーケティングの全体像を学びます。', 1, 'published'
FROM courses WHERE slug = 'line-marketing'
ON CONFLICT (course_id, slug) DO NOTHING;

INSERT INTO articles (category, slug, title, body, status, published_at)
VALUES ('line-operation', 'rich-menu-basics', 'リッチメニューの基本設定', 'リッチメニューの設定手順を解説します。', 'published', now())
ON CONFLICT (category, slug) DO NOTHING;

INSERT INTO usecases (slug, client_name, title, body, status, published_at)
VALUES ('sample-salon', 'サンプルサロン', 'LINE公式アカウント導入事例', '導入により予約率が向上しました。', 'published', now())
ON CONFLICT (slug) DO NOTHING;
