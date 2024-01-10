package queries

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestFromPaginator(t *testing.T) {
	cases := []struct {
		description string
		paginator   *query.Paginator
		expected    []bson.M
	}{
		{
			description: "succeeds with nil when PerPage is 0",
			paginator:   &query.Paginator{Page: 1, PerPage: 0},
			expected:    nil,
		},
		{
			description: "skips 0 documents when Page is 1",
			paginator:   &query.Paginator{Page: 1, PerPage: 10},
			expected: []bson.M{
				{"$skip": 0},
				{"$limit": 10},
			},
		},
		{
			description: "skips N documents when Page is > 1",
			paginator:   &query.Paginator{Page: 3, PerPage: 100},
			expected: []bson.M{
				{"$skip": 200},
				{"$limit": 100},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, FromPaginator(tc.paginator))
		})
	}
}
