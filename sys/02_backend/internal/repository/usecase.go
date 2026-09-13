package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Usecase mirrors the `usecases` table (dev-plan-02-database 2.3,
// thumbnail_url/related_demo_app added by dev-plan-10-frontend-usecase
// migration 006).
type Usecase struct {
	ID             int64
	Slug           string
	ClientName     string
	Title          string
	Body           string
	Status         string
	ThumbnailURL   string
	RelatedDemoApp string
	PublishedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const usecasePublished = "published"

// UsecaseRepository queries the `usecases` table.
type UsecaseRepository struct {
	db *pgxpool.Pool
}

func NewUsecaseRepository(db *pgxpool.Pool) *UsecaseRepository {
	return &UsecaseRepository{db: db}
}

// ListAll returns every usecase regardless of status, for the admin UI
// (dev-plan-11-frontend-admin) which needs to see and manage drafts too.
func (r *UsecaseRepository) ListAll(ctx context.Context) ([]Usecase, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
		FROM usecases
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usecases []Usecase
	for rows.Next() {
		u, err := scanUsecase(rows)
		if err != nil {
			return nil, err
		}
		usecases = append(usecases, *u)
	}
	return usecases, rows.Err()
}

func (r *UsecaseRepository) ListPublished(ctx context.Context) ([]Usecase, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
		FROM usecases WHERE status = $1
		ORDER BY published_at DESC NULLS LAST, id DESC
	`, usecasePublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usecases []Usecase
	for rows.Next() {
		u, err := scanUsecase(rows)
		if err != nil {
			return nil, err
		}
		usecases = append(usecases, *u)
	}
	return usecases, rows.Err()
}

func (r *UsecaseRepository) GetPublishedBySlug(ctx context.Context, slug string) (*Usecase, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
		FROM usecases WHERE slug = $1 AND status = $2
	`, slug, usecasePublished)
	return scanUsecase(row)
}

func (r *UsecaseRepository) GetByID(ctx context.Context, id int64) (*Usecase, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
		FROM usecases WHERE id = $1
	`, id)
	return scanUsecase(row)
}

func (r *UsecaseRepository) Create(ctx context.Context, slug, clientName, title, body, status, thumbnailURL, relatedDemoApp string, publishedAt *time.Time) (*Usecase, error) {
	publishedAt = defaultPublishedAt(status, publishedAt)
	row := r.db.QueryRow(ctx, `
		INSERT INTO usecases (slug, client_name, title, body, status, thumbnail_url, related_demo_app, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
	`, slug, clientName, title, body, status, nullIfEmpty(thumbnailURL), nullIfEmpty(relatedDemoApp), publishedAt)
	return scanUsecase(row)
}

func (r *UsecaseRepository) Update(ctx context.Context, id int64, slug, clientName, title, body, status, thumbnailURL, relatedDemoApp string, publishedAt *time.Time) (*Usecase, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if publishedAt == nil {
		publishedAt = existing.PublishedAt
	}
	publishedAt = defaultPublishedAt(status, publishedAt)

	row := r.db.QueryRow(ctx, `
		UPDATE usecases SET slug = $2, client_name = $3, title = $4, body = $5, status = $6,
			thumbnail_url = $7, related_demo_app = $8, published_at = $9, updated_at = now()
		WHERE id = $1
		RETURNING id, slug, client_name, title, COALESCE(body, ''), status,
			COALESCE(thumbnail_url, ''), COALESCE(related_demo_app, ''), published_at, created_at, updated_at
	`, id, slug, clientName, title, body, status, nullIfEmpty(thumbnailURL), nullIfEmpty(relatedDemoApp), publishedAt)
	return scanUsecase(row)
}

func (r *UsecaseRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM usecases WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// nullIfEmpty converts an empty string to SQL NULL so optional text columns
// (and the related_demo_app CHECK, which would otherwise reject "") store
// "unset" consistently.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanUsecase(row rowScanner) (*Usecase, error) {
	var u Usecase
	err := row.Scan(&u.ID, &u.Slug, &u.ClientName, &u.Title, &u.Body, &u.Status,
		&u.ThumbnailURL, &u.RelatedDemoApp, &u.PublishedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &u, nil
}
