package query

// The sort directions a query may ask for. Anything else normalizes to OrderDesc.
const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// Sorter represents the sorting order in a query.
type Sorter struct {
	By       string `query:"sort_by"`
	Order    string `query:"order_by" validate:"omitempty,oneof=asc desc"`
	Tiebreak string // stable secondary sort column, set by service layer
}

// Normalize ensures that the sorting order is valid.
// If an invalid order is provided, it defaults to descending order.
func (s *Sorter) Normalize() {
	if s.Order != OrderAsc && s.Order != OrderDesc {
		s.Order = OrderDesc
	}
}

// Sorted is a request that carries a [Sorter]. Every request type embedding one satisfies it, so a
// caller holding only the request can normalize the sort order without knowing its concrete type.
type Sorted interface {
	GetSorter() *Sorter
}

// GetSorter returns the sorter itself, satisfying [Sorted] for every type that embeds it.
func (s *Sorter) GetSorter() *Sorter {
	return s
}
