// Package repository holds the query layer over the schema from
// dev-plan-02-database. Content/exam/progress repositories are added in
// dev-plan-05/06-*; this file only covers what dev-plan-04-auth needs.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUserNotFound is returned when a requested user no longer exists.
var ErrUserNotFound = errors.New("user not found")

// User mirrors the `users` table (dev-plan-02-database 2.2).
type User struct {
	ID             int64
	Provider       string
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
	Role           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) IsAdmin() bool { return u.Role == "admin" }

// UserRepository queries the `users` table.
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// UpsertByProvider registers a user on first login, or refreshes their
// profile fields on subsequent logins. isAdmin (from config.AdminEmails)
// promotes reader -> admin on login; it never demotes an existing admin,
// so removing an address from ADMIN_EMAILS doesn't silently revoke access
// that may have been granted by other means (dev-plan-04-auth 4.5).
func (r *UserRepository) UpsertByProvider(ctx context.Context, provider, providerUserID, email, displayName, avatarURL string, isAdmin bool) (*User, error) {
	role := "reader"
	if isAdmin {
		role = "admin"
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO users (provider, provider_user_id, email, display_name, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider, provider_user_id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			role = CASE WHEN $6 = 'admin' THEN 'admin' ELSE users.role END,
			updated_at = now()
		RETURNING id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), role, created_at, updated_at
	`, provider, providerUserID, email, displayName, avatarURL, role)

	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), role, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

// DeleteByID permanently removes a user. Database foreign keys cascade the
// deletion to every session and current user-owned quiz record.
func (r *UserRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Provider, &u.ProviderUserID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
