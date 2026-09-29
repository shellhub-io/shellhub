package main

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutesThatRefuseAPIKeys(t *testing.T) {
	compose := environment.New(t).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")

	auth := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)
	compose.JWT(auth.Token)

	key := compose.CreateAPIKey(t, &requests.CreateAPIKey{
		Name:      "automation",
		ExpiresAt: -1,
		OptRole:   authorizer.RoleAdministrator,
	})

	withKey := func(t *testing.T) *resty.Request {
		t.Helper()

		return compose.Anonymous(t.Context()).SetHeader("X-API-Key", key.Key)
	}

	t.Run("the key authenticates on a namespace route", func(t *testing.T) {
		resp, err := withKey(t).Get("/api/devices")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	cases := []struct {
		description string
		method      string
		path        string
	}{
		{
			description: "switching to a namespace token",
			method:      http.MethodGet,
			path:        "/api/auth/token/" + ShellHubNamespace,
		},
		{
			description: "listing the namespaces",
			method:      http.MethodGet,
			path:        "/api/namespaces",
		},
		{
			description: "updating the user",
			method:      http.MethodPatch,
			path:        "/api/users",
		},
		{
			description: "updating the user through the deprecated route",
			method:      http.MethodPatch,
			path:        "/api/users/" + auth.ID + "/data",
		},
		{
			description: "changing the password through the deprecated route",
			method:      http.MethodPatch,
			path:        "/api/users/" + auth.ID + "/password",
		},
		{
			description: "listing the API keys",
			method:      http.MethodGet,
			path:        "/api/namespaces/api-key",
		},
		{
			description: "creating an API key",
			method:      http.MethodPost,
			path:        "/api/namespaces/api-key",
		},
		{
			description: "creating a provisioning key",
			method:      http.MethodPost,
			path:        "/api/namespaces/provisioning-key",
		},
		{
			description: "listing the provisioning keys",
			method:      http.MethodGet,
			path:        "/api/namespaces/provisioning-key",
		},
		{
			description: "updating a provisioning key",
			method:      http.MethodPatch,
			path:        "/api/namespaces/provisioning-key/any",
		},
		{
			description: "revealing a provisioning key",
			method:      http.MethodGet,
			path:        "/api/namespaces/provisioning-key/any/reveal",
		},
		{
			description: "reading a provisioning key's history",
			method:      http.MethodGet,
			path:        "/api/namespaces/provisioning-key/any/history",
		},
		{
			description: "generating an invitation link",
			method:      http.MethodPost,
			path:        "/api/namespaces/" + ShellHubNamespace + "/invitations/links",
		},
		{
			description: "accepting an invitation",
			method:      http.MethodPatch,
			path:        "/api/namespaces/" + ShellHubNamespace + "/invitations/accept",
		},
		{
			description: "resolving a device login code",
			method:      http.MethodGet,
			path:        "/api/devices/login-code/any",
		},
		{
			description: "accepting a device pairing",
			method:      http.MethodPost,
			path:        "/api/devices/pairing/any/accept",
		},
		{
			description: "reading an SSH approval",
			method:      http.MethodGet,
			path:        "/api/ssh-approvals/any",
		},
		{
			description: "confirming an SSH approval",
			method:      http.MethodPost,
			path:        "/api/ssh-approvals/any/confirm",
		},
		{
			description: "rejecting an SSH approval",
			method:      http.MethodPost,
			path:        "/api/ssh-approvals/any/reject",
		},
		{
			description: "renaming an API key",
			method:      http.MethodPatch,
			path:        "/api/namespaces/api-key/" + key.Name,
		},
		{
			description: "deleting an API key",
			method:      http.MethodDelete,
			path:        "/api/namespaces/api-key/" + key.Name,
		},
		{
			description: "leaving the namespace",
			method:      http.MethodDelete,
			path:        "/api/namespaces/" + ShellHubNamespace + "/members",
		},
	}

	for _, tc := range cases {
		t.Run("refuses "+tc.description, func(t *testing.T) {
			resp, err := withKey(t).Execute(tc.method, tc.path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		})
	}
}
