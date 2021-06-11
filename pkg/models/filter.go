package models

import (
	"encoding/json"
)

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
