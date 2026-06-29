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
	// "online" mirrors production DeviceFilterFields: it accepts both "bool" and
	// "eq" so the virtual-field eq-with-bool-value path is exercised.
	// "realclosed" simulates a real boolean column that has both "bool" and "eq"
	// operators but is NOT declared as a virtual field — it must never accept
	// bool-convertible values for eq/ne because ParseFilterProperty will not
	// intercept it and the value would be bound directly to a Postgres boolean
	// column via fromEq, where a non-string bind causes a type-mismatch 500.
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
		// bool-operator value type validation
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
			// nil is a JSON primitive (isPrimitive passes), but isBoolConvertible(nil)
			// returns false, so the bool-operator guard must catch and reject it before
			// it reaches the store's fromBool/fromOnlineFilter (which return
			// ErrUnsupportedBoolType on nil, producing a 500 instead of 400).
			name: "bool operator with nil value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "online", Operator: "bool", Value: nil}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		// eq/ne operator scalar enforcement
		{
			name: "eq operator with array value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "eq", Value: []interface{}{"a", "b"}}},
			}},
			wantErr: ErrFilterPropertyInvalid,
		},
		{
			name: "ne operator with array value is rejected",
			filters: &Filters{Data: []Filter{
				{Type: FilterTypeProperty, Params: &FilterProperty{Name: "status", Operator: "ne", Value: []interface{}{"a"}}},
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
		// eq/ne must reject non-string scalars (float64/bool/nil) for text columns;
		// a numeric or bool bind parameter against a varchar column causes a Postgres
		// type-mismatch error (500) instead of a clean 400.
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
		// eq/ne with nil value is rejected regardless of field type.
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
		// Virtual bool-backed fields (those also allowing "bool") may use eq/ne with
		// bool-convertible values because ParseFilterProperty routes them to
		// fromOnlineFilter before any SQL column binding occurs.
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
		// realclosed is a real boolean column (not virtual): "bool" and "eq" are
		// both allowed operators, but since it is NOT declared as a virtual bool
		// field, eq/ne must require a string value.  Accepting a bool or float64
		// here would produce a Postgres type-mismatch 500 because ParseFilterProperty
		// does not intercept "realclosed" and would pass the raw bool/numeric to
		// fromEq, binding it against a boolean column via "boolean = numeric/bool".
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
		in   interface{}
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
		{"slice", []interface{}{true}, false},
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
