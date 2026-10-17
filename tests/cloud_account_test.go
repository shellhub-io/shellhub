package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
)

// TestCloudAccountDeletion runs a cloud instance, the only edition where users delete their own
// account, for what the deletion does to the namespaces the user belongs to without owning them.
func TestCloudAccountDeletion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("deleting the account removes the member's paired devices", func(t *testing.T) {
		tenant := uuid.Generate()
		compose.NewUser(t, "account-owner", "account-owner"+stripeTestEmailDomain, ShellHubPassword)
		compose.NewNamespace(t, "account-owner", "account-space", tenant)

		compose.NewUser(t, "account-deleter", "account-deleter"+stripeTestEmailDomain, ShellHubPassword)
		compose.NewMember(t, "account-deleter", "account-space", string(authorizer.RoleOperator))
		member := signIntoTenant(t, compose, "account-deleter", ShellHubPassword, tenant)

		pairing := startPairing(t, compose, newPairingRequest(t, "gone-with-account", "02:00:00:00:a6:41"))
		paired := acceptPairingInto(t, asBearer(t, compose, member.Token), tenant, pairing.Code)
		require.Equal(t, member.ID, paired.OwnerID)

		resp, err := asBearer(t, compose, member.Token).Delete("/api/user")
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())

		compose.JWT(compose.AuthUser(t, "account-owner", ShellHubPassword).Token)

		requireRemovedDevice(t, compose, paired.UID)
	})
}
