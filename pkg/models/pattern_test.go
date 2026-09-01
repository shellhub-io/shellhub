package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPattern(t *testing.T) {
	cases := []struct {
		description   string
		pattern       string
		value         string
		expectedMatch bool
		expectedErr   bool
	}{
		{
			description:   "an empty pattern imposes no restriction",
			pattern:       "",
			value:         "anything",
			expectedMatch: true,
		},
		{
			description:   "a literal pattern matches the whole value",
			pattern:       "staging",
			value:         "staging",
			expectedMatch: true,
		},
		{
			description:   "a literal pattern does not match a value that only contains it",
			pattern:       "staging",
			value:         "notstaging",
			expectedMatch: false,
		},
		{
			description:   "a literal pattern does not match a value it only prefixes",
			pattern:       "staging",
			value:         "staging-db",
			expectedMatch: false,
		},
		{
			description:   "a literal pattern does not match a value it sits inside",
			pattern:       "staging",
			value:         "prod-staging-mirror",
			expectedMatch: false,
		},
		{
			description:   "an alternation is anchored as a whole, not only its first branch",
			pattern:       "web|db",
			value:         "notdb",
			expectedMatch: false,
		},
		{
			description:   "an alternation still matches each of its branches in full",
			pattern:       "web|db",
			value:         "db",
			expectedMatch: true,
		},
		{
			description:   "a wildcard pattern still matches every value",
			pattern:       ".*",
			value:         "web-01",
			expectedMatch: true,
		},
		{
			description:   "a pattern the author already anchored keeps its meaning",
			pattern:       "^web-[0-9]+$",
			value:         "web-01",
			expectedMatch: true,
		},
		{
			description:   "a multiline anchor cannot match one line of a multiline value",
			pattern:       "(?m)^root$",
			value:         "root\nevil",
			expectedMatch: false,
		},
		{
			description:   "an inline flag group keeps applying to the pattern",
			pattern:       "(?i)web-01",
			value:         "WEB-01",
			expectedMatch: true,
		},
		{
			description:   "a pattern that does not compile returns an error",
			pattern:       "[",
			value:         "web-01",
			expectedMatch: false,
			expectedErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			matched, err := MatchPattern(tc.pattern, tc.value)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedMatch, matched)
		})
	}
}
