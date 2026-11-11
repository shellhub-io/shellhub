package query

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			data:     "ewogICAgInR5cGUiOiAib3BlcmF0b3IiLAogICAgInBhcmFtcyI6IHsKICAgICAgICAibmFtZSI6ICJhbmQiCiAgICB9Cn0=",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			raw, err := base64.StdEncoding.DecodeString(tc.data)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, tc.filter.UnmarshalJSON(raw))
		})
	}
}

func TestFilters_Unmarshal(t *testing.T) {
	const (
		b64Padded0 = "W3sidHlwZSI6Im9wZXJhdG9yIiwicGFyYW1zIjp7Im5hbWUiOiJhbmQifX1d"

		b64Padded1  = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoiYiJ9fV0="
		b64Missing1 = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoiYiJ9fV0"

		b64Padded2  = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhYSIsIm9wZXJhdG9yIjoiZXEiLCJ2YWx1ZSI6ImJiIn19XQ=="
		b64Missing2 = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhYSIsIm9wZXJhdG9yIjoiZXEiLCJ2YWx1ZSI6ImJiIn19XQ"
	)

	b64URLTripleGT := base64.RawURLEncoding.EncodeToString(
		[]byte(`[{"type":"property","params":{"name":"a","operator":"eq","value":">>>"}}]`),
	)

	cases := []struct {
		description   string
		raw           string
		wantErr       error
		wantFirstType string
		wantData      []Filter
		stdFails      bool
	}{
		{
			description:   "padded base64 (0 padding) round-trips unchanged",
			raw:           b64Padded0,
			wantErr:       nil,
			wantFirstType: FilterTypeOperator,
			wantData: []Filter{
				{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}},
			},
		},
		{
			description:   "padded base64 (1 padding) round-trips unchanged",
			raw:           b64Padded1,
			wantErr:       nil,
			wantFirstType: FilterTypeProperty,
			wantData: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "a", Operator: "eq", Value: "b"}},
			},
		},
		{
			description:   "padded base64 (2 padding) round-trips unchanged",
			raw:           b64Padded2,
			wantErr:       nil,
			wantFirstType: FilterTypeProperty,
			wantData: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "aa", Operator: "eq", Value: "bb"}},
			},
		},
		{
			description:   "one '=' missing still decodes",
			raw:           b64Missing1,
			wantErr:       nil,
			wantFirstType: FilterTypeProperty,
			wantData: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "a", Operator: "eq", Value: "b"}},
			},
		},
		{
			description:   "two '=' missing still decodes",
			raw:           b64Missing2,
			wantErr:       nil,
			wantFirstType: FilterTypeProperty,
			wantData: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "aa", Operator: "eq", Value: "bb"}},
			},
		},
		{
			description:   "zero-padding raw input (no '=' at all) decodes",
			raw:           b64Padded0,
			wantErr:       nil,
			wantFirstType: FilterTypeOperator,
			wantData: []Filter{
				{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}},
			},
		},
		{
			description:   "base64url alphabet (value '>>>') decodes correctly",
			raw:           b64URLTripleGT,
			wantErr:       nil,
			wantFirstType: FilterTypeProperty,
			stdFails:      true,
			wantData: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "a", Operator: "eq", Value: ">>>"}},
			},
		},
		{
			description: "non-base64 garbage returns ErrFilterInvalid",
			raw:         "!!!not-base64!!!",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "invalid JSON in valid base64 returns ErrFilterInvalid",
			raw:         base64.StdEncoding.EncodeToString([]byte("not valid json")),
			wantErr:     ErrFilterInvalid,
		},
		{
			description:   "a group decodes the items it holds",
			raw:           base64.RawURLEncoding.EncodeToString([]byte(`[{"type":"group","params":[{"type":"property","params":{"name":"a","operator":"eq","value":"b"}},{"type":"operator","params":{"name":"or"}}]}]`)),
			wantErr:       nil,
			wantFirstType: FilterTypeGroup,
			wantData: []Filter{
				{Type: FilterTypeGroup, Params: FilterGroup{
					{Type: FilterTypeProperty, Params: &FilterProperty{Name: "a", Operator: "eq", Value: "b"}},
					{Type: FilterTypeOperator, Params: &FilterOperator{Name: "or"}},
				}},
			},
		},
		{
			description: "a group whose params is not a list returns ErrFilterInvalid",
			raw:         base64.RawURLEncoding.EncodeToString([]byte(`[{"type":"group","params":{"name":"a"}}]`)),
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a group inside a group returns ErrFilterInvalid",
			raw:         base64.RawURLEncoding.EncodeToString([]byte(`[{"type":"group","params":[{"type":"group","params":[{"type":"property","params":{"name":"a","operator":"eq","value":"b"}}]}]}]`)),
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "valid base64 + valid JSON with unknown filter type returns ErrFilterInvalid",
			raw:         base64.StdEncoding.EncodeToString([]byte(`[{"type":"unknown","params":{}}]`)),
			wantErr:     ErrFilterInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.stdFails {
				_, errStd := base64.RawStdEncoding.DecodeString(tc.raw)
				require.Error(t, errStd, "RawStdEncoding must not decode a RawURLEncoding-only string")
			}

			fs := &Filters{Raw: tc.raw}
			err := fs.Unmarshal()

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}

			if tc.wantFirstType != "" {
				assert.Equal(t, tc.wantData, fs.Data)
			}
		})
	}
}

// TestFilters_Unmarshal_base64url verifies that Unmarshal accepts base64url-encoded
// input (RFC 4648 §5, characters - and _ in place of + and /), both unpadded and padded.
func TestFilters_Unmarshal_base64url(t *testing.T) {
	const (
		b64URLRaw    = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoi6bi_In19XQ"
		b64URLPadded = "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhIiwib3BlcmF0b3IiOiJlcSIsInZhbHVlIjoi6bi_In19XQ=="
	)

	cases := []struct {
		description string
		raw         string
		wantErr     error
	}{
		{
			description: "unpadded base64url (RFC 4648 §5) decodes correctly",
			raw:         b64URLRaw,
			wantErr:     nil,
		},
		{
			description: "padded base64url decodes correctly",
			raw:         b64URLPadded,
			wantErr:     nil,
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

func TestFilters_Properties(t *testing.T) {
	top := &FilterProperty{Name: "a", Operator: "eq", Value: "1"}
	first := &FilterProperty{Name: "b", Operator: "eq", Value: "2"}
	second := &FilterProperty{Name: "c", Operator: "eq", Value: "3"}

	filters := &Filters{Data: []Filter{
		{Type: FilterTypeProperty, Params: top},
		{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}},
		{Type: FilterTypeGroup, Params: FilterGroup{
			{Type: FilterTypeProperty, Params: first},
			{Type: FilterTypeOperator, Params: &FilterOperator{Name: "or"}},
			{Type: FilterTypeProperty, Params: second},
		}},
	}}

	var visited []*FilterProperty
	for property := range filters.Properties() {
		visited = append(visited, property)
	}

	assert.Equal(t, []*FilterProperty{top, first, second}, visited)
}
