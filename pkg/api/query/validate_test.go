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
		"name":   {"contains", "eq"},
		"status": {"eq", "ne"},
	})

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
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: []interface{}{"a", "b"}}},
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
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: map[string]interface{}{"$ne": "accepted"}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "array of objects as value",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: []interface{}{map[string]interface{}{"$ne": "accepted"}}}},
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
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: func() []interface{} {
					a := make([]interface{}, MaxArrayLen+1)
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
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "name", Operator: "contains", Value: []interface{}{strings.Repeat("A", MaxStringValueLen+1)}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantErr, ValidateFilters(tc.filters, allowed))
		})
	}
}

func TestIsPrimitive(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want bool
	}{
		{"nil", nil, true},
		{"bool", true, true},
		{"string", "x", true},
		{"float64 (JSON number)", 3.14, true},
		{"int is not a JSON primitive", 42, false},
		{"int64 is not a JSON primitive", int64(42), false},
		{"array of primitives", []interface{}{"a", 3.14, true}, true},
		{"empty array", []interface{}{}, true},
		{"map", map[string]interface{}{"$ne": "x"}, false},
		{"array with map", []interface{}{map[string]interface{}{"$ne": "x"}}, false},
		{"nested array of primitives", []interface{}{[]interface{}{"a", "b"}}, true},
		{"struct", struct{ X int }{X: 1}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isPrimitive(tc.in))
		})
	}
}
