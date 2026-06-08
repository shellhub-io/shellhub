package pg_test

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestMongoInfraDeleted verifies that the legacy Mongo store infrastructure has
// been fully removed from the repository. It checks that:
//   - api/store/mongo/ is gone
//   - api/store/migrate/ is gone
//   - api/store/storetest/mongoprovider/ is gone
//   - api/pkg/dbtest/ is gone
//
// These are deleted as one atomic change because they reference each other.
func TestMongoInfraDeleted(t *testing.T) {
	// Locate the repo root relative to this test file.
	// This file lives at api/store/pg/cleanup_test.go so we go up four levels.
	_, file, _, _ := runtime.Caller(0)
	// api/store/pg -> api/store -> api -> (repo root)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..")

	deleted := []string{
		filepath.Join(repoRoot, "api", "store", "mongo"),
		filepath.Join(repoRoot, "api", "store", "migrate"),
		filepath.Join(repoRoot, "api", "store", "storetest", "mongoprovider"),
		filepath.Join(repoRoot, "api", "pkg", "dbtest"),
	}

	for _, dir := range deleted {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				t.Errorf("expected directory %s to be deleted, but it still exists (err=%v)", dir, err)
			}
		})
	}
}

// TestMongoDepsPruned verifies that go mod tidy has been run and that all
// MongoDB-related dependencies have been removed from api/go.mod and api/go.sum.
//
// The expected state after pruning:
//   - go.mongodb.org/mongo-driver    — removed (no mongo store)
//   - github.com/shellhub-io/mongotest — removed
//   - github.com/square/mongo-lock   — removed
//   - github.com/xakep666/mongo-migrate — removed
//   - testcontainers-go/modules/mongodb — removed
//
// testcontainers-go core and modules/postgres must remain (used by pg/dbtest).
func TestMongoDepsPruned(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	// api/store/pg -> api/store -> api
	apiRoot := filepath.Join(filepath.Dir(file), "..", "..")

	gomod := filepath.Join(apiRoot, "go.mod")
	gosum := filepath.Join(apiRoot, "go.sum")

	// Patterns that must NOT appear anywhere in go.mod or go.sum.
	banned := []string{
		"go.mongodb.org/mongo-driver",
		"github.com/shellhub-io/mongotest",
		"github.com/square/mongo-lock",
		"github.com/xakep666/mongo-migrate",
		"testcontainers-go/modules/mongodb",
	}

	for _, path := range []string{gomod, gosum} {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(path) //nolint:gosec // path is constructed from runtime.Caller, not user input.
			if err != nil {
				t.Fatalf("open %s: %v", path, err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			lineno := 0
			for scanner.Scan() {
				lineno++
				line := scanner.Text()
				for _, b := range banned {
					if strings.Contains(line, b) {
						t.Errorf("%s:%d: banned mongo dependency %q still present: %s", name, lineno, b, strings.TrimSpace(line))
					}
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("scan %s: %v", path, err)
			}
		})
	}

	// Positive assertion: modules/postgres must still be present in go.mod.
	t.Run("postgres_kept", func(t *testing.T) {
		f, err := os.Open(gomod) //nolint:gosec // path is constructed from runtime.Caller, not user input.
		if err != nil {
			t.Fatalf("open go.mod: %v", err)
		}
		defer f.Close()

		found := false
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "testcontainers-go/modules/postgres") {
				found = true

				break
			}
		}
		if !found {
			t.Error("testcontainers-go/modules/postgres must remain in go.mod (used by pg/dbtest)")
		}
	})
}
