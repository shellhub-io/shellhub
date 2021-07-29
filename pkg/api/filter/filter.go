package filter

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
)

var (
	ErrFilterTypeUnknow       = errors.New("unknow filter type")
	ErrPropertyNameInvalid    = errors.New("invalid property name")
	ErrPropertyOperatorUnknow = errors.New("unknow property operator")
)

type (
	// Filter holds data for filtering results in a query.
	Filter struct {
		Type   string      `json:"type,omitempty"`
		Params interface{} `json:"params,omitempty"`
	}

	// FilterTypeProperty holds data to filter a property based on value and comparison operator.
	FilterTypeProperty struct {
		// Property name
		Name string `json:"name"`
		// Comparison operator
		//
		// contains: check if property contains the value
		// eq: check if the property is equal to value
		// bool: check if the property is true or false based on value
		// gt: check if the property is greater than value
		// lt: check if the property is less than value
		Operator string `json:"operator"`
		// Value to compare
		Value interface{} `json:"value"`
	}

	// FilterTypeOperator holds data to apply a conditional operator in a filter.
	FilterTypeOperator struct {
		// Conditional operator name
		//
		// and: AND conditional operator
		// or: OR conditional operator
		Name string `json:"name"`
	}

	// FilterList is a slice of Filter.
	FilterList []*Filter
)

func (f *Filter) UnmarshalJSON(data []byte) error {
	var params json.RawMessage

	type filter Filter
	aux := filter{
		Params: &params,
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch aux.Type {
	case "property":
		var property FilterTypeProperty
		if err := json.Unmarshal(params, &property); err != nil {
			return err
		}
		f.Params = &property
	case "operator":
		var operator FilterTypeOperator
		if err := json.Unmarshal(params, &operator); err != nil {
			return err
		}
		f.Params = &operator
	}

	f.Type = aux.Type

	return nil
}

// IsValid check if is valid.
func (f *Filter) IsValid() error {
	switch f.Type {
	case "property":
		if filterType, ok := f.Params.(*FilterTypeProperty); ok {
			return filterType.isValid()
		}
	case "operator":
		if filterType, ok := f.Params.(*FilterTypeOperator); ok {
			return filterType.isValid()
		}
	}

	return ErrFilterTypeUnknow
}

// isValid check if is valid.
func (f *FilterTypeProperty) isValid() error {
	supportedOperators := map[string]bool{
		"contains": true,
		"eq":       true,
		"bool":     true,
		"gt":       true,
		"lt":       true,
	}

	_, validOperator := supportedOperators[f.Operator]

	if !validOperator {
		return errors.Wrap(ErrPropertyOperatorUnknow, f.Operator)
	}

	if f.Name == "" {
		return ErrPropertyNameInvalid
	}

	return nil
}

// isValid check if is valid.
func (f *FilterTypeOperator) isValid() error {
	supportedOperators := map[string]bool{
		"and": true,
		"or":  true,
	}

	if _, ok := supportedOperators[f.Name]; !ok {
		return ErrPropertyOperatorUnknow
	}

	return nil
}

// UnmarshalParam converts a base64 filter string to a FilterList.
func (f *FilterList) UnmarshalParam(value string) error {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return errors.Wrap(err, "filters")
	}

	var filter FilterList
	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return err
	}

	*f = filter

	return nil
}

// IsValid check if filter list is valid.
func (f *FilterList) IsValid() error {
	for _, filter := range *f {
		if err := filter.IsValid(); err != nil {
			return err
		}
	}

	return nil
}
