package query

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var (
	// ErrFilterInvalid is returned when a filter is not base64url-encoded JSON of the expected shape.
	ErrFilterInvalid = errors.New("filter is invalid")
	// ErrFilterPropertyInvalid is returned when a property node names a field the store does not
	// allow filtering on.
	ErrFilterPropertyInvalid = errors.New("filter property is not valid")
	// ErrFilterOperatorInvalid is returned when a node names an operator the store does not support.
	ErrFilterOperatorInvalid = errors.New("filter operator is not valid")
	// ErrFilterTooLarge is returned when the encoded filter is longer than the cap, which bounds the
	// work a single query can ask for.
	ErrFilterTooLarge = errors.New("filter exceeds the maximum size")
	// ErrSorterFieldInvalid is returned when a sort names a field the store does not allow sorting on.
	ErrSorterFieldInvalid = errors.New("sort field is not valid")
)

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
		// Fall back to RawURLEncoding (RFC 4648 §5) whose alphabet uses '-' and '_'
		// instead of '+' and '/'.
		raw, err = base64.RawURLEncoding.DecodeString(unpadded)
		if err != nil {
			return ErrFilterInvalid
		}
	}

	if err := json.Unmarshal(raw, &fs.Data); len(raw) > 0 && err != nil {
		return ErrFilterInvalid
	}

	return nil
}

// Filter is one node of a query filter: a tagged union whose Type picks the shape of Params.
// Unmarshal one rather than building it by hand, or Params holds a map instead of a params struct.
type Filter struct {
	Type   string `json:"type,omitempty"`
	Params any    `json:"params,omitempty"`
}

// UnmarshalJSON decodes Params into the struct named by Type. An unrecognized Type leaves Params
// nil rather than failing, so a filter the server does not know narrows nothing.
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
			return err
		}
		f.Params = &property

		return nil
	case FilterTypeOperator:
		var operator FilterOperator
		if err := json.Unmarshal(params, &operator); err != nil {
			return err
		}
		f.Params = &operator

		return nil
	default:
		return ErrFilterInvalid
	}
}
