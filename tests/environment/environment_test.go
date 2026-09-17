package environment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseSwitchOnlyPostgres(t *testing.T) {
	require.Error(t, onlyPostgresAllowed("mongo"))
	require.NoError(t, onlyPostgresAllowed("postgres"))
}
