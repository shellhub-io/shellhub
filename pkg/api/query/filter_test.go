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

func TestFilters_Unmarshal(t *testing.T) {
	// Pre-computed base64 strings for various padding scenarios:
	//
	//   json0pad  = `[{"type":"operator","params":{"name":"and"}}]`          → 0 padding chars
	//   json1pad  = `[{"type":"property","params":{"name":"a","operator":"eq","value":"b"}}]`  → 1 padding char
	//   json2pad  = `[{"type":"property","params":{"name":"aa","operator":"eq","value":"bb"}}]` → 2 padding chars
	const (
		// 0-padding: StdEncoding and RawStdEncoding produce the same string.
		b64Padded0 = "W3sidHlwZSI6Im9wZXJhdG9yIiwicGFyYW1zIjp7Im5hbWUiOiJhbmQifX1d"

		// 1-padding: StdEncoding appends one '='.
		b64Padded1  = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoiYiJ9fV0="
		b64Missing1 = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoiYiJ9fV0"

		// 2-padding: StdEncoding appends two '='.
		b64Padded2  = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhYSIsIm9wZXJhdG9yIjoiZXEiLCJ2YWx1ZSI6ImJiIn19XQ=="
		b64Missing2 = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhYSIsIm9wZXJhdG9yIjoiZXEiLCJ2YWx1ZSI6ImJiIn19XQ"
	)

	cases := []struct {
		description string
		raw         string
		wantErr     error
	}{
		{
			description: "padded base64 (0 padding) round-trips unchanged",
			raw:         b64Padded0,
			wantErr:     nil,
		},
		{
			description: "padded base64 (1 padding) round-trips unchanged",
			raw:         b64Padded1,
			wantErr:     nil,
		},
		{
			description: "padded base64 (2 padding) round-trips unchanged",
			raw:         b64Padded2,
			wantErr:     nil,
		},
		{
			description: "one '=' missing still decodes",
			raw:         b64Missing1,
			wantErr:     nil,
		},
		{
			description: "two '=' missing still decodes",
			raw:         b64Missing2,
			wantErr:     nil,
		},
		{
			description: "zero-padding raw input (no '=' at all) decodes",
			// b64Padded0 already has no padding; use it directly as the raw input.
			raw:     b64Padded0,
			wantErr: nil,
		},
		{
			description: "non-base64 garbage returns ErrFilterInvalid",
			raw:         "!!!not-base64!!!",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "invalid JSON in valid base64 returns ErrFilterInvalid",
			// base64("not valid json")
			raw:     base64.StdEncoding.EncodeToString([]byte("not valid json")),
			wantErr: ErrFilterInvalid,
		},
		{
			description: "valid base64 + valid JSON with unknown filter type returns ErrFilterInvalid",
			// base64(`[{"type":"unknown","params":{}}]`)
			raw:     base64.StdEncoding.EncodeToString([]byte(`[{"type":"unknown","params":{}}]`)),
			wantErr: ErrFilterInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			fs := &Filters{Raw: tc.raw}
			err := fs.Unmarshal()

			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
			}
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
