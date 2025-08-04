package queries

import (
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"go.mongodb.org/mongo-driver/bson"
)

// FromPaginator converts the Paginator instance to a BSON pagination expression for MongoDB queries.
// If the per-page count is less than 1, it returns nil.
func FromPaginator(p *query.Paginator) []bson.M {
	if p.PerPage < 1 {
		return nil
	}

	return []bson.M{
		{"$skip": p.PerPage * (p.Page - 1)},
		{"$limit": p.PerPage},
	}
}

// FromSorter converts the Sort instance to a BSON sorting expression for MongoDB queries.
// If an invalid value of `Sort.By` is provided, it defaults to ascending order (OrderAsc).
func FromSorter(s *query.Sorter) []bson.M {
	options := map[string]int{
		query.OrderAsc:  1,
		query.OrderDesc: -1,
	}

	order, ok := options[s.Order]
	if !ok {
		order = -1
	}

	return []bson.M{
		{
			"$sort": bson.M{
				s.By: order,
			},
		},
	}
}
