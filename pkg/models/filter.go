package models

import (
	"encoding/json"
)

// Filter is one node of a query filter as it arrives from a client. It is a tagged union: Type
// picks which shape Params holds, and UnmarshalJSON is what enforces the pairing — build one by
// unmarshalling rather than by hand, or Params ends up a map instead of a params struct.
type Filter struct {
	// Type os the filter. Type can be "property" or "operator". When Type is "property", the Params field must is set
	// to PropertyParams structure and when set "operator", the Params field must be set to OperatorParams structure.
	Type string `json:"type,omitempty"`
	// Params is the filter params. Params can be either PropertyParams or OperatorParams.
	Params any `json:"params,omitempty"`
}

// UnmarshalJSON decodes Params into the struct Type names. An unrecognized Type is not an error:
// it leaves Params nil, so a filter the server does not understand narrows nothing rather than
// failing the request.
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

// PropertyParams compares one field against one value — the leaf of a filter tree.
type PropertyParams struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// OperatorParams joins the filters around it ("and", "or"), so it is what makes a filter list a
// tree rather than a conjunction.
type OperatorParams struct {
	Name string `json:"name"`
}
