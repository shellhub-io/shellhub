package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migrator orchestrates the migration of data from MongoDB to PostgreSQL.
type Migrator struct {
	mongo *mongo.Database
	pg    *bun.DB
}

// New creates a new Migrator instance.
func New(mongo *mongo.Database, pg *bun.DB) *Migrator {
	return &Migrator{mongo: mongo, pg: pg}
}

// tableFunc represents a function that migrates a single table.
type tableFunc struct {
	name string
	fn   func(ctx context.Context) error
}

// Run executes the full migration pipeline: migrate all tables in FK order, then validate.
func (m *Migrator) Run(ctx context.Context) error {
	if err := initStateTable(ctx, m.pg); err != nil {
		return fmt.Errorf("failed to initialize migration state table: %w", err)
	}

	// Migration phases ordered by FK dependencies.
	tables := []tableFunc{
		// Phase 1: no dependencies
		{"systems", m.migrateSystems},
		// Phase 2
		{"namespaces", m.migrateNamespaces},
		// Phase 3: depends on namespaces
		{"users", m.migrateUsers},
		{"tags", m.migrateTags},
		// Phase 4: depends on users + namespaces
		{"memberships", m.migrateMemberships},
		{"api_keys", m.migrateAPIKeys},
		{"public_keys", m.migratePublicKeys},
		{"devices", m.migrateDevices},
		// Phase 5: depends on phase 4
		{"device_tags", m.migrateDeviceTags},
		{"public_key_tags", m.migratePublicKeyTags},
		{"sessions", m.migrateSessions},
		// Phase 6: depends on sessions
		{"session_events", m.migrateSessionEvents},
	}

	for _, t := range tables {
		if err := m.migrateTable(ctx, t); err != nil {
			return fmt.Errorf("migration of %s failed: %w", t.name, err)
		}
	}

	log.Info("All tables migrated, running validation")

	if err := m.validate(ctx); err != nil {
		return fmt.Errorf("post-migration validation failed: %w", err)
	}

	return nil
}

// migrateTable handles the state machine for a single table: skip if completed, truncate+retry if
// previously in_progress, or run fresh if pending.
func (m *Migrator) migrateTable(ctx context.Context, t tableFunc) error {
	state, err := getState(ctx, m.pg, t.name)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if state != nil && state.Status == statusCompleted {
		log.WithField("table", t.name).Info("Already migrated, skipping")

		return nil
	}

	if state != nil && state.Status == statusInProgress {
		log.WithField("table", t.name).Warn("Previous migration was interrupted, truncating and retrying")

		if _, err := m.pg.NewTruncateTable().TableExpr(t.name).Cascade().Exec(ctx); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", t.name, err)
		}
	}

	if err := setState(ctx, m.pg, t.name, statusInProgress); err != nil {
		return err
	}

	log.WithField("table", t.name).Info("Starting migration")

	if err := t.fn(ctx); err != nil {
		return err
	}

	if err := setState(ctx, m.pg, t.name, statusCompleted); err != nil {
		return err
	}

	log.WithField("table", t.name).Info("Migration completed")

	return nil
}
