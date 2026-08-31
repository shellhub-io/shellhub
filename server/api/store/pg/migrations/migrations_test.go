package migrations

import (
	"io/fs"
	"strings"
	"testing"
)

// TestNoDuplicateMigrationVersions guards against two migrations sharing the same
// numeric prefix. bun derives a migration's name from the text before the first
// underscore, so a duplicate number makes it silently apply one and skip the
// other (it records the name as applied and never runs the second file).
func TestNoDuplicateMigrationVersions(t *testing.T) {
	files, err := fs.Glob(sqlMigrations, "*.up.sql")
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}

	seen := make(map[string]string, len(files))
	for _, file := range files {
		version := strings.SplitN(file, "_", 2)[0]
		if prev, ok := seen[version]; ok {
			t.Errorf("duplicate migration version %q: %q and %q must have distinct numbers", version, prev, file)
		}

		seen[version] = file
	}
}

var nonTransactionalStatements = []string{
	"VACUUM",
	"ALTER SYSTEM",
	"CREATE INDEX CONCURRENTLY",
	"DROP INDEX CONCURRENTLY",
	"REINDEX INDEX CONCURRENTLY",
	"REINDEX TABLE CONCURRENTLY",
	"CREATE DATABASE",
	"DROP DATABASE",
}

// TestNonTransactionalMigrations guards the two conditions such a statement needs, neither of
// which is visible in the SQL itself: bun decides transactionality from the ".tx." filename
// suffix, and the pool runs in pgx simple-protocol mode where a multi-statement Exec is itself
// an implicit transaction block. Getting either wrong fails at boot, not in review.
func TestNonTransactionalMigrations(t *testing.T) {
	files, err := fs.Glob(sqlMigrations, "*.sql")
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}

	for _, file := range files {
		raw, err := fs.ReadFile(sqlMigrations, file)
		if err != nil {
			t.Fatalf("failed to read %q: %v", file, err)
		}

		for chunk := range strings.SplitSeq(string(raw), "--bun:split") {
			statement := stripSQLComments(chunk)

			keyword := findNonTransactionalStatement(statement)
			if keyword == "" {
				continue
			}

			if strings.Contains(file, ".tx.") {
				t.Errorf("%s: %q cannot run inside a transaction, so the file must not carry the .tx. suffix", file, keyword)
			}

			if n := countStatements(statement); n > 1 {
				t.Errorf("%s: %q shares its --bun:split chunk with %d other statement(s), which simple-protocol mode batches into an implicit transaction", file, keyword, n-1)
			}
		}
	}
}

func stripSQLComments(chunk string) string {
	lines := strings.Split(chunk, "\n")
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

func findNonTransactionalStatement(chunk string) string {
	for statement := range strings.SplitSeq(chunk, ";") {
		normalized := strings.Join(strings.Fields(strings.ToUpper(statement)), " ")

		for _, keyword := range nonTransactionalStatements {
			if strings.HasPrefix(normalized, keyword) {
				return keyword
			}
		}
	}

	return ""
}

// TestNonTransactionalDetection covers the two things the guard above has to get right: that
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

func countStatements(statement string) int {
	count := 0

	for s := range strings.SplitSeq(statement, ";") {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}

	return count
}
