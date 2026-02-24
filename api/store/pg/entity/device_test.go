package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceFromModel(t *testing.T) {
	now := time.Now()
	disconnectedAt := now.Add(-time.Hour)

	tests := []struct {
		name  string
		model *models.Device
		check func(t *testing.T, result *Device)
	}{
		{
			name: "full fields",
			model: &models.Device{
				UID:            "device-uid-1",
				TenantID:       "tenant-id-1",
				CreatedAt:      now,
				LastSeen:       now,
				Status:         models.DeviceStatusAccepted,
				Name:           "my-device",
				PublicKey:      "ssh-rsa AAAA...",
				DisconnectedAt: &disconnectedAt,
				Identity:       &models.DeviceIdentity{MAC: "00:11:22:33:44:55"},
				Position:       &models.DevicePosition{Longitude: 1.23, Latitude: 4.56},
				Info: &models.DeviceInfo{
					ID:         "device-info-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "amd64",
					Platform:   "linux",
				},
			},
			check: func(t *testing.T, result *Device) {
				assert.Equal(t, "device-uid-1", result.ID)
				assert.Equal(t, "tenant-id-1", result.NamespaceID)
				assert.Equal(t, "accepted", result.Status)
				assert.Equal(t, "my-device", result.Name)
				assert.Equal(t, "ssh-rsa AAAA...", result.PublicKey)
				assert.Equal(t, disconnectedAt, result.DisconnectedAt)
				assert.Equal(t, "00:11:22:33:44:55", result.MAC)
				assert.InDelta(t, 1.23, result.Longitude, 0.001)
				assert.InDelta(t, 4.56, result.Latitude, 0.001)
				assert.Equal(t, "device-info-id", result.Identifier)
				assert.Equal(t, "My Device", result.PrettyName)
				assert.Equal(t, "1.0.0", result.Version)
				assert.Equal(t, "amd64", result.Arch)
				assert.Equal(t, "linux", result.Platform)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, now, result.LastSeen)
				assert.True(t, result.UpdatedAt.IsZero())
			},
		},
		{
			name: "empty Status defaults to pending",
			model: &models.Device{
				UID:    "device-uid-2",
				Status: "",
			},
			check: func(t *testing.T, result *Device) {
				assert.Equal(t, "pending", result.Status)
			},
		},
		{
			name: "nil Identity, Position, Info, DisconnectedAt",
			model: &models.Device{
				UID:            "device-uid-3",
				Status:         models.DeviceStatusPending,
				Identity:       nil,
				Position:       nil,
				Info:           nil,
				DisconnectedAt: nil,
			},
			check: func(t *testing.T, result *Device) {
				assert.Equal(t, "", result.MAC)
				assert.InDelta(t, 0.0, result.Longitude, 0.001)
				assert.InDelta(t, 0.0, result.Latitude, 0.001)
				assert.Equal(t, "", result.Identifier)
				assert.Equal(t, "", result.PrettyName)
				assert.True(t, result.DisconnectedAt.IsZero())
			},
		},
		{
			name: "Tags from model.Tags (full)",
			model: &models.Device{
				UID:    "device-uid-4",
				Status: models.DeviceStatusAccepted,
				Taggable: models.Taggable{
					Tags: []models.Tag{
						{ID: "tag-1", Name: "prod", TenantID: "t1"},
						{ID: "tag-2", Name: "staging", TenantID: "t1"},
					},
				},
			},
			check: func(t *testing.T, result *Device) {
				require.Len(t, result.Tags, 2)
				assert.Equal(t, "tag-1", result.Tags[0].ID)
				assert.Equal(t, "prod", result.Tags[0].Name)
				assert.Equal(t, "tag-2", result.Tags[1].ID)
			},
		},
		{
			name: "Tags from model.TagIDs (IDs only)",
			model: &models.Device{
				UID:    "device-uid-5",
				Status: models.DeviceStatusAccepted,
				Taggable: models.Taggable{
					TagIDs: []string{"tag-1", "tag-2"},
				},
			},
			check: func(t *testing.T, result *Device) {
				require.Len(t, result.Tags, 2)
				assert.Equal(t, "tag-1", result.Tags[0].ID)
				assert.Equal(t, "tag-2", result.Tags[1].ID)
				// Only ID is set when using TagIDs
				assert.Equal(t, "", result.Tags[0].Name)
			},
		},
		{
			name: "no tags",
			model: &models.Device{
				UID:    "device-uid-6",
				Status: models.DeviceStatusAccepted,
			},
			check: func(t *testing.T, result *Device) {
				assert.Empty(t, result.Tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeviceFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestDeviceToModel(t *testing.T) {
	now := time.Now()
	disconnectedAt := now.Add(-time.Hour)

	tests := []struct {
		name   string
		entity *Device
		check  func(t *testing.T, result *models.Device)
	}{
		{
			name: "full fields with Namespace loaded",
			entity: &Device{
				ID:             "device-uid-1",
				NamespaceID:    "tenant-id-1",
				CreatedAt:      now,
				LastSeen:       now,
				Status:         "accepted",
				Name:           "my-device",
				PublicKey:      "ssh-rsa AAAA...",
				Online:         true,
				Acceptable:     false,
				DisconnectedAt: disconnectedAt,
				MAC:            "00:11:22:33:44:55",
				Longitude:      1.23,
				Latitude:       4.56,
				Identifier:     "info-id",
				PrettyName:     "My Device",
				Version:        "1.0.0",
				Arch:           "amd64",
				Platform:       "linux",
				Namespace:      &Namespace{Name: "my-namespace"},
				Tags: []*Tag{
					{ID: "tag-1", NamespaceID: "t1", Name: "prod"},
				},
			},
			check: func(t *testing.T, result *models.Device) {
				assert.Equal(t, "device-uid-1", result.UID)
				assert.Equal(t, "tenant-id-1", result.TenantID)
				assert.Equal(t, models.DeviceStatusAccepted, result.Status)
				assert.Equal(t, "my-device", result.Name)
				assert.Equal(t, "ssh-rsa AAAA...", result.PublicKey)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, now, result.LastSeen)
				assert.Equal(t, "my-namespace", result.Namespace)
				assert.True(t, result.Online)
				assert.False(t, result.Acceptable)
				require.NotNil(t, result.DisconnectedAt)
				assert.Equal(t, disconnectedAt, *result.DisconnectedAt)
				require.NotNil(t, result.Position)
				assert.InDelta(t, 1.23, result.Position.Longitude, 0.001)
				assert.InDelta(t, 4.56, result.Position.Latitude, 0.001)
				require.NotNil(t, result.Info)
				assert.Equal(t, "info-id", result.Info.ID)
				assert.Equal(t, "My Device", result.Info.PrettyName)
				assert.Equal(t, "1.0.0", result.Info.Version)
				assert.Equal(t, "amd64", result.Info.Arch)
				assert.Equal(t, "linux", result.Info.Platform)
				require.NotNil(t, result.Identity)
				assert.Equal(t, "00:11:22:33:44:55", result.Identity.MAC)
				require.Len(t, result.Tags, 1)
				assert.Equal(t, "tag-1", result.Tags[0].ID)
				require.Len(t, result.TagIDs, 1)
				assert.Equal(t, "tag-1", result.TagIDs[0])
			},
		},
		{
			name: "nil Namespace regression - must not panic",
			entity: &Device{
				ID:        "device-uid-2",
				Status:    "pending",
				Namespace: nil,
			},
			check: func(t *testing.T, result *models.Device) {
				assert.Equal(t, "", result.Namespace)
			},
		},
		{
			name: "zero DisconnectedAt",
			entity: &Device{
				ID:             "device-uid-3",
				Status:         "accepted",
				DisconnectedAt: time.Time{},
			},
			check: func(t *testing.T, result *models.Device) {
				assert.Nil(t, result.DisconnectedAt)
			},
		},
		{
			name: "non-zero DisconnectedAt",
			entity: &Device{
				ID:             "device-uid-4",
				Status:         "accepted",
				DisconnectedAt: disconnectedAt,
			},
			check: func(t *testing.T, result *models.Device) {
				require.NotNil(t, result.DisconnectedAt)
				assert.Equal(t, disconnectedAt, *result.DisconnectedAt)
			},
		},
		{
			name: "without Tags",
			entity: &Device{
				ID:     "device-uid-5",
				Status: "accepted",
				Tags:   []*Tag{},
			},
			check: func(t *testing.T, result *models.Device) {
				assert.Empty(t, result.Tags)
				assert.Nil(t, result.TagIDs)
			},
		},
		{
			name: "nil Tags",
			entity: &Device{
				ID:     "device-uid-6",
				Status: "accepted",
				Tags:   nil,
			},
			check: func(t *testing.T, result *models.Device) {
				assert.Empty(t, result.Tags)
				assert.Nil(t, result.TagIDs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeviceToModel(tt.entity)
			tt.check(t, result)
		})
	}
}
