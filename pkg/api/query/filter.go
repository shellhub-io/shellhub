package query

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"iter"
	"strings"
)

var (
	// ErrFilterInvalid is returned when a filter is not base64url-encoded JSON of the expected shape,
	// or when a bracket filter key lacks its field or operator, has a bracket inside either, carries
	// more, or is sent twice.
	ErrFilterInvalid = errors.New("filter is invalid")
	// ErrFilterPropertyInvalid is returned when a property node, at the top level or inside a group,
	// names a field or operator the store does not allow or carries a value it does not take, or when
	// the filter's shape is one the store does not evaluate: too many nodes, a group inside a group,
	// a group with no property, or a node of an unknown type or with Params of another shape.
	ErrFilterPropertyInvalid = errors.New("filter property is not valid")
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
	Raw string `query:"filter" json:"-"`

	// Data stores the decoded filters, filled by Unmarshal from Raw or by UnmarshalQuery from the
	// bracket filter parameters.
	Data []Filter `json:"-"`
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

// UnmarshalJSON decodes Params into the shape named by Type: a [FilterProperty], a [FilterOperator],
// or the [FilterGroup] of nodes a group holds, decoded the same way. It returns [ErrFilterInvalid]
// for any other Type, so a server that does not know a node refuses it instead of ignoring it, and
// for a group inside a group, before decoding it, so nesting cannot multiply the decoding work. It
// returns the decoding error when Params does not match the shape Type names.
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
	case FilterTypeGroup:
		group, err := unmarshalGroup(params)
		if err != nil {
			return err
		}
		f.Params = group

		return nil
	default:
		return ErrFilterInvalid
	}
}

func unmarshalGroup(data []byte) (FilterGroup, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	group := make(FilterGroup, len(items))
	for i, item := range items {
		var head struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(item, &head); err != nil {
			return nil, err
		}

		if head.Type == FilterTypeGroup {
			return nil, ErrFilterInvalid
		}

		if err := json.Unmarshal(item, &group[i]); err != nil {
			return nil, err
		}
	}

	return group, nil
}

// Properties yields every property node in order, those inside groups included, so a caller that
// checks or rewrites properties cannot miss one a group holds. Nodes whose Params is not a
// [FilterProperty] or a [FilterGroup] are skipped.
func (fs *Filters) Properties() iter.Seq[*FilterProperty] {
	return func(yield func(*FilterProperty) bool) {
		walkProperties(fs.Data, yield)
	}
}

func walkProperties(nodes []Filter, yield func(*FilterProperty) bool) bool {
	for _, node := range nodes {
		switch params := node.Params.(type) {
		case *FilterProperty:
			if !yield(params) {
				return false
			}
		case FilterGroup:
			if !walkProperties(params, yield) {
				return false
			}
		}
	}

	return true
}
