package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned by Get*/scan helpers when no matching row exists,
// so callers (handlers) only need to check one sentinel instead of importing
// pgx themselves.
var ErrNotFound = errors.New("not found")

func wrapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
