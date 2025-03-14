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

func fetchMigrationsPath() (string, error) {
	migrationsPath := filepath.Join("/", "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.WithField("path", migrationsPath).Info("Migrations directory not found, creating it")
		if err := os.MkdirAll(migrationsPath, 0755); err != nil {
			return "", errors.New("failed to create migrations directory: " + err.Error())
		}
	} else if err != nil {
		return "", err
	}

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.WithError(err).WithField("path", migrationsPath).Error("failed to read migrations directory")
		return "", err
	}

	if len(files) == 0 {
		log.Error("no migration files found in directory")
	} else {
		log.WithField("path", migrationsPath).WithField("total_files", len(files)).Info("migration files directory content:")

		for i, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				log.WithField("filename", file.Name()).WithError(err).Warn("failed to get file info")
				continue
			}

			log.WithFields(log.Fields{
				"index":  i,
				"name":   file.Name(),
				"size":   fileInfo.Size(),
				"is_dir": file.IsDir(),
				"mode":   fileInfo.Mode().String(),
			}).Info("migration file")
		}
	}

	return migrationsPath, nil
}
