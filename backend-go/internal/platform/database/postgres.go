// Package database is framework plumbing: it builds the pgx pool and runs a
// module's embedded migrations. It holds no bounded-context knowledge.
package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" scheme
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx connection pool and verifies it with a ping.
func Connect(ctx context.Context, dsn string, maxOpenConns int) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	if maxOpenConns > 0 {
		poolCfg.MaxConns = int32(maxOpenConns)
	}
	poolCfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// Migrate applies the migrations embedded in fsys (rooted at dir) against dsn.
// versionTable isolates each bounded context's migration history so modules
// version independently — the first step toward a database per module.
func Migrate(dsn string, fsys fs.FS, dir, versionTable string) error {
	src, err := iofs.New(fsys, dir)
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	// golang-migrate's pgx driver speaks database/sql; give it the sql-style DSN.
	m, err := migrate.NewWithSourceInstance("iofs", src, "pgx5://"+trimScheme(dsn)+migrationsTableParam(dsn, versionTable))
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// migrationsTableParam appends the per-module version table as a query param.
func migrationsTableParam(dsn, table string) string {
	sep := "?"
	if hasQuery(dsn) {
		sep = "&"
	}
	return sep + "x-migrations-table=" + table
}

func trimScheme(dsn string) string {
	for _, p := range []string{"postgres://", "postgresql://", "pgx5://", "pgx://"} {
		if len(dsn) >= len(p) && dsn[:len(p)] == p {
			return dsn[len(p):]
		}
	}
	return dsn
}

func hasQuery(dsn string) bool {
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '?' {
			return true
		}
	}
	return false
}
