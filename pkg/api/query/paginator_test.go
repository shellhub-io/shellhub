package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginatorNormalize(t *testing.T) {
	cases := []struct {
		description string
		paginator   *Paginator
		expected    *Paginator
	}{
		{
			description: "set Page to MinParge when Page is lower than 1",
			paginator:   &Paginator{Page: -2, PerPage: 100},
			expected:    &Paginator{Page: 1, PerPage: 100},
		},
		{
			description: "set PerPage to MinPerParge when PerPage is lower than 1",
			paginator:   &Paginator{Page: 1, PerPage: -2},
			expected:    &Paginator{Page: 1, PerPage: 1},
		},
		{
			description: "set PerPage to MaxPerParge when PerPage is greather than 100",
			paginator:   &Paginator{Page: 1, PerPage: 101},
			expected:    &Paginator{Page: 1, PerPage: 100},
		},
		{
			description: "successfully parse query",
			paginator:   &Paginator{Page: 8, PerPage: 78},
			expected:    &Paginator{Page: 8, PerPage: 78},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.paginator.Normalize()
			assert.Equal(t, tc.expected, tc.paginator)
		})
	}
}
