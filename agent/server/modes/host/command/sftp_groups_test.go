package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSFTPGroups_RoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		groups   []uint32
		encoded  string
		expected []int
	}{
		{name: "no groups", groups: nil, encoded: "", expected: []int{}},
		{name: "empty groups", groups: []uint32{}, encoded: "", expected: []int{}},
		{name: "one group", groups: []uint32{27}, encoded: "27", expected: []int{27}},
		{name: "several groups", groups: []uint32{4, 27, 1000}, encoded: "4,27,1000", expected: []int{4, 27, 1000}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := FormatSFTPGroups(tc.groups)
			assert.Equal(t, tc.encoded, encoded)

			parsed, err := ParseSFTPGroups(encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, parsed)
		})
	}
}

func TestParseSFTPGroups_RefusesMalformedLists(t *testing.T) {
	for _, value := range []string{"wheel", "4,,27", "4,", "-1", "4294967296", " 4"} {
		t.Run(value, func(t *testing.T) {
			_, err := ParseSFTPGroups(value)
			assert.Error(t, err)
		})
	}
}
