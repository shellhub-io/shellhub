package query_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listRequest is shaped like the request types the API binds: it embeds both a paginator and a
// sorter, which is exactly the case that makes a promoted Normalize method ambiguous.
type listRequest struct {
	query.Paginator
	query.Sorter
}

func TestEmbeddedAccessorsReachBothValues(t *testing.T) {
	req := &listRequest{}

	paginated, ok := any(req).(query.Paginated)
	require.True(t, ok, "a request embedding query.Paginator must satisfy query.Paginated")

	sorted, ok := any(req).(query.Sorted)
	require.True(t, ok, "a request embedding query.Sorter must satisfy query.Sorted")

	paginated.GetPaginator().Normalize()
	sorted.GetSorter().Normalize()

	assert.Equal(t, query.MinPage, req.Paginator.Page)
	assert.Equal(t, query.DefaultPerPage, req.Paginator.PerPage)
	assert.Equal(t, query.OrderDesc, req.Sorter.Order)
}

// TestAccessorsAreAbsentWhenNotEmbedded pins the other half: the wrapper skips normalization for a
// request that carries neither, instead of normalizing something it invented.
func TestAccessorsAreAbsentWhenNotEmbedded(t *testing.T) {
	req := &struct{ UID string }{}

	_, paginated := any(req).(query.Paginated)
	_, sorted := any(req).(query.Sorted)

	assert.False(t, paginated)
	assert.False(t, sorted)
}
