// Package service holds logic shared across the content repositories/handlers
// from dev-plan-05-content-api: status/category validation and turning a
// Postgres unique-constraint violation (duplicate slug) into a clean error
// the handler can map to 409, instead of a raw driver error leaking to 500.
package service

import (
	"errors"
)

var (
	ErrInvalidStatus         = errors.New("status must be draft or published")
	ErrInvalidCategory       = errors.New("category must be line-operation or ai")
	ErrInvalidRelatedDemoApp = errors.New("related_demo_app must be membership, salon-reservation, sweets-shop, or empty")
	ErrDuplicateSlug         = errors.New("slug already in use")
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

// ValidateRelatedDemoApp checks a usecase's related_demo_app against the
// schema's CHECK constraint (dev-plan-10-frontend-usecase migration 006).
// Empty is valid — the association is optional.
func ValidateRelatedDemoApp(relatedDemoApp string) error {
	switch relatedDemoApp {
	case "", "membership", "salon-reservation", "sweets-shop":
		return nil
	default:
		return ErrInvalidRelatedDemoApp
	}
}

// AsDuplicateSlug returns ErrDuplicateSlug if err is a unique-constraint
// violation, and the original err otherwise (including nil).
func AsDuplicateSlug(err error) error {
	if isUniqueViolation(err, "") {
		return ErrDuplicateSlug
	}
	return err
}
