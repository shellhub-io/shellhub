package internal

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromCustomFieldsFilter(t *testing.T) {
	cases := []struct {
		description string
		operator    string
		value       any
		wantSQL     string
		wantArgs    []any
		wantOk      bool
		wantErr     error
	}{
		{
			description: "returns SQL for contains with string value",
			operator:    "contains",
			value:       "production",
			wantSQL:     `EXISTS (SELECT 1 FROM jsonb_each_text("device"."custom_fields") WHERE value ILIKE ?)`,
			wantArgs:    []any{"%production%"},
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "wraps value with %% wildcards",
			operator:    "contains",
			value:       "team",
			wantSQL:     `EXISTS (SELECT 1 FROM jsonb_each_text("device"."custom_fields") WHERE value ILIKE ?)`,
			wantArgs:    []any{"%team%"},
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "returns not-ok for unsupported operator eq",
			operator:    "eq",
			value:       "production",
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     nil,
		},
		{
			description: "returns not-ok for unsupported operator ne",
			operator:    "ne",
			value:       "production",
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     nil,
		},
		{
			description: "returns error when value is not a string",
			operator:    "contains",
			value:       42,
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     ErrUnsupportedContainsType,
		},
		{
			description: "returns error when value is a slice",
			operator:    "contains",
			value:       []any{"a", "b"},
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     ErrUnsupportedContainsType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sql, args, ok, err := fromCustomFieldsFilter(tc.operator, tc.value)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantSQL, sql)
			assert.Equal(t, tc.wantArgs, args)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestParseFilterProperty_CustomFields(t *testing.T) {
	cases := []struct {
		description string
		fp          *query.FilterProperty
		wantSQL     string
		wantArgs    []any
		wantOk      bool
		wantErr     bool
	}{
		{
			description: "routes custom_fields contains to JSONB subquery",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: "prod"},
			wantSQL:     `EXISTS (SELECT 1 FROM jsonb_each_text("device"."custom_fields") WHERE value ILIKE ?)`,
			wantArgs:    []any{"%prod%"},
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "returns error for custom_fields contains with non-string value",
			fp:          &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: 123},
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sql, args, ok, err := ParseFilterProperty(tc.fp, "device")
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantSQL, sql)
			assert.Equal(t, tc.wantArgs, args)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
