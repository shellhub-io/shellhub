package paginator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	assert.Equal(t, &Query{1, 25}, NewQuery())
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		description string
		query       *Query
		expected    *Query
	}{
		{description: "Failed when page is lower then 0", query: &Query{Page: -1, PerPage: 25}, expected: &Query{Page: 1, PerPage: 25}},
		{description: "Failed when page is lower than 1 and per page is greater then 100", query: &Query{Page: -1, PerPage: 101}, expected: &Query{Page: 1, PerPage: 100}},
		{description: "Failed when page is lower then -1", query: &Query{Page: -2, PerPage: 100}, expected: &Query{Page: 1, PerPage: 100}},
		{description: "Failed when per page is greater then 100", query: &Query{Page: 1, PerPage: 101}, expected: &Query{Page: 1, PerPage: 100}},
		{description: "Failed when per page is lower than 0", query: &Query{Page: 1, PerPage: -1}, expected: &Query{Page: 1, PerPage: 1}},
		{description: "Failed when per page is lower than -1", query: &Query{Page: 1, PerPage: -2}, expected: &Query{Page: 1, PerPage: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.query.Normalize()
			assert.Equal(t, tc.expected, tc.query)
		})
	}
}
