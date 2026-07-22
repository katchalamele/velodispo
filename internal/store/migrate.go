package store

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(dsn string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("source migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL(dsn))
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("application migrations: %w", err)
	}
	return nil
}

// migrateURL adapte le DSN au driver golang-migrate, qui attend le schéma pgx5.
func migrateURL(dsn string) string {
	if after, ok := strings.CutPrefix(dsn, "postgres://"); ok {
		return "pgx5://" + after
	}
	if after, ok := strings.CutPrefix(dsn, "postgresql://"); ok {
		return "pgx5://" + after
	}
	return dsn
}
