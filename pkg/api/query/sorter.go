package query

const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// Sorter represents the sorting order in a query.
type Sorter struct {
	By    string `query:"sort_by"`
	Order string `query:"order_by" validate:"omitempty,oneof=asc desc"`
}

// NewOrder creates and returns a new Sort instance with the default descending order.
func NewSorter() *Sorter {
	return &Sorter{
		By:    "",
		Order: OrderDesc,
	}
}

// Normalize ensures that the sorting order is valid.
// If an invalid order is provided, it defaults to descending order.
func (s *Sorter) Normalize() {
	if s.Order != OrderAsc && s.Order != OrderDesc {
		s.Order = OrderDesc
	}
}
