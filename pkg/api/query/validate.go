package query

// Size limits applied to client-supplied filters. Together they bound the
// cost of serving a malicious filter payload: the outer limit caps the raw
// request size, the inner limits cap the structural complexity after decode.
const (
	// MaxFilterItems caps the number of entries in [Filters.Data].
	MaxFilterItems = 8
	// MaxStringValueLen caps a string in [FilterProperty.Value], and each
	// item when the value is an array.
	MaxStringValueLen = 256
	// MaxArrayLen caps the length of an array in [FilterProperty.Value].
	MaxArrayLen = 4
	// MaxFilterRawBytes caps the base64-encoded filter query parameter
	// before decode.
	MaxFilterRawBytes = 16 * 1024
)

// FieldSet is a set of field names allowed for use as a database identifier
// in sort options.
type FieldSet map[string]struct{}

// NewFieldSet returns a FieldSet containing the given names.
func NewFieldSet(names ...string) FieldSet {
	s := make(FieldSet, len(names))
	for _, n := range names {
		s[n] = struct{}{}
	}

	return s
}

// Allows reports whether name is in the set.
func (s FieldSet) Allows(name string) bool {
	_, ok := s[name]

	return ok
}

// FieldConstraints maps a filter field name to the set of operators allowed
// against it. It lets the handler reject field+operator combinations that
// the database rejects with a server-side error (e.g. ILIKE on an enum),
// turning them into a clean HTTP 400 before they reach the store.
type FieldConstraints map[string]FieldSet

// NewFieldConstraints returns a FieldConstraints initialized with the given
// field→operators pairs. An empty operators slice means the field is
// rejected entirely.
func NewFieldConstraints(entries map[string][]string) FieldConstraints {
	c := make(FieldConstraints, len(entries))
	for name, ops := range entries {
		c[name] = NewFieldSet(ops...)
	}

	return c
}

// Allows reports whether operator is valid for the given field name.
func (c FieldConstraints) Allows(name, operator string) bool {
	ops, ok := c[name]
	if !ok {
		return false
	}

	return ops.Allows(operator)
}

// ValidateSorter returns [ErrSorterFieldInvalid] if the sort field is set and
// not in allowed. An empty [Sorter.By] is valid (the store falls back to a
// stable default).
func ValidateSorter(sorter *Sorter, allowed FieldSet) error {
	if sorter == nil || sorter.By == "" {
		return nil
	}

	if !allowed.Allows(sorter.By) {
		return ErrSorterFieldInvalid
	}

	return nil
}

// ValidateFilters returns [ErrFilterPropertyInvalid] if any property filter
// references a (field, operator) pair not in constraints, carries a
// non-primitive Value, or exceeds the configured size limits. Operator
// filters (and/or) are left to the store to parse.
func ValidateFilters(filters *Filters, constraints FieldConstraints) error {
	if filters == nil {
		return nil
	}

	if len(filters.Data) > MaxFilterItems {
		return ErrFilterPropertyInvalid
	}

	for _, f := range filters.Data {
		if f.Type != FilterTypeProperty {
			continue
		}

		prop, ok := f.Params.(*FilterProperty)
		if !ok {
			return ErrFilterPropertyInvalid
		}

		if !constraints.Allows(prop.Name, prop.Operator) {
			return ErrFilterPropertyInvalid
		}

		if !isPrimitive(prop.Value) || !isValueWithinLimits(prop.Value) {
			return ErrFilterPropertyInvalid
		}
	}

	return nil
}

// isValueWithinLimits reports whether v respects the filter size limits:
// strings up to [MaxStringValueLen] and arrays up to [MaxArrayLen] (with
// each string inside the array also bounded by [MaxStringValueLen]).
func isValueWithinLimits(v interface{}) bool {
	switch x := v.(type) {
	case string:
		return len(x) <= MaxStringValueLen
	case []interface{}:
		if len(x) > MaxArrayLen {
			return false
		}
		for _, item := range x {
			if !isValueWithinLimits(item) {
				return false
			}
		}

		return true
	default:
		return true
	}
}

// isPrimitive reports whether v is a JSON primitive (string, number, bool, nil)
// or a slice of primitives. Maps and slices of maps are rejected because they
// smuggle operator-shaped keys into filter values. Only the types produced by
// [json.Unmarshal] into [interface{}] are listed — internal programmatic
// callers don't pass through this validator.
func isPrimitive(v interface{}) bool {
	switch x := v.(type) {
	case nil, bool, string, float64:
		return true
	case []interface{}:
		for _, item := range x {
			if !isPrimitive(item) {
				return false
			}
		}

		return true
	default:
		return false
	}
}
