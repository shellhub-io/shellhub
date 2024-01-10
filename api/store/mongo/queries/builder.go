package queries

import (
	"github.com/shellhub-io/shellhub/api/store/mongo/queries/internal"
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

// FromFilters converts the Filters instance to a BSON filter expression for MongoDB queries.
// Returns an error when an invalid filter is found.
func FromFilters(fs *query.Filters) ([]bson.M, error) {
	if len(fs.Data) < 1 {
		return []bson.M{}, nil
	}

	queryFilter := make([]bson.M, 0)
	queryMatcher := make([]bson.M, 0)

	for _, filter := range fs.Data {
		switch filter.Type {
		case query.FilterTypeProperty:
			param, ok := filter.Params.(*query.FilterProperty)
			if !ok {
				return nil, query.ErrFilterInvalid
			}

			prop, ok, err := internal.ParseFilterProperty(param)
			if err != nil {
				return nil, query.ErrFilterPropertyInvalid
			}

			if !ok {
				continue
			}

			queryFilter = append(queryFilter, bson.M{param.Name: prop})
		case query.FilterTypeOperator:
			param, ok := filter.Params.(*query.FilterOperator)
			if !ok {
				return nil, query.ErrFilterInvalid
			}

			op, ok := internal.ParseFilterOperator(param)
			if !ok {
				continue
			}

			queryMatcher = append(queryMatcher, bson.M{
				"$match": bson.M{op: queryFilter},
			})

			queryFilter = nil
		default:
			return nil, query.ErrFilterInvalid
		}
	}

	if len(queryFilter) > 0 {
		queryMatcher = []bson.M{
			{
				"$match": bson.M{"$or": queryFilter},
			},
		}
	}

	return queryMatcher, nil
}
