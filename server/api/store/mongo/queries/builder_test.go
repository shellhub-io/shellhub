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

func TestFromFilters(t *testing.T) {
	type Expected struct {
		data []bson.M
		err  error
	}
	cases := []struct {
		description string
		filters     *query.Filters
		expected    Expected
	}{
		{
			description: "Fail when filter type is not valid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "invalid",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "valid",
							Value:    "test",
						},
					},
				},
			},
			expected: Expected{nil, query.ErrFilterInvalid},
		},
		{
			description: "Fail when operator in property is invalid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "invalid",
							Value:    "valid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{},
				err:  nil,
			},
		},
		{
			description: "Success when one operator in property is valid and other is invalid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "invalid",
							Value:    "test",
						},
					},
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "eq",
							Value:    "valid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "valid"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Success when operator in property is valid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "eq",
							Value:    "valid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "valid"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Fail when operator in operator is invalid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "invalid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{},
				err:  nil,
			},
		},
		{
			description: "Fail when operator in operator is valid and other invalid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "and",
						},
					},
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "invalid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M{}}}},
				err:  nil,
			},
		},
		{
			description: "Success when operator in operator is valid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "and",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M{}}}},
				err:  nil,
			},
		},
		{
			description: "Fail when property operator is invalid and operator is valid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "invalid",
							Value:    "test",
						},
					},
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "and",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M{}}}},
				err:  nil,
			},
		},
		{
			description: "Fail when property operator is valid and operator is invalid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "eq",
							Value:    "test",
						},
					},
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "invalid",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "test"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Success when property and operator is valid",
			filters: &query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "test",
							Operator: "eq",
							Value:    "test",
						},
					},
					{
						Type: "operator",
						Params: &query.FilterOperator{
							Name: "and",
						},
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M{{"test": bson.M{"$eq": "test"}}}}}},
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			query, err := FromFilters(tc.filters)
			assert.Equal(t, tc.expected, Expected{query, err})
		})
	}
}
