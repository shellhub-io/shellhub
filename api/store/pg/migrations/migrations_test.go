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
