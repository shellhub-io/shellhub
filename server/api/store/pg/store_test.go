package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
)

// TestPgStore runs all store tests against PostgreSQL
// Each sub-suite gets a fresh database with migrations to prevent test pollution
func TestPgStore(t *testing.T) {
	for _, group := range storetest.Groups {
		runSubSuite(t, group.Name, func(t *testing.T, suite *storetest.Suite) {
			t.Helper()

			for _, test := range group.Tests {
				test(suite, t)
			}
		})
	}
}

func runSubSuite(t *testing.T, name string, testFunc func(*testing.T, *storetest.Suite)) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		ctx := context.Background()
		provider, err := pgprovider.NewProvider(ctx)
		if err != nil {
			t.Fatalf("Failed to create PostgreSQL provider for %s: %v", name, err)
		}
		defer provider.Close(t) //nolint:errcheck

		suite := storetest.NewSuite(provider)
		testFunc(t, suite)
	})
}
