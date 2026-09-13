// Package service holds logic shared across the content repositories/handlers
// from dev-plan-05-content-api: status/category validation and turning a
// Postgres unique-constraint violation (duplicate slug) into a clean error
// the handler can map to 409, instead of a raw driver error leaking to 500.
package service

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidStatus   = errors.New("status must be draft or published")
	ErrInvalidCategory = errors.New("category must be line-operation or ai")
	ErrDuplicateSlug   = errors.New("slug already in use")
)

// ValidateStatus checks a content status against the schema's CHECK
// constraint (dev-plan-02-database) before it reaches the database.
func ValidateStatus(status string) error {
	if status != "draft" && status != "published" {
		return ErrInvalidStatus
	}
	return nil
}

// ValidateArticleCategory checks an article category against the schema's
// CHECK constraint.
func ValidateArticleCategory(category string) error {
	if category != "line-operation" && category != "ai" {
		return ErrInvalidCategory
	}
	return nil
}

// postgresUniqueViolation is the SQLSTATE code for a unique_violation.
const postgresUniqueViolation = "23505"

// AsDuplicateSlug returns ErrDuplicateSlug if err is a unique-constraint
// violation, and the original err otherwise (including nil).
func AsDuplicateSlug(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
		return fmt.Errorf("%w: %s", ErrDuplicateSlug, pgErr.ConstraintName)
	}
	return err
}
