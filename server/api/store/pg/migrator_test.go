package pg_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrator"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

func testTables(name string) migrator.Tables {
	return migrator.Tables{Migrations: name + "_migrations", Locks: name + "_migration_locks"}
}

func loadTestMigrations(t *testing.T, files fstest.MapFS, goMigrations ...migrator.Migration) []migrator.Migration {
	t.Helper()

	migrations, err := migrator.Load(files, goMigrations...)
	require.NoError(t, err)

	return migrations
}

func tableExists(t *testing.T, ctx context.Context, db *bun.DB, table string) bool {
	t.Helper()

	var exists bool
	require.NoError(t, db.QueryRowContext(ctx, "SELECT to_regclass(?) IS NOT NULL", table).Scan(&exists))

	return exists
}

func recordedMigrations(t *testing.T, ctx context.Context, db *bun.DB, table string) []string {
	t.Helper()

	var names []string
	require.NoError(t, db.NewSelect().TableExpr(table).Column("name").Order("name").Scan(ctx, &names))

	return names
}

func TestMigratorApply(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	t.Run("a transactional migration whose record fails leaves no trace", func(t *testing.T) {
		tables := testTables("atomic")

		_, err := migrator.Apply(ctx, db, tables, nil)
		require.NoError(t, err)

		execSQL(t, ctx, db, `CREATE FUNCTION refuse_record() RETURNS trigger LANGUAGE plpgsql AS $$
			BEGIN RAISE EXCEPTION 'record refused'; END $$`)
		execSQL(t, ctx, db, `CREATE TRIGGER refuse_record BEFORE INSERT ON atomic_migrations
			FOR EACH ROW EXECUTE FUNCTION refuse_record()`)

		migrations := loadTestMigrations(t, fstest.MapFS{
			"001_create_probe.tx.up.sql": {Data: []byte("CREATE TABLE atomic_probe (id int);")},
		})

		_, err = migrator.Apply(ctx, db, tables, migrations)
		require.ErrorContains(t, err, "record refused")

		assert.False(t, tableExists(t, ctx, db, "atomic_probe"))
		assert.Empty(t, recordedMigrations(t, ctx, db, tables.Migrations))
	})

	t.Run("a failing migration is not recorded and runs on the next apply", func(t *testing.T) {
		tables := testTables("retry")

		execSQL(t, ctx, db, "CREATE TABLE retry_probe (planted int)")

		migrations := loadTestMigrations(t, fstest.MapFS{
			"001_first.tx.up.sql":        {Data: []byte("CREATE TABLE retry_first (id int);")},
			"002_create_probe.tx.up.sql": {Data: []byte("CREATE TABLE retry_probe (id int);\n--bun:split\nINSERT INTO retry_probe VALUES (1);")},
		})

		ran, err := migrator.Apply(ctx, db, tables, migrations)
		require.ErrorContains(t, err, "002_create_probe")
		assert.Equal(t, 1, ran)
		assert.Equal(t, []string{"001"}, recordedMigrations(t, ctx, db, tables.Migrations))

		execSQL(t, ctx, db, "DROP TABLE retry_probe")

		ran, err = migrator.Apply(ctx, db, tables, migrations)
		require.NoError(t, err)
		assert.Equal(t, 1, ran)
		assert.Equal(t, []string{"001", "002"}, recordedMigrations(t, ctx, db, tables.Migrations))

		var id int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM retry_probe").Scan(&id))
		assert.Equal(t, 1, id)
	})

	t.Run("bun reads every kind of migration the runner applied as one group", func(t *testing.T) {
		tables := testTables("history")

		files := fstest.MapFS{
			"001_transactional.tx.up.sql":    {Data: []byte("CREATE TABLE history_tx (id int);")},
			"001_transactional.tx.down.sql":  {Data: []byte("DROP TABLE history_tx;")},
			"002_non_transactional.up.sql":   {Data: []byte("CREATE TABLE history_conn (id int);\n--bun:split\nCREATE INDEX CONCURRENTLY history_conn_id ON history_conn (id);")},
			"002_non_transactional.down.sql": {Data: []byte("DROP TABLE history_conn;")},
		}

		migrations := loadTestMigrations(t, files, migrator.Go("003", "go", func(ctx context.Context, db *bun.DB) error {
			_, err := db.ExecContext(ctx, "CREATE TABLE history_go (id int)")

			return err
		}))

		ran, err := migrator.Apply(ctx, db, tables, migrations)
		require.NoError(t, err)
		assert.Equal(t, 3, ran)

		registry := migrate.NewMigrations()
		require.NoError(t, registry.Discover(files))
		registry.Add(migrate.Migration{Name: "003", Comment: "go"})

		status, err := migrate.NewMigrator(db, registry, migrate.WithTableName(tables.Migrations)).MigrationsWithStatus(ctx)
		require.NoError(t, err)
		require.Len(t, status, 3)
		assert.Empty(t, status.Unapplied())
		assert.Len(t, status.LastGroup().Migrations, 3)

		ran, err = migrator.Apply(ctx, db, tables, migrations)
		require.NoError(t, err)
		assert.Zero(t, ran)
	})
}

func noop(context.Context, *bun.DB) error { return nil }

func TestMigratorLoad(t *testing.T) {
	tests := []struct {
		name         string
		files        fstest.MapFS
		goMigrations []migrator.Migration
		err          string
	}{
		{
			name:  "unknown directive",
			files: fstest.MapFS{"001_a.tx.up.sql": {Data: []byte("SELECT 1;\n--bun:nosplit\nSELECT 2;")}},
			err:   "unknown directive",
		},
		{
			name:  "file name bun would reject",
			files: fstest.MapFS{"first.tx.up.sql": {Data: []byte("SELECT 1;")}},
			err:   "unsupported migration file name",
		},
		{
			name:         "go migration sharing a file's name",
			files:        fstest.MapFS{"001_a.tx.up.sql": {Data: []byte("SELECT 1;")}},
			goMigrations: []migrator.Migration{migrator.Go("001", "b", noop)},
			err:          "share a name",
		},
		{
			name:         "go migration without a function",
			files:        fstest.MapFS{},
			goMigrations: []migrator.Migration{migrator.Go("001", "a", nil)},
			err:          "has no function",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := migrator.Load(tc.files, tc.goMigrations...)
			require.ErrorContains(t, err, tc.err)
		})
	}
}
