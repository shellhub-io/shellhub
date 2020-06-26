package paginator

import (
	"math"
)

type Query struct {
	Page    int `query:"page"`
	PerPage int `query:"per_page"`
}

func NewQuery() *Query {
	return &Query{1, 25}
}

func (q *Query) Normalize() {
	// min value allowed 1 and max 100
	q.PerPage = int(math.Max(math.Min(float64(q.PerPage), 100), 1))
	// min value allowed 1
	q.Page = int(math.Max(1, float64(q.Page)))
}
