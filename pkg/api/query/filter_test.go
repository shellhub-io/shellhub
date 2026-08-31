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
	// Pre-computed base64 strings for various padding scenarios:
	//
	//   json0pad  = `[{"type":"operator","params":{"name":"and"}}]`          → 0 padding chars
	//   json1pad  = `[{"type":"property","params":{"name":"a","operator":"eq","value":"b"}}]`  → 1 padding char
	//   json2pad  = `[{"type":"property","params":{"name":"aa","operator":"eq","value":"bb"}}]` → 2 padding chars
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
			// base64url alphabet uses '-' in place of '+'; the JSON payload contains
			// the string ">>>" whose base64 representation differs between the two
			// alphabets. This row verifies that Unmarshal handles base64url (RFC 4648 §5).
			// stdFails is set so the test asserts RawStdEncoding cannot decode this input,
			// confirming the two alphabets are genuinely distinct for this payload.
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
	// The character 鸿 (U+9E3F, UTF-8: E9 B8 BF) encodes to a base64 group that
	// contains '/' in RawStdEncoding and '_' in RawURLEncoding, making it an ideal
	// probe for the two alphabets.
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
