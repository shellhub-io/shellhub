package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionAutovacuumThresholds covers migration 022. The knobs are per-table storage
// parameters, so the only honest assertion is what pg_class.reloptions holds on a migrated
// database — reading the migration file back would assert nothing but its own text.
func TestSessionAutovacuumThresholds(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { provider.Close(t) })

	for _, table := range []string{"sessions", "session_events"} {
		t.Run(table, func(t *testing.T) {
			// unnest rather than reading the array whole: it yields one row per option, so
			// the assertion matches a setting exactly instead of finding it as a substring
			// of another. A table with no storage parameters yields no rows.
			var options []string
			require.NoError(t,
				provider.DB().NewSelect().
					ColumnExpr("unnest(reloptions)").
					TableExpr("pg_class").
					Where("relname = ?", table).
					Where("relkind = 'r'").
					Scan(ctx, &options),
			)

			assert.Contains(t, options, "autovacuum_vacuum_scale_factor=0.05")
			assert.Contains(t, options, "autovacuum_analyze_scale_factor=0.02")

			// The new analyze factor only decides when the daemon next looks; it does nothing
			// about statistics that are already stale, which on the instance this came from
			// were 45 days old. The migration refreshes them once itself.
			var analyzed int
			require.NoError(t,
				provider.DB().NewSelect().
					ColumnExpr("count(*)").
					TableExpr("pg_stat_user_tables").
					Where("relname = ?", table).
					Where("last_analyze IS NOT NULL").
					Scan(ctx, &analyzed),
			)

			assert.Equal(t, 1, analyzed, "migration must leave %s with fresh statistics", table)
		})
	}
}
