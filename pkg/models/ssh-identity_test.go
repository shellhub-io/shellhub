package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSSHIdentityActive(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		description string
		identity    SSHIdentity
		expected    bool
	}{
		{
			description: "a durable human identity is always active",
			identity:    SSHIdentity{},
			expected:    true,
		},
		{
			description: "an unexpired, unconsumed key is active",
			identity:    SSHIdentity{ExpiresAt: &future, SingleUse: true},
			expected:    true,
		},
		{
			description: "a consumed key is dead",
			identity:    SSHIdentity{ConsumedAt: &past},
			expected:    false,
		},
		{
			description: "an expired key is dead",
			identity:    SSHIdentity{ExpiresAt: &past},
			expected:    false,
		},
		{
			description: "a key expiring exactly now is dead",
			identity:    SSHIdentity{ExpiresAt: &now},
			expected:    false,
		},
		{
			description: "consumed wins even with a future expiry",
			identity:    SSHIdentity{ExpiresAt: &future, ConsumedAt: &past},
			expected:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.identity.Active(now))
		})
	}
}
