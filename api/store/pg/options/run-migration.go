package options

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

var (
	ErrMigrationFail          = errors.New("failed to apply migration")
	ErrMigrationsPathNotFound = errors.New("migrations path not found")
)

func RunMigrations() Option {
	return func(ctx context.Context, db *sql.DB) error {
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return err
		}

		migrationsPath, err := fetchMigrationsPath()
		if err != nil {
			return ErrMigrationsPathNotFound
		}

		m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, envs.DefaultBackend.Get("POSTGRES_DB"), driver)
		if err != nil {
			return errors.Join(ErrMigrationFail, err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return errors.Join(ErrMigrationFail, err)
		}

		return nil
	}
}

func fetchMigrationsPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	migrationsPath := filepath.Join(cwd, "store", "pg", "migrations")
	if _, err := os.Stat(migrationsPath); err != nil {
		return "", err
	}

	return migrationsPath, nil
}
