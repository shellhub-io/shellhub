package migrate

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrationExtension is a function that performs additional migration steps
// after the core migration completes. Cloud/enterprise packages register
// extensions to migrate their own tables and extra columns.
type MigrationExtension func(ctx context.Context, mongo *mongo.Database, pg *bun.DB) error

// migrationExtensions holds all registered migration extensions.
var migrationExtensions []MigrationExtension

// RegisterMigrationExtension registers a migration extension. Must be called
// before Migrator.Run() — typically from a cloud package's init() function.
func RegisterMigrationExtension(ext MigrationExtension) {
	migrationExtensions = append(migrationExtensions, ext)
}

// applyMigrationExtensions invokes all registered migration extensions.
func applyMigrationExtensions(ctx context.Context, mongo *mongo.Database, pg *bun.DB) error {
	for _, ext := range migrationExtensions {
		if err := ext(ctx, mongo, pg); err != nil {
			log.WithError(err).Error("failed to apply migration extension")

			return err
		}
	}

	return nil
}
