package query_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	assert.Same(t, &req.Paginator, paginated.GetPaginator())
	assert.Same(t, &req.Sorter, sorted.GetSorter())
}

func TestAccessorsAreAbsentWhenNotEmbedded(t *testing.T) {
	req := &struct{ UID string }{}

	_, paginated := any(req).(query.Paginated)
	_, sorted := any(req).(query.Sorted)

	assert.False(t, paginated)
	assert.False(t, sorted)
}
