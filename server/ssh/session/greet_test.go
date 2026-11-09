package session

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestKeepsEarlyPayload(t *testing.T) {
	cases := []struct {
		description string
		device      *models.Device
		expected    bool
	}{
		{description: "no device", device: nil, expected: false},
		{description: "a device that reported no info", device: &models.Device{}, expected: false},
		{description: "the release that started keeping it", device: deviceAt("v0.27.0"), expected: true},
		{description: "a later release", device: deviceAt("v0.29.3"), expected: true},
		{description: "a development build", device: deviceAt("latest"), expected: true},
		{description: "a release candidate before it", device: deviceAt("v0.27.0-rc.11"), expected: false},
		{description: "an older release", device: deviceAt("v0.26.0"), expected: false},
		{description: "a version that does not parse", device: deviceAt("nightly"), expected: false},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, keepsEarlyPayload(tc.device))
		})
	}
}

func deviceAt(version string) *models.Device {
	return &models.Device{Info: &models.DeviceInfo{Version: version}}
}
