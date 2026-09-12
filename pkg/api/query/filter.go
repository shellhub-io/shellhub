package query

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var (
	// ErrFilterNotBase64 is returned when a filter's raw value decodes under neither base64 alphabet.
	ErrFilterNotBase64 = errors.New("filter is not valid base64")
	// ErrFilterInvalid is returned when a filter decodes as base64 but does not parse as JSON. A raw
	// value that is not base64 at all returns [ErrFilterNotBase64] instead, and JSON that parses but
	// is not a filter returns [ErrFilterShapeInvalid].
	ErrFilterInvalid = errors.New("filter is invalid")
	// ErrFilterShapeInvalid is returned when a filter parses as JSON but is not a filter: not an
	// array of nodes, a node naming a type this package does not define, or a node whose params are
	// not the shape its type names. The JSON is well formed, which separates it from
	// [ErrFilterInvalid].
	ErrFilterShapeInvalid = errors.New("filter shape is not valid")
	// ErrFilterPropertyInvalid is returned when a property node names a field the contract does not
	// allow filtering on. A field the contract names but with an operator it does not allow returns
	// [ErrFilterOperatorInvalid] instead.
	ErrFilterPropertyInvalid = errors.New("filter property is not valid")
	// ErrFilterOperatorInvalid is returned when a property node names a field the contract allows
	// and an operator that field does not accept.
	ErrFilterOperatorInvalid = errors.New("filter operator is not valid")
	// ErrFilterValueInvalid is returned when a property node's value is not a primitive, or is of a
	// type the field's operator cannot compare against. A value of the right type that is merely too
	// long returns [ErrFilterValueTooLarge] instead.
	ErrFilterValueInvalid = errors.New("filter value is not valid")
	// ErrFilterValueTooLarge is returned when a property node's value is a type the field accepts but
	// exceeds [MaxStringValueLen] or [MaxArrayLen], directly or through an array item.
	ErrFilterValueTooLarge = errors.New("filter value exceeds the maximum size")
	// ErrFilterTooManyItems is returned when a filter carries more conditions than [MaxFilterItems].
	ErrFilterTooManyItems = errors.New("filter carries too many conditions")
	// ErrFilterTooLarge is returned when the encoded filter is longer than the cap, which bounds the
	// work a single query can ask for.
	ErrFilterTooLarge = errors.New("filter exceeds the maximum size")
	// ErrSorterFieldInvalid is returned when a sort names a field the store does not allow sorting on.
	ErrSorterFieldInvalid = errors.New("sort field is not valid")
)

// FilterError carries the input a refusal is about alongside the sentinel naming the refusal,
// so a caller can say which field or operator offended without validating the filter a second time
// to find out. Field is empty when the refusal is about the filter as a whole rather than one of
// its nodes; Operator is empty unless the operator itself is what was refused.
//
// It wraps its sentinel, so errors.Is against one of this package's sentinels holds either way.
type FilterError struct {
	Err      error
	Field    string
	Operator string
}

// Error returns the sentinel's own message, so a rejection reads exactly as the sentinel it wraps
// and no caller has to special-case the carrier when formatting it.
func (e *FilterError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the sentinel, which is what makes errors.Is hold against this package's sentinels
// for a rejection as it does for a bare one.
func (e *FilterError) Unwrap() error {
	return e.Err
}

func rejectFilter(err error) error {
	return &FilterError{Err: err}
}

func rejectField(err error, field string) error {
	return &FilterError{Err: err, Field: field}
}

func rejectOperator(err error, field, operator string) error {
	return &FilterError{Err: err, Field: field, Operator: operator}
}

// Filters represents a set of filters that can be applied to queries.
type Filters struct {
	// Raw holds the raw data of the filter. It must be a base64url-encoded JSON
	// (RFC 4648 §5, unpadded); also accepts padded/unpadded standard base64.
	Raw string `query:"filter"`

	// Data stores the decoded filters; it's automatically populated with the Unmarshal method.
	Data []Filter
}

// NewFilters creates a new instance of Filters with an empty Data slice.
func NewFilters() *Filters {
	return &Filters{Data: nil}
}

// Unmarshal decodes and unmarshals the raw filters, populating the Data attribute.
// It rejects payloads larger than [MaxFilterRawBytes] before decode to keep
// a hostile caller from allocating large buffers at JSON decode time.
//
// Both base64 alphabets are accepted, standard and URL-safe, with padding stripped first so either
// can be tried with its unpadded decoder.
func (fs *Filters) Unmarshal() error {
	if len(fs.Raw) > MaxFilterRawBytes {
		return ErrFilterTooLarge
	}

	unpadded := strings.TrimRight(fs.Raw, "=")

	raw, err := base64.RawStdEncoding.DecodeString(unpadded)
	if err != nil {
		raw, err = base64.RawURLEncoding.DecodeString(unpadded)
		if err != nil {
			return ErrFilterNotBase64
		}
	}

	if err := json.Unmarshal(raw, &fs.Data); len(raw) > 0 && err != nil {
		var syntax *json.SyntaxError
		if errors.As(err, &syntax) {
			return ErrFilterInvalid
		}

		return ErrFilterShapeInvalid
	}

	return nil
}

// Filtered is a request that carries a [Filters]. It is the filtering counterpart of [Paginated]
// and [Sorted], and exists for the same reason: it lets a caller holding only the request decode
// the filter without knowing the request's concrete type.
type Filtered interface {
	GetFilters() *Filters
}

// GetFilters returns the filters themselves, satisfying [Filtered] for every type that embeds it.
func (fs *Filters) GetFilters() *Filters {
	return fs
}

// Filter is one node of a query filter: a tagged union whose Type picks the shape of Params.
// Unmarshal one rather than building it by hand, or Params holds a map instead of a params struct.
type Filter struct {
	Type   string `json:"type,omitempty"`
	Params any    `json:"params,omitempty"`
}

// UnmarshalJSON decodes Params into the struct named by Type, and returns [ErrFilterShapeInvalid]
// when Type names no known shape or Params is not the shape it names. The enclosing JSON has
// already parsed by then, which is why this is not [ErrFilterInvalid].
func (f *Filter) UnmarshalJSON(data []byte) error {
	var params json.RawMessage

	type filter Filter
	aux := filter{
		Params: &params,
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	f.Type = aux.Type

	switch f.Type {
	case FilterTypeProperty:
		var property FilterProperty
		if err := json.Unmarshal(params, &property); err != nil {
			return ErrFilterShapeInvalid
		}
		f.Params = &property

		return nil
	case FilterTypeOperator:
		var operator FilterOperator
		if err := json.Unmarshal(params, &operator); err != nil {
			return ErrFilterShapeInvalid
		}
		f.Params = &operator

		return nil
	default:
		return ErrFilterShapeInvalid
	}
}
