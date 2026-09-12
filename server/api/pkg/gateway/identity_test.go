package gateway

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentityRoundTrip is what keeps [Identity.WriteTo] and [IdentityFrom] naming the same headers.
// A field added to one and forgotten in the other survives review and the compiler, and shows up
// only as an identity that silently loses part of itself between the authenticator and the handler.
func TestIdentityRoundTrip(t *testing.T) {
	cases := []struct {
		description string
		identity    Identity
	}{
		{
			description: "a user token",
			identity: Identity{
				ID:       "user-id",
				Username: "username",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Role:     authorizer.RoleOwner,
			},
		},
		{
			description: "an api key",
			identity: Identity{
				TenantID: "00000000-0000-4000-0000-000000000000",
				APIKey:   "key",
				Role:     authorizer.RoleObserver,
			},
		},
		{
			description: "a device token",
			identity: Identity{
				DeviceUID: "device-uid",
				TenantID:  "00000000-0000-4000-0000-000000000000",
			},
		},
		{
			description: "an admin browsing the admin console",
			identity: Identity{
				Username: "username",
				Role:     authorizer.RoleOwner,
				Admin:    true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			header := http.Header{}
			tc.identity.WriteTo(header)

			require.Equal(tt, tc.identity, IdentityFrom(header))
		})
	}
}

// TestIdentityActorDropsAuthorization pins what an actor deliberately is not. Role and admin decide
// what the caller may do, which the middleware owns; handing them to a handler invites it to make
// that decision a second time.
func TestIdentityActorDropsAuthorization(t *testing.T) {
	identity := Identity{
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
	}, identity.Actor())
}
