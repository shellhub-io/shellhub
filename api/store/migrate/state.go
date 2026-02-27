package migrate

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// MigrationState tracks the progress of each table's migration.
type MigrationState struct {
	bun.BaseModel `bun:"table:migration_state"`

	TableName   string     `bun:"table_name,pk"`
	Status      string     `bun:"status"`
	SourceCount int64      `bun:"source_count"`
	TargetCount int64      `bun:"target_count"`
	StartedAt   *time.Time `bun:"started_at"`
	CompletedAt *time.Time `bun:"completed_at"`
}

const (
	statusPending    = "pending"
	statusInProgress = "in_progress"
	statusCompleted  = "completed"
)

func initStateTable(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*MigrationState)(nil)).
		IfNotExists().
		Exec(ctx)

	return err
}

func getState(ctx context.Context, db *bun.DB, table string) (*MigrationState, error) {
	state := new(MigrationState)
	err := db.NewSelect().Model(state).Where("table_name = ?", table).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func setState(ctx context.Context, db *bun.DB, table, status string) error {
	state := &MigrationState{
		TableName: table,
		Status:    status,
	}

	now := time.Now()
	switch status {
	case statusInProgress:
		state.StartedAt = &now
	case statusCompleted:
		state.CompletedAt = &now
	}

	_, err := db.NewInsert().
		Model(state).
		On("CONFLICT (table_name) DO UPDATE").
		Set("status = EXCLUDED.status").
		Set("started_at = COALESCE(EXCLUDED.started_at, migration_state.started_at)").
		Set("completed_at = EXCLUDED.completed_at").
		Exec(ctx)

	return err
}

func setStateCounts(ctx context.Context, db *bun.DB, table string, source, target int64) error {
	_, err := db.NewUpdate().
		Model((*MigrationState)(nil)).
		Set("source_count = ?", source).
		Set("target_count = ?", target).
		Where("table_name = ?", table).
		Exec(ctx)

	return err
}
