package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Article mirrors the `articles` table (dev-plan-02-database 2.3).
type Article struct {
	ID          int64
	Category    string
	Slug        string
	Title       string
	Body        string
	Status      string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Tag mirrors the `tags` table.
type Tag struct {
	ID   int64
	Name string
	Slug string
}

const articlePublished = "published"

// ArticleRepository queries `articles`, `tags` and `article_tags`.
type ArticleRepository struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// ListAll returns every article regardless of status, optionally narrowed to
// one category, for the admin UI (dev-plan-11-frontend-admin) which needs to
// see and manage drafts too. Pass "" for category to list both.
func (r *ArticleRepository) ListAll(ctx context.Context, category string) ([]Article, error) {
	query := `SELECT id, category, slug, title, COALESCE(body, ''), status, published_at, created_at, updated_at FROM articles`
	var args []any
	if category != "" {
		query += ` WHERE category = $1`
		args = append(args, category)
	}
	query += ` ORDER BY id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *a)
	}
	return articles, rows.Err()
}

// ListPublished returns published articles for a category, optionally
// narrowed to those carrying tagSlug (dev-plan-05 5.2). Draft articles never
// appear here.
func (r *ArticleRepository) ListPublished(ctx context.Context, category, tagSlug string) ([]Article, error) {
	query := `
		SELECT DISTINCT a.id, a.category, a.slug, a.title, COALESCE(a.body, ''), a.status, a.published_at, a.created_at, a.updated_at
		FROM articles a`
	args := []any{category, articlePublished}
	if tagSlug != "" {
		query += `
		JOIN article_tags at ON at.article_id = a.id
		JOIN tags t ON t.id = at.tag_id AND t.slug = $3`
		args = append(args, tagSlug)
	}
	query += ` WHERE a.category = $1 AND a.status = $2 ORDER BY a.published_at DESC NULLS LAST, a.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *a)
	}
	return articles, rows.Err()
}

func (r *ArticleRepository) GetPublishedByCategoryAndSlug(ctx context.Context, category, slug string) (*Article, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, category, slug, title, COALESCE(body, ''), status, published_at, created_at, updated_at
		FROM articles WHERE category = $1 AND slug = $2 AND status = $3
	`, category, slug, articlePublished)
	return scanArticle(row)
}

func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (*Article, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, category, slug, title, COALESCE(body, ''), status, published_at, created_at, updated_at
		FROM articles WHERE id = $1
	`, id)
	return scanArticle(row)
}

// Create inserts an article. publishedAt is set to now() automatically when
// status is "published" and the caller didn't provide one.
func (r *ArticleRepository) Create(ctx context.Context, category, slug, title, body, status string, publishedAt *time.Time) (*Article, error) {
	publishedAt = defaultPublishedAt(status, publishedAt)
	row := r.db.QueryRow(ctx, `
		INSERT INTO articles (category, slug, title, body, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, category, slug, title, COALESCE(body, ''), status, published_at, created_at, updated_at
	`, category, slug, title, body, status, publishedAt)
	return scanArticle(row)
}

func (r *ArticleRepository) Update(ctx context.Context, id int64, category, slug, title, body, status string, publishedAt *time.Time) (*Article, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if publishedAt == nil {
		publishedAt = existing.PublishedAt
	}
	publishedAt = defaultPublishedAt(status, publishedAt)

	row := r.db.QueryRow(ctx, `
		UPDATE articles SET category = $2, slug = $3, title = $4, body = $5, status = $6, published_at = $7, updated_at = now()
		WHERE id = $1
		RETURNING id, category, slug, title, COALESCE(body, ''), status, published_at, created_at, updated_at
	`, id, category, slug, title, body, status, publishedAt)
	return scanArticle(row)
}

func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM articles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetOrCreateTag looks up a tag by slug, creating it if it doesn't exist yet.
func (r *ArticleRepository) GetOrCreateTag(ctx context.Context, name, slug string) (*Tag, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO tags (name, slug) VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, slug
	`, name, slug)
	var t Tag
	if err := row.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
		return nil, err
	}
	return &t, nil
}

// AttachTag links a tag to an article. Attaching the same tag twice is a
// no-op.
func (r *ArticleRepository) AttachTag(ctx context.Context, articleID, tagID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, articleID, tagID)
	return err
}

// ListTagsForArticle returns the tags attached to an article.
func (r *ArticleRepository) ListTagsForArticle(ctx context.Context, articleID int64) ([]Tag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.name, t.slug FROM tags t
		JOIN article_tags at ON at.tag_id = t.id
		WHERE at.article_id = $1
		ORDER BY t.slug
	`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func defaultPublishedAt(status string, publishedAt *time.Time) *time.Time {
	if status == articlePublished && publishedAt == nil {
		now := time.Now()
		return &now
	}
	return publishedAt
}

func scanArticle(row rowScanner) (*Article, error) {
	var a Article
	err := row.Scan(&a.ID, &a.Category, &a.Slug, &a.Title, &a.Body, &a.Status, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &a, nil
}
