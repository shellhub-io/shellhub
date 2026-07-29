package storetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextDeviceIdentifiersAreDistinct(t *testing.T) {
	const calls = 1024

	for _, tc := range []struct {
		description string
		next        func() string
	}{
		{
			description: "UID",
			next:        nextDeviceUID,
		},
		{
			description: "MAC",
			next:        nextDeviceMAC,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			seen := make(map[string]int, calls)

			for i := range calls {
				got := tc.next()
				if first, dup := seen[got]; dup {
					t.Fatalf("call %d produced %q, already produced by call %d", i, got, first)
				}

				seen[got] = i
			}

			assert.Len(t, seen, calls)
		})
	}
}
