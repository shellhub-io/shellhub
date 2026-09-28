package query

// Contract is what one resource's list endpoint accepts from a client: the filter fields with the
// operators valid against each, the fields it may be ordered by, and the order it is served in when
// the client names none.
//
// The zero value accepts no filter and no sort, which is what a resource taking neither declares —
// an omitted contract is indistinguishable from a forgotten one, so there is no such thing.
type Contract struct {
	Filter FieldConstraints
	Sort   FieldSet

	// DefaultSort is applied to a sorter the client left blank. It is applied before the sort is
	// validated, so a default naming a field outside Sort is refused exactly as a client's would be.
	DefaultSort Sorter
}

// NormalizeSorter fills from the contract's default whatever the client left blank, then holds the
// order to asc or desc. A zero contract leaves the sorter as [Sorter.Normalize] would.
func (c Contract) NormalizeSorter(sorter *Sorter) {
	if sorter == nil {
		return
	}

	if sorter.By == "" {
		sorter.By = c.DefaultSort.By
	}

	if sorter.Order == "" {
		sorter.Order = c.DefaultSort.Order
	}

	sorter.Normalize()
}
