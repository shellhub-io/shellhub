package query

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

var (
	ErrFilterInvalid         = errors.New("filter is invalid")
	ErrFilterPropertyInvalid = errors.New("filter property is not valid")
	ErrFilterOperatorInvalid = errors.New("filter operator is not valid")
)

// Filters represents a set of filters that can be applied to queries.
type Filters struct {
	// Raw holds the raw data of the filter and it's a base64-encoded JSON.
	Raw string `query:"filter"`

	// Data stores the decoded filters; it's automatically populated with the Unmarshal method.
	Data []Filter
}

// NewFilters creates a new instance of Filters with an empty Data slice.
func NewFilters() *Filters {
	return &Filters{Data: nil}
}

// Unmarshal decodes and unmarshals the raw filters, populating the Data attribute.
func (fs *Filters) Unmarshal() error {
	raw, err := base64.StdEncoding.DecodeString(fs.Raw)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(raw, &fs.Data); len(raw) > 0 && err != nil {
		return err
	}

	return nil
}

type Filter struct {
	Type   string      `json:"type,omitempty"`
	Params interface{} `json:"params,omitempty"`
}

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
