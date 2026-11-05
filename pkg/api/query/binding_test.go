package query_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/binding/bindingtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryFieldNamesItsSource(t *testing.T) {
	problems, err := bindingtest.CheckDir(".")
	require.NoError(t, err)
	assert.Empty(t, problems)
}
