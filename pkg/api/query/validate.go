package query

import "strconv"

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
//
// A subset of fields may be declared as virtual bool-backed at construction
// time (see NewFieldConstraints). Virtual fields are intercepted by
// ParseFilterProperty before any SQL column binding, so they safely accept
// bool-convertible values with eq/ne. Real boolean columns must NOT be
// declared virtual unless a corresponding ParseFilterProperty intercept exists.
type FieldConstraints struct {
	operators    map[string]FieldSet
	virtualBools FieldSet
}

// NewFieldConstraints returns a FieldConstraints initialized with the given
// field→operators pairs. An empty operators slice means the field is rejected
// entirely. virtualBools lists the field names that are virtual bool-backed:
// they are intercepted by ParseFilterProperty before SQL column binding, so
// they may accept bool-convertible values with eq/ne. Only fields that have a
// corresponding ParseFilterProperty intercept should be listed here.
func NewFieldConstraints(entries map[string][]string, virtualBools ...string) FieldConstraints {
	operators := make(map[string]FieldSet, len(entries))
	for name, ops := range entries {
		operators[name] = NewFieldSet(ops...)
	}

	return FieldConstraints{
		operators:    operators,
		virtualBools: NewFieldSet(virtualBools...),
	}
}

// Allows reports whether operator is valid for the given field name.
func (c FieldConstraints) Allows(name, operator string) bool {
	ops, ok := c.operators[name]
	if !ok {
		return false
	}

	return ops.Allows(operator)
}

// IsVirtualBoolField reports whether name is a virtual bool-backed field —
// one explicitly declared as virtual at construction time via NewFieldConstraints
// and therefore intercepted by ParseFilterProperty before any SQL column binding.
//
// Only fields in the virtualBools registry return true. Using the presence of
// the "bool" operator as a proxy is incorrect: a real boolean column that only
// allows "bool" (not "eq"/"ne") is safe today, but the moment "eq" is added to
// it without a ParseFilterProperty intercept, the validator would silently accept
// bool/float64 values that produce a Postgres type-mismatch 500 at runtime.
func (c FieldConstraints) IsVirtualBoolField(name string) bool {
	return c.virtualBools.Allows(name)
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
//
// Equality on a virtual bool-backed field (see [FieldConstraints.IsVirtualBoolField]) accepts
// anything bool-convertible, because ParseFilterProperty intercepts those before any column is
// bound. Every other field must be compared against a string, or Postgres answers a type mismatch
// with a 500.
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

		if prop.Operator == "bool" && !isBoolConvertible(prop.Value) {
			return ErrFilterPropertyInvalid
		}

		if prop.Operator == "eq" || prop.Operator == "ne" {
			if constraints.IsVirtualBoolField(prop.Name) {
				if !isBoolConvertible(prop.Value) {
					return ErrFilterPropertyInvalid
				}
			} else {
				if _, ok := prop.Value.(string); !ok {
					return ErrFilterPropertyInvalid
				}
			}
		}
	}

	return nil
}

func isValueWithinLimits(v any) bool {
	switch x := v.(type) {
	case string:
		return len(x) <= MaxStringValueLen
	case []any:
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

func isBoolConvertible(v any) bool {
	switch x := v.(type) {
	case bool:
		return true
	case float64:
		return true
	case string:
		_, err := strconv.ParseBool(x)

		return err == nil
	default:
		return false
	}
}

func isPrimitive(v any) bool {
	switch x := v.(type) {
	case nil, bool, string, float64:
		return true
	case []any:
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
