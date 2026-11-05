package handlers

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/binding/bindingtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryStructNamingASourceNamesOneForEachField(t *testing.T) {
	problems, err := bindingtest.CheckTree("../../../../..")
	require.NoError(t, err)
	assert.Empty(t, problems)
}
