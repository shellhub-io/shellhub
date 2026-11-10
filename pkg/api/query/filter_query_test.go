package query

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiltersUnmarshalQuery(t *testing.T) {
	and := Filter{Type: FilterTypeOperator, Params: &FilterOperator{Name: "and"}}
	property := func(name, operator, value string) Filter {
		return Filter{Type: FilterTypeProperty, Params: &FilterProperty{Name: name, Operator: operator, Value: value}}
	}

	cases := []struct {
		description string
		query       string
		want        []Filter
		wantErr     error
	}{
		{
			description: "no bracket parameter leaves the filter empty",
			query:       "page=2&per_page=10&filter=e30",
			want:        nil,
		},
		{
			description: "one parameter becomes one property",
			query:       "filters[hostname][contains]=kitchen",
			want:        []Filter{and, property("hostname", "contains", "kitchen")},
		},
		{
			description: "each operator is passed through as named",
			query:       "filters[a][eq]=1&filters[b][ne]=2&filters[c][gt]=3&filters[d][lt]=4&filters[e][bool]=true",
			want: []Filter{
				and,
				property("a", "eq", "1"),
				property("b", "ne", "2"),
				property("c", "gt", "3"),
				property("d", "lt", "4"),
				property("e", "bool", "true"),
			},
		},
		{
			description: "parameters combine with and, in key order",
			query:       "filters[status][eq]=pending&filters[hostname][contains]=pi",
			want:        []Filter{and, property("hostname", "contains", "pi"), property("status", "eq", "pending")},
		},
		{
			description: "an encoded key and value are decoded",
			query:       "filters%5Bhostname%5D%5Bcontains%5D=pi%20kitchen%26co",
			want:        []Filter{and, property("hostname", "contains", "pi kitchen&co")},
		},
		{
			description: "an empty value is kept",
			query:       "filters[hostname][contains]=",
			want:        []Filter{and, property("hostname", "contains", "")},
		},
		{
			description: "a key without an operator is rejected",
			query:       "filters[hostname]=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a key without a field is rejected",
			query:       "filters[][contains]=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a key without an operator name is rejected",
			query:       "filters[hostname][]=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a key with a bracket inside the field is rejected",
			query:       "filters[host[name][contains]=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a key with a third segment is rejected",
			query:       "filters[hostname][contains][x]=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a key with trailing text is rejected",
			query:       "filters[hostname][contains]x=pi",
			wantErr:     ErrFilterInvalid,
		},
		{
			description: "a repeated key is rejected",
			query:       "filters[hostname][contains]=pi&filters[hostname][contains]=kitchen",
			wantErr:     ErrFilterInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			params, err := url.ParseQuery(tc.query)
			require.NoError(t, err)

			var filters Filters
			err = filters.UnmarshalQuery(params)

			require.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, filters.Data)
		})
	}
}
