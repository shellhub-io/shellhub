package options

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

var (
	ErrMigrationFail = errors.New("failed to apply migration")
	ErrFilePathEmpty = errors.New("no migration files found in directory")
)

func Migrate(basePath string) Option {
	return func(ctx context.Context, db *bun.DB) error {
		driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return err
		}

		migrationsPath, err := fetchMigrationsPath(basePath)
		if err != nil {
			return err
		}

		m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, envs.DefaultBackend.Get("POSTGRES_DB"), driver)
		if err != nil {
			log.WithError(err).Error("failed to create migrate instance")

			return errors.Join(ErrMigrationFail, err)
		}

		if version, dirty, _ := m.Version(); dirty {
			log.WithField("version", version).
				WithField("dirty", dirty).
				WithError(err).
				Error("migrations are dirty. manual fix required")

			return errors.Join(ErrMigrationFail, err)
		}

		log.WithField("path", migrationsPath).Info("applying migrations")

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.WithError(err).Error("failed to apply migrations")

			return errors.Join(ErrMigrationFail, err)
		}

		log.Info("migrations applied")

		return nil
	}
}

func fetchMigrationsPath(basePath string) (string, error) {
	path := filepath.Join(basePath, "migrations")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		log.WithError(err).WithField("path", path).Error("failed to read migrations directory")
		return "", err
	}

	if len(files) == 0 {
		return "", ErrFilePathEmpty
	}

	return path, nil
}
