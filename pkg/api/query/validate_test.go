package query

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldSet_Allows(t *testing.T) {
	set := NewFieldSet("name", "status", "tags.name")

	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"allowed single", "name", true},
		{"allowed dotted", "tags.name", true},
		{"not allowed", "tenant_id", false},
		{"empty", "", false},
		{"operator prefix", "$where", false},
		{"case sensitive", "Name", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, set.Allows(tc.in))
		})
	}
}

func TestFieldConstraints_Allows(t *testing.T) {
	constraints := NewFieldConstraints(map[string][]string{
		"name":   {"contains", "eq", "ne"},
		"status": {"eq", "ne"},
		"online": {"bool"},
		"banned": {},
	})

	cases := []struct {
		name     string
		field    string
		operator string
		want     bool
	}{
		{"allowed field+operator", "name", "contains", true},
		{"allowed field, other allowed operator", "name", "eq", true},
		{"allowed field, disallowed operator", "status", "contains", false},
		{"field with single operator allowed", "online", "bool", true},
		{"field with single operator disallowed", "online", "eq", false},
		{"field not in constraints", "tenant_id", "eq", false},
		{"empty operators slice rejects the field", "banned", "eq", false},
		{"empty field name", "", "eq", false},
		{"empty operator", "name", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, constraints.Allows(tc.field, tc.operator))
		})
	}
}

func TestValidateSorter(t *testing.T) {
	allowed := NewFieldSet("name", "last_seen")

	cases := []struct {
		name    string
		sorter  *Sorter
		wantErr error
	}{
		{"nil sorter", nil, nil},
		{"empty By", &Sorter{By: ""}, nil},
		{"allowed field", &Sorter{By: "name"}, nil},
		{"disallowed field", &Sorter{By: "tenant_id"}, ErrSorterFieldInvalid},
		{"operator as field", &Sorter{By: "$where"}, ErrSorterFieldInvalid},
		{"oversized", &Sorter{By: string(make([]byte, 5000))}, ErrSorterFieldInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantErr, ValidateSorter(tc.sorter, allowed))
		})
	}
}

func TestValidateFilters(t *testing.T) {
	allowed := NewFieldConstraints(map[string][]string{
		"name":       {"contains", "eq"},
		"status":     {"eq", "ne"},
		"online":     {"bool", "eq"},
		"realclosed": {"bool", "eq"},
	}, "online" /* virtual bool fields */)

	cases := []struct {
		name    string
		filters *Filters
		wantErr error
	}{
		{
			name:    "nil filters",
			filters: nil,
			wantErr: nil,
		},
		{
			name:    "empty data",
			filters: &Filters{Data: nil},
			wantErr: nil,
		},
		{
			name: "operator-only filter is passed through",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}},
			}},
			wantErr: nil,
		},
		{
			name: "allowed property with primitive value",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: "srv"}},
			}},
			wantErr: nil,
		},
		{
			name: "allowed property with array of primitives",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: []any{"a", "b"}}},
			}},
			wantErr: nil,
		},
		{
			name: "field not in allowlist",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "tenant_id", Operator: "eq", Value: "x"}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "operator not allowed for field",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "contains", Value: "x"}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "mongo operator as field",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "$where", Operator: "contains", Value: "x"}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "nested object as value",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: map[string]any{"$ne": "accepted"}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "array of objects as value",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: []any{map[string]any{"$ne": "accepted"}}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "wrong params type",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: "not-a-filter-property"},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "too many items",
			filters: func() *Filters {
				data := make([]Filter, MaxFilterItems+1)
				for i := range data {
					data[i] = Filter{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "eq", Value: "x"}}
				}

				return &Filters{Data: data}
			}(),
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "string value over limit",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "eq", Value: strings.Repeat("A", MaxStringValueLen+1)}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "array over length limit",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: func() []any {
					a := make([]any, MaxArrayLen+1)
					for i := range a {
						a[i] = "x"
					}

					return a
				}()}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "array item string over limit",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: []any{strings.Repeat("A", MaxStringValueLen+1)}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "bool operator with bool value is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: true}},
			}},
			wantErr: nil,
		},
		{
			name: "bool operator with float64 (JSON number) is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: float64(1)}},
			}},
			wantErr: nil,
		},
		{
			name: "bool operator with bool-parseable string is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: "true"}},
			}},
			wantErr: nil,
		},
		{
			name: "bool operator with non-bool string is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: "yes"}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "bool operator with nil value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: nil}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "eq operator with array value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: []any{"a", "b"}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "ne operator with array value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: []any{"a"}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "eq operator with scalar string is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: "pending"}},
			}},
			wantErr: nil,
		},
		{
			name: "ne operator with scalar string is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: "pending"}},
			}},
			wantErr: nil,
		},
		{
			name: "eq operator with float64 (JSON number) is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: float64(123)}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "eq operator with bool value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: true}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "ne operator with float64 (JSON number) is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: float64(0)}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "ne operator with bool value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: false}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "eq operator with nil value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: nil}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "ne operator with nil value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: nil}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "online: eq operator with bool true is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: true}},
			}},
			wantErr: nil,
		},
		{
			name: "online: eq operator with bool false is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: false}},
			}},
			wantErr: nil,
		},
		{
			name: "online: eq operator with float64 (JSON number) is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: float64(1)}},
			}},
			wantErr: nil,
		},
		{
			name: "online: eq operator with bool-parseable string is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: "true"}},
			}},
			wantErr: nil,
		},
		{
			name: "online: eq operator with non-bool string is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: "yes"}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "online: eq operator with nil is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "eq", Value: nil}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "realclosed (non-virtual): eq with bool value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "realclosed", Operator: "eq", Value: true}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "realclosed (non-virtual): eq with float64 (JSON number) is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "realclosed", Operator: "eq", Value: float64(1)}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "realclosed (non-virtual): eq with string value is accepted",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "realclosed", Operator: "eq", Value: "true"}},
			}},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantErr, ValidateFilters(tc.filters, allowed))
		})
	}
}

func TestIsBoolConvertible(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want bool
	}{
		{"bool true", true, true},
		{"bool false", false, true},
		{"float64 nonzero (JSON number)", float64(1), true},
		{"float64 zero (JSON number)", float64(0), true},
		{"string true", "true", true},
		{"string 1", "1", true},
		{"string false", "false", true},
		{"string 0", "0", true},
		{"string t", "t", true},
		{"string yes — rejected by ParseBool", "yes", false},
		{"string on — rejected by ParseBool", "on", false},
		{"string x — invalid", "x", false},
		{"nil", nil, false},
		{"int — not a JSON type", 1, false},
		{"slice", []any{true}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isBoolConvertible(tc.in))
		})
	}
}

func TestIsPrimitive(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want bool
	}{
		{"nil", nil, true},
		{"bool", true, true},
		{"string", "x", true},
		{"float64 (JSON number)", 3.14, true},
		{"int is not a JSON primitive", 42, false},
		{"int64 is not a JSON primitive", int64(42), false},
		{"array of primitives", []any{"a", 3.14, true}, true},
		{"empty array", []any{}, true},
		{"map", map[string]any{"$ne": "x"}, false},
		{"array with map", []any{map[string]any{"$ne": "x"}}, false},
		{"nested array of primitives", []any{[]any{"a", "b"}}, true},
		{"struct", struct{ X int }{X: 1}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isPrimitive(tc.in))
		})
	}
}
