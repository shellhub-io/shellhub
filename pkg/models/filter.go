package models

import (
	"encoding/json"
)

// Filter is a helper struct to filter results from the database.
// TODO: Gives a better explanation about the filter and how to use it.
type Filter struct {
	// Type os the filter. Type can be "property" or "operator". When Type is "property", the Params field must is set
	// to PropertyParams structure and when set "operator", the Params field must be set to OperatorParams structure.
	Type string `json:"type,omitempty"`
	// Params is the filter params. Params can be either PropertyParams or OperatorParams.
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

	switch aux.Type {
	case "property":
		var property PropertyParams
		if err := json.Unmarshal(params, &property); err != nil {
			return err
		}
		f.Params = &property
	case "operator":
		var operator OperatorParams
		if err := json.Unmarshal(params, &operator); err != nil {
			return err
		}
		f.Params = &operator
	}

	f.Type = aux.Type

	return nil
}

type PropertyParams struct {
	Name     string      `json:"name"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type OperatorParams struct {
	Name string `json:"name"`
}
