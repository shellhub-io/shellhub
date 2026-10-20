package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceJSONKeyVerified(t *testing.T) {
	verifiedAt := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description string
		device      Device
		expected    bool
	}{
		{
			description: "a device that never signed",
			device:      Device{UID: "uid"},
			expected:    false,
		},
		{
			description: "a device that signed",
			device:      Device{UID: "uid", KeyVerifiedAt: &verifiedAt},
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			data, err := json.Marshal(tc.device)
			require.NoError(t, err)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(data, &decoded))

			assert.Equal(t, tc.expected, decoded["key_verified"])
			assert.Equal(t, "uid", decoded["uid"])
		})
	}
}
