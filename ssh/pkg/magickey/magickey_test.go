package magickey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReference(t *testing.T) {
	cases := []struct {
		description string
	}{
		{
			description: "succedds when function generate a valid key",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			pv := GetRerefence()
			assert.NotNil(t, pv)
		})
	}
}
