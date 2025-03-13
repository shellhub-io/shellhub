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
	log "github.com/sirupsen/logrus"
)

var (
	ErrMigrationFail = errors.New("failed to apply migration")
)

func RunMigrations() Option {
	return func(ctx context.Context, db *sql.DB) error {
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return err
		}

		migrationsPath, err := fetchMigrationsPath()
		if err != nil {
			return err
		}

		log.WithField("path", migrationsPath).Info("Applying migrations")

		m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, envs.DefaultBackend.Get("POSTGRES_DB"), driver)
		if err != nil {
			return errors.Join(ErrMigrationFail, err)
		}

		if version, dirty, _ := m.Version(); dirty {
			log.WithField("version", version).
				WithField("dirty", dirty).
				Info("migrations are dirty. manual fix required")

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

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.WithField("path", migrationsPath).Info("Migrations directory not found, creating it")
		if err := os.MkdirAll(migrationsPath, 0755); err != nil {
			return "", errors.New("failed to create migrations directory: " + err.Error())
		}
	} else if err != nil {
		return "", err
	}

	return migrationsPath, nil
}
