package paginator

import (
	"math"
)

const (
	MinPage    = 1   // MinPage represents the minimum allowed value for the pagination query's Page parameter.
	MinPerPage = 1   // MinPerPage represents the minimum allowed value for the pagination query's PerPage parameter.
	MaxPerPage = 100 // MaxPerPage represents the maximum allowed value for the pagination query's PerPage parameter.
)

// Query represents the pagination query parameters.
type Query struct {
	Page    int `query:"page"`     // Page represents the current page number.
	PerPage int `query:"per_page"` // PerPage represents the number of items per page.
}

// NewQuery creates a new pagination query with default values. If a custom pagination is
// provided in the request, it is advisable to normalize the values using the `Query.Normalize()`
// method.
func NewQuery() *Query {
	return &Query{
		Page:    MinPage,
		PerPage: MinPerPage,
	}
}

// Normalize ensures valid values for Page and PerPage in the pagination query.
// If query.PerPage is less than one, it is set to `MinPerPage`.
// If query.Page is less than one, it is set to `MinPage`.
// The maximum allowed value for query.PerPage is `MaxPerPage`.
func (q *Query) Normalize() {
	q.PerPage = int(math.Max(math.Min(float64(q.PerPage), float64(MaxPerPage)), float64(MinPerPage)))
	q.Page = int(math.Max(float64(MinPage), float64(q.Page)))
}
