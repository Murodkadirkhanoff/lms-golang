package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Constraint returns the Postgres constraint name for a unique/FK/check
// violation, or "" if err is not a constraint error. Equivalent to inspecting
// pq.Error.Constraint in the original Go services.
func Constraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}
