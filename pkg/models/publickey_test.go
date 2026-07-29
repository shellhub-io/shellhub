package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyFilterMatches(t *testing.T) {
	cases := []struct {
		description   string
		filter        PublicKeyFilter
		device        *Device
		expectedMatch bool
		expectedErr   bool
	}{
		{
			description:   "empty filter matches any device",
			filter:        PublicKeyFilter{},
			device:        &Device{Name: "any-device"},
			expectedMatch: true,
		},
		{
			description:   "hostname regexp matches the device name",
			filter:        PublicKeyFilter{Hostname: ".*"},
			device:        &Device{Name: "web-01"},
			expectedMatch: true,
		},
		{
			description:   "hostname regexp does not match the device name",
			filter:        PublicKeyFilter{Hostname: "^db-"},
			device:        &Device{Name: "web-01"},
			expectedMatch: false,
		},
		{
			description:   "invalid hostname regexp returns an error",
			filter:        PublicKeyFilter{Hostname: "["},
			device:        &Device{Name: "web-01"},
			expectedMatch: false,
			expectedErr:   true,
		},
		{
			description:   "tag filter matches when the device shares a tag",
			filter:        PublicKeyFilter{Taggable: Taggable{TagIDs: []string{"tag1", "tag2"}}},
			device:        &Device{Taggable: Taggable{TagIDs: []string{"tag2"}}},
			expectedMatch: true,
		},
		{
			description:   "tag filter does not match when the device shares no tag",
			filter:        PublicKeyFilter{Taggable: Taggable{TagIDs: []string{"tag1"}}},
			device:        &Device{Taggable: Taggable{TagIDs: []string{"tag3"}}},
			expectedMatch: false,
		},
		{
			description:   "tag filter does not match a device with no tags",
			filter:        PublicKeyFilter{Taggable: Taggable{TagIDs: []string{"tag1"}}},
			device:        &Device{},
			expectedMatch: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			matched, err := tc.filter.Matches(tc.device)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expectedMatch, matched)
		})
	}
}
