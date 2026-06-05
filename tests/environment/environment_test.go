package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDatabaseSwitchOnlyPostgres verifies that the Up configurator only accepts
// "postgres" as a valid SHELLHUB_DATABASE value and panics on any other value,
// including the now-removed "mongo" option.
//
// This is a compile-time + runtime guard: onlyPostgresAllowed must be defined
// in the package (its presence confirms the mongo leg has been removed from
// configurator.go), and the function itself is called to confirm behaviour.
func TestDatabaseSwitchOnlyPostgres(t *testing.T) {
	assert.Panics(t, func() {
		onlyPostgresAllowed("mongo")
	}, "passing 'mongo' to onlyPostgresAllowed should panic")

	assert.NotPanics(t, func() {
		onlyPostgresAllowed("postgres")
	}, "passing 'postgres' to onlyPostgresAllowed should not panic")
}
