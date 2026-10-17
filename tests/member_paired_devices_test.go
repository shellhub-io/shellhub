package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPairedDevicesFollowTheirMember covers the devices a member pairs: they belong to that member,
// and they leave the namespace with them, when the member is demoted below accepting devices,
// removed, or leaves. The owner may keep some of them as team devices on a demotion or a removal.
// The cases share one stack, so each signs in its own member and pairs devices under hostnames and
// MAC addresses no other case uses.
func TestPairedDevicesFollowTheirMember(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), newNamespace)

	t.Run("accepting a pairing makes the accepting member the device's owner", func(t *testing.T) {
		member := signInPairingMember(t, compose, "pairing-owner")

		accepted := pairDeviceAs(t, compose, member, "owned-by-acceptor", "02:00:00:00:a6:01")

		assert.Equal(t, member.ID, accepted.OwnerID)
		assert.Equal(t, member.ID, requireDevice(t, compose, accepted.UID).OwnerID)
	})

	t.Run("demoting a member to observer removes their paired devices but the ones kept as team devices", func(t *testing.T) {
		member := signInPairingMember(t, compose, "demoted-pairer")
		kept := pairDeviceAs(t, compose, member, "kept-on-demotion", "02:00:00:00:a6:11")
		gone := pairDeviceAs(t, compose, member, "gone-on-demotion", "02:00:00:00:a6:12")

		resp, err := compose.R(t.Context()).
			SetBody(map[string]any{"role": authorizer.RoleObserver, "keep_devices": []string{kept.UID}}).
			Patch("/api/namespaces/" + ShellHubNamespace + "/members/" + member.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		requireTeamDevice(t, compose, kept.UID)
		requireRemovedDevice(t, compose, gone.UID)
	})

	t.Run("removing a member removes their paired devices but the ones kept as team devices", func(t *testing.T) {
		member := signInPairingMember(t, compose, "removed-pairer")
		kept := pairDeviceAs(t, compose, member, "kept-on-removal", "02:00:00:00:a6:21")
		gone := pairDeviceAs(t, compose, member, "gone-on-removal", "02:00:00:00:a6:22")

		resp, err := compose.R(t.Context()).
			SetQueryParam("keep_devices", kept.UID).
			Delete("/api/namespaces/" + ShellHubNamespace + "/members/" + member.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		requireTeamDevice(t, compose, kept.UID)
		requireRemovedDevice(t, compose, gone.UID)
	})

	t.Run("leaving the namespace removes the member's paired devices", func(t *testing.T) {
		member := signInPairingMember(t, compose, "leaving-pairer")
		paired := pairDeviceAs(t, compose, member, "gone-on-leave", "02:00:00:00:a6:31")

		resp, err := asBearer(t, compose, member.Token).Delete("/api/namespaces/" + ShellHubNamespace + "/members")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		requireRemovedDevice(t, compose, paired.UID)
	})
}

func signInPairingMember(t *testing.T, compose *environment.DockerCompose, username string) *models.UserAuthResponse {
	t.Helper()

	compose.NewUser(t, username, username+"@ossystems.com.br", ShellHubPassword)
	compose.NewMember(t, username, ShellHubNamespaceName, string(authorizer.RoleOperator))

	return signInto(t, compose, username, ShellHubPassword)
}

func pairDeviceAs(t *testing.T, compose *environment.DockerCompose, member *models.UserAuthResponse, hostname, mac string) *models.DevicePairingAccepted {
	t.Helper()

	pairing := startPairing(t, compose, newPairingRequest(t, hostname, mac))

	return acceptPairing(t, asBearer(t, compose, member.Token), pairing.Code)
}

func requireTeamDevice(t *testing.T, compose *environment.DockerCompose, uid string) {
	t.Helper()

	device := requireDevice(t, compose, uid)
	assert.Equal(t, models.DeviceStatusAccepted, device.Status, "a device kept as a team device left the namespace")
	assert.Empty(t, device.OwnerID, "a device kept as a team device still has an owner")
}

func requireRemovedDevice(t *testing.T, compose *environment.DockerCompose, uid string) {
	t.Helper()

	device := requireDevice(t, compose, uid)
	assert.Equal(t, models.DeviceStatusRemoved, device.Status, "a paired device stayed after its member lost it")
	assert.Empty(t, device.OwnerID, "a removed device still has an owner")
}
