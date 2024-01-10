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
