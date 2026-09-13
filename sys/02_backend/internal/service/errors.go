package service

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// postgresUniqueViolation is the SQLSTATE code for a unique_violation.
const postgresUniqueViolation = "23505"

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation, optionally narrowed to a specific constraint name (pass "" to
// match any unique violation).
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != postgresUniqueViolation {
		return false
	}
	return constraintName == "" || pgErr.ConstraintName == constraintName
}
