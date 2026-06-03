package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func fingerprintOf(keys ...string) eventFingerprint {
	var f eventFingerprint
	for _, k := range keys {
		f.add(k)
	}

	return f
}

func TestEventFingerprint(t *testing.T) {
	cases := []struct {
		description string
		a           eventFingerprint
		b           eventFingerprint
		equal       bool
	}{
		{
			description: "equal regardless of order",
			a:           fingerprintOf("a", "b", "c"),
			b:           fingerprintOf("c", "a", "b"),
			equal:       true,
		},
		{
			description: "equal for the same multiset with duplicates",
			a:           fingerprintOf("a", "a", "b"),
			b:           fingerprintOf("b", "a", "a"),
			equal:       true,
		},
		{
			description: "differs when an event is missing",
			a:           fingerprintOf("a", "b", "c"),
			b:           fingerprintOf("a", "b"),
			equal:       false,
		},
		{
			description: "differs on duplicate count",
			a:           fingerprintOf("a", "a", "b"),
			b:           fingerprintOf("a", "b"),
			equal:       false,
		},
		{
			description: "differs when an event payload changes",
			a:           fingerprintOf("a", "b", "c"),
			b:           fingerprintOf("a", "b", "x"),
			equal:       false,
		},
		{
			description: "empty fingerprints are equal",
			a:           fingerprintOf(),
			b:           fingerprintOf(),
			equal:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.equal, tc.a.equal(tc.b))
		})
	}
}

func TestEventFingerprintCount(t *testing.T) {
	f := fingerprintOf("a", "b", "c")
	assert.Equal(t, int64(3), f.count)
}
