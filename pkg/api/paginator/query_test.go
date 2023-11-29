package paginator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		description string
		query       *Query
		expected    *Query
	}{
		{
			description: "set Page to MinParge when Page is lower than 1",
			query:       &Query{Page: -2, PerPage: 100},
			expected:    &Query{Page: 1, PerPage: 100},
		},
		{
			description: "set PerPage to MinPerParge when PerPage is lower than 1",
			query:       &Query{Page: 1, PerPage: -2},
			expected:    &Query{Page: 1, PerPage: 1},
		},
		{
			description: "set PerPage to MaxPerParge when PerPage is greather than 100",
			query:       &Query{Page: 1, PerPage: 101},
			expected:    &Query{Page: 1, PerPage: 100},
		},
		{
			description: "successfully parse query",
			query:       &Query{Page: 8, PerPage: 78},
			expected:    &Query{Page: 8, PerPage: 78},
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.query.Normalize()
			assert.Equal(t, tc.expected, tc.query)
		})
	}
}
