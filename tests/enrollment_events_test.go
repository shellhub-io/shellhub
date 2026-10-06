package main

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	eventedPublicKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAwb2v98Gh0bKThSWcNBP/l2o7VQzWlCffrXFV5S8KUllTufZVmf1t
30gLBkW40scbfYHzgvvBTh/skPJLQNyRCu3ks+mSywT+dn5zES78i7uyMNPxzVso
tLNh+5DfUlyzhE+rt/eE9B44gf3OOJp/dibxpqLd4jVFIGsqFU9UAC+6miRCDS7o
i97/IevRDfkIfTBYnkw58d6cbCrniHxgzBi1AzWNdPl7JxKVc4VvfaqJJdaEvOFb
b19C8A/sfVFzf3HzUK+JPgoObr8e4EnAKRMwnCrG6Qla3HzYgvPeTPd7SDnUAQLX
qBXQ/u3v3+yI8LXihvNEvuaI1JW8XVHh7QIDAQAB
-----END RSA PUBLIC KEY-----
`
	eventedPublicKeyFingerprint = "SHA256:pPJ4qUrhmqHC5+kSyI7rWsiwXnMllKxi1zEobSeqnXY"
)

func testEnrollmentEvents(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
		Name: "evented",
		Mode: string(models.ProvisioningKeyModeManual),
	})

	t.Run("an enrollment records the device as it enrolled", func(t *testing.T) {
		req := newKeyedDeviceAuthRequest(t, key.Key, "evented", "02:00:00:00:70:01")
		req.PublicKey = eventedPublicKey
		before := clock.Now()

		device := enroll(t, compose, req)

		event := eventOf(compose.ProvisioningKeyHistory(t, key.ID), device.UID)
		require.NotNil(t, event, "the enrollment recorded no event")

		assert.NotEmpty(t, event.ID)
		assert.Equal(t, key.ID, event.ProvisioningKeyID)
		assert.Equal(t, ShellHubNamespace, event.TenantID)
		assert.Equal(t, "evented", event.Hostname)
		assert.Equal(t, "02:00:00:00:70:01", event.Identity)
		assert.Equal(t, &models.DeviceInfo{
			ID:         req.Info.ID,
			PrettyName: req.Info.PrettyName,
			Version:    req.Info.Version,
			Arch:       req.Info.Arch,
			Platform:   req.Info.Platform,
		}, event.Info)
		assert.NotEmpty(t, event.SourceIP)
		assert.Equal(t, eventedPublicKey, event.PublicKey)
		assert.Equal(t, eventedPublicKeyFingerprint, event.Fingerprint)
		assert.False(t, event.Ephemeral)
		assert.False(t, event.ReRegistration)
		assert.WithinRange(t, event.Timestamp, before.Add(-time.Minute), clock.Now().Add(time.Minute))
		assert.Equal(t, models.DeviceStatusPending, event.DeviceStatus)
		assert.Empty(t, event.DecidedStatus)
		assert.Nil(t, event.DecidedAt)
		assert.True(t, event.IsCurrent)
	})

	cases := []struct {
		description string
		hostname    string
		mac         string
		decision    environment.DeviceStatusAction
		status      models.DeviceStatus
	}{
		{
			description: "accepting the device stamps the event accepted",
			hostname:    "evented-accepted",
			mac:         "02:00:00:00:70:02",
			decision:    environment.DeviceActionAccept,
			status:      models.DeviceStatusAccepted,
		},
		{
			description: "rejecting the device stamps the event rejected",
			hostname:    "evented-rejected",
			mac:         "02:00:00:00:70:03",
			decision:    environment.DeviceActionReject,
			status:      models.DeviceStatusRejected,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, tc.hostname, tc.mac))
			before := clock.Now()

			compose.UpdateDeviceStatus(t, device.UID, tc.decision)

			event := eventOf(compose.ProvisioningKeyHistory(t, key.ID), device.UID)
			require.NotNil(t, event, "the enrollment recorded no event")
			assert.Equal(t, tc.status, event.DecidedStatus)
			if assert.NotNil(t, event.DecidedAt) {
				assert.WithinRange(t, *event.DecidedAt, before.Add(-time.Minute), clock.Now().Add(time.Minute))
			}
		})
	}
}
