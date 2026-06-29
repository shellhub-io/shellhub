package requests

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListSessionsFiltersEmbedding verifies that ListSessions embeds
// query.Filters such that its methods (Unmarshal, etc.) are accessible
// through the request struct and work end-to-end.
func TestListSessionsFiltersEmbedding(t *testing.T) {
	filters := []query.Filter{
		{
			Type: query.FilterTypeProperty,
			Params: &query.FilterProperty{
				Name:     "closed",
				Operator: "bool",
				Value:    true,
			},
		},
	}

	b, err := json.Marshal(filters)
	require.NoError(t, err)

	req := ListSessions{}
	req.Filters.Raw = base64.StdEncoding.EncodeToString(b)

	require.NoError(t, req.Filters.Unmarshal())
	require.Len(t, req.Filters.Data, 1)

	prop, ok := req.Filters.Data[0].Params.(*query.FilterProperty)
	require.True(t, ok)
	assert.Equal(t, "closed", prop.Name)
	assert.Equal(t, "bool", prop.Operator)
	assert.Equal(t, true, prop.Value)
}
