package migrator

import "testing"

// TestNonTransactionalDetection covers the two things checkTransactionality has to get right: that
// prose merely naming one of these statements does not count, and that a statement sharing its
// chunk with another one does.
func TestNonTransactionalDetection(t *testing.T) {
	tests := []struct {
		name       string
		chunk      string
		keyword    string
		statements int
	}{
		{
			name:       "vacuum cannot run in a transaction",
			chunk:      "VACUUM (FULL, ANALYZE) devices;",
			keyword:    "VACUUM",
			statements: 1,
		},
		{
			name:       "prose naming a vacuum is not a statement",
			chunk:      "-- recover by running VACUUM (FULL, ANALYZE) devices; by hand\nSELECT 1;",
			keyword:    "",
			statements: 1,
		},
		{
			name:       "keyword split across lines is still found",
			chunk:      "CREATE INDEX CONCURRENTLY\n    devices_last_seen ON devices USING btree (last_seen);",
			keyword:    "CREATE INDEX CONCURRENTLY",
			statements: 1,
		},
		{
			name:       "a shared chunk is counted",
			chunk:      "SET lock_timeout = '60s';\nVACUUM (FULL, ANALYZE) devices;",
			keyword:    "VACUUM",
			statements: 2,
		},
		{
			name:       "ordinary ddl is transactional",
			chunk:      "DROP INDEX IF EXISTS devices_last_seen;",
			keyword:    "",
			statements: 1,
		},
		{
			name:       "a storage parameter merely spelling a keyword is not that statement",
			chunk:      "ALTER TABLE sessions SET (autovacuum_vacuum_scale_factor = 0.05);",
			keyword:    "",
			statements: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statement := stripSQLComments(tc.chunk)

			if keyword := findNonTransactionalStatement(statement); keyword != tc.keyword {
				t.Errorf("findNonTransactionalStatement() = %q, want %q", keyword, tc.keyword)
			}

			if count := countStatements(statement); count != tc.statements {
				t.Errorf("countStatements() = %d, want %d", count, tc.statements)
			}
		})
	}
}
