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

func TestFromSorter(t *testing.T) {
	cases := []struct {
		description string
		sorter      *query.Sorter
		expected    []bson.M
	}{
		{
			description: "sets sort to -1 when order.By is invalid",
			sorter:      &query.Sorter{By: "date", Order: "foo"},
			expected: []bson.M{
				{
					"$sort": bson.M{
						"date": -1,
					},
				},
			},
		},
		{
			description: "sets sort to 1 when order.By is asc",
			sorter:      &query.Sorter{By: "date", Order: "asc"},
			expected: []bson.M{
				{
					"$sort": bson.M{
						"date": 1,
					},
				},
			},
		},
		{
			description: "sets sort to -1 when order.By is desc",
			sorter:      &query.Sorter{By: "date", Order: "desc"},
			expected: []bson.M{
				{
					"$sort": bson.M{
						"date": -1,
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, FromSorter(tc.sorter))
		})
	}
}
