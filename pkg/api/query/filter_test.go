package query

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterUnmarshalJSON(t *testing.T) {
	cases := []struct {
		description string
		filter      *Filter
		data        string
		expected    error
	}{
		{
			description: "",
			filter: &Filter{
				Type: "property",
				Params: FilterProperty{
					Name:     "online",
					Operator: "bool",
					Value:    "true",
				},
			},
			// {
			//     "type": "property",
			//     "params": {
			//         "name": "online",
			//         "operator": "bool",
			//         "value": "true"
			//     }
			// }
			data:     "ewogICAgInR5cGUiOiAicHJvcGVydHkiLAogICAgInBhcmFtcyI6IHsKICAgICAgICAibmFtZSI6ICJvbmxpbmUiLAogICAgICAgICJvcGVyYXRvciI6ICJib29sIiwKICAgICAgICAidmFsdWUiOiAidHJ1ZSIKICAgIH0KfQ==",
			expected: nil,
		},
		{
			description: "",
			filter: &Filter{
				Type: "operator",
				Params: FilterOperator{
					Name: "and",
				},
			},
			// {
			//     "type": "operator",
			//     "params": {
			//         "name": "and"
			//     }
			// }
			data:     "ewogICAgInR5cGUiOiAib3BlcmF0b3IiLAogICAgInBhcmFtcyI6IHsKICAgICAgICAibmFtZSI6ICJhbmQiCiAgICB9Cn0=",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			raw, err := base64.StdEncoding.DecodeString(tc.data)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, tc.filter.UnmarshalJSON(raw))
		})
	}
}

func TestFilters_Unmarshal_sizeLimit(t *testing.T) {
	t.Run("raw under the limit passes", func(t *testing.T) {
		fs := &Filters{Raw: base64.StdEncoding.EncodeToString([]byte(`[{"type":"operator","params":{"name":"and"}}]`))}
		assert.NoError(t, fs.Unmarshal())
	})

	t.Run("raw over the limit is rejected pre-decode", func(t *testing.T) {
		fs := &Filters{Raw: string(make([]byte, MaxFilterRawBytes+1))}
		assert.Equal(t, ErrFilterTooLarge, fs.Unmarshal())
	})
}
