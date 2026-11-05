package gateway

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/stretchr/testify/assert"
)

func TestActorOfDropsAuthorization(t *testing.T) {
	caller := identity.Identity{
		ID:        "user-id",
		Username:  "username",
		TenantID:  "00000000-0000-4000-0000-000000000000",
		DeviceUID: "device-uid",
		APIKey:    "key",
		Role:      authorizer.RoleOwner,
		Admin:     true,
	}

	assert.Equal(t, Actor{
		ID:        "user-id",
		Username:  "username",
		APIKey:    "key",
		DeviceUID: "device-uid",
	}, actorOf(caller))
}

func TestActorOfAnAdminWithoutUserScope(t *testing.T) {
	caller := identity.Identity{ID: "admin-id", Admin: true}

	assert.Equal(t, Actor{ID: "admin-id"}, actorOf(caller.WithoutUserScope()))
}
