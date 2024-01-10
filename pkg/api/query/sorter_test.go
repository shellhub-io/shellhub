package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortNormalize(t *testing.T) {
	cases := []struct {
		description string
		order       *Sorter
		expected    *Sorter
	}{
		{
			description: "sets By to desc when old By is invalid",
			order:       &Sorter{By: "date", Order: "foo"},
			expected:    &Sorter{By: "date", Order: "desc"},
		},
		{
			description: "successfully parse query",
			order:       &Sorter{By: "date", Order: "asc"},
			expected:    &Sorter{By: "date", Order: "asc"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.order.Normalize()
			assert.Equal(t, tc.expected, tc.order)
		})
	}
}
