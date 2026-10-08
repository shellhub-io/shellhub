package main

import (
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutesThatRefuseAPIKeys(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
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

	t.Run("the key authenticates on a namespace route", func(t *testing.T) {
		resp, err := withAPIKey(t, compose, key.Key).Get("/api/devices")
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
			resp, err := withAPIKey(t, compose, key.Key).Execute(tc.method, tc.path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		})
	}
}

// TestNamespaceAPIKeyAuthentication covers what a namespace API key can do once minted: manage the
// namespace's tags, act only within its role, keep working when it never expires, and stop working
// once its expiry passes. It also covers a key carrying the instance key prefix, which the server
// honours only on the admin API, being refused on a namespace route. The routes that refuse a key whatever its role are covered by
// [TestRoutesThatRefuseAPIKeys].
func TestNamespaceAPIKeyAuthentication(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")

	compose.JWT(compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token)

	_, device := startAcceptedAgent(t, t.Context(), compose)

	deviceTags := func(t *testing.T) []string {
		t.Helper()

		current, resp, err := compose.GetDevice(t.Context(), device.UID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return tagNames(current.Tags)
	}

	namespaceTags := func(t *testing.T, req *resty.Request) []string {
		t.Helper()

		tags := []models.Tag{}
		resp, err := req.SetResult(&tags).Get("/api/tags")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return tagNames(tags)
	}

	t.Run("a key creates, lists, attaches, renames, detaches and deletes tags", func(t *testing.T) {
		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "tagger", ExpiresAt: -1, OptRole: authorizer.RoleOperator})

		resp, err := withAPIKey(t, compose, key.Key).SetBody(map[string]string{"name": "staging"}).Post("/api/tags")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.Contains(t, namespaceTags(t, withAPIKey(t, compose, key.Key)), "staging")

		resp, err = withAPIKey(t, compose, key.Key).Post("/api/devices/" + device.UID + "/tags/staging")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, []string{"staging"}, deviceTags(t))

		renamed := new(models.Tag)
		resp, err = withAPIKey(t, compose, key.Key).
			SetBody(map[string]string{"name": "production"}).
			SetResult(renamed).
			Patch("/api/tags/staging")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, "production", renamed.Name)
		assert.Equal(t, ShellHubNamespace, renamed.TenantID)
		assert.True(t, renamed.UpdatedAt.After(renamed.CreatedAt))

		tagsAfterRename := namespaceTags(t, withAPIKey(t, compose, key.Key))
		assert.Contains(t, tagsAfterRename, "production")
		assert.NotContains(t, tagsAfterRename, "staging")
		assert.Equal(t, []string{"production"}, deviceTags(t))

		resp, err = withAPIKey(t, compose, key.Key).Delete("/api/devices/" + device.UID + "/tags/production")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Empty(t, deviceTags(t))

		resp, err = withAPIKey(t, compose, key.Key).Delete("/api/tags/production")
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())

		assert.NotContains(t, namespaceTags(t, compose.R(t.Context())), "production")
	})

	t.Run("a key acts only within its role", func(t *testing.T) {
		listDevices := func(req *resty.Request) (*resty.Response, error) { return req.Get("/api/devices") }
		createTag := func(req *resty.Request) (*resty.Response, error) {
			return req.SetBody(map[string]string{"name": "role" + uuid.Generate()[:8]}).Post("/api/tags")
		}
		listAccessPolicies := func(req *resty.Request) (*resty.Response, error) { return req.Get("/api/access-policies") }

		cases := []struct {
			role                    authorizer.Role
			devices, tags, policies int
		}{
			{role: authorizer.RoleObserver, devices: http.StatusOK, tags: http.StatusForbidden, policies: http.StatusForbidden},
			{role: authorizer.RoleOperator, devices: http.StatusOK, tags: http.StatusOK, policies: http.StatusForbidden},
			{role: authorizer.RoleAdministrator, devices: http.StatusOK, tags: http.StatusOK, policies: http.StatusOK},
		}

		for _, tc := range cases {
			t.Run(tc.role.String(), func(t *testing.T) {
				key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: tc.role.String(), ExpiresAt: -1, OptRole: tc.role})
				require.Equal(t, tc.role, key.Role)

				for _, check := range []struct {
					request func(*resty.Request) (*resty.Response, error)
					want    int
				}{
					{request: listDevices, want: tc.devices},
					{request: createTag, want: tc.tags},
					{request: listAccessPolicies, want: tc.policies},
				} {
					resp, err := check.request(withAPIKey(t, compose, key.Key))
					require.NoError(t, err)
					assert.Equal(t, check.want, resp.StatusCode(), "%s %s: %s", resp.Request.Method, resp.Request.URL, resp.String())
				}
			})
		}
	})

	t.Run("a key that never expires is stored without an expiry and authenticates", func(t *testing.T) {
		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "forever", ExpiresAt: -1})
		require.Equal(t, int64(-1), key.ExpiresIn)

		assert.Equal(t, int64(-1), listedAPIKey(t, compose, "forever").ExpiresIn,
			"-1 is the marker for a key that never expires, not a time")

		resp, err := withAPIKey(t, compose, key.Key).Get("/api/devices")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	t.Run("a key carrying the instance key prefix is refused on a namespace route", func(t *testing.T) {
		resp, err := withAPIKey(t, compose, models.InstanceAPIKeyPrefix+uuid.Generate()).Get("/api/devices")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
	})

	t.Run("a key is refused once its expiry passes", func(t *testing.T) {
		const ttl = 10 * time.Second

		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "expiring", ExpiresAt: 30})
		compose.ExpireAPIKeyIn(t, key.Name, ttl)

		resp, err := withAPIKey(t, compose, key.Key).Get("/api/devices")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := withAPIKey(t, compose, key.Key).Get("/api/devices")
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, http.StatusUnauthorized, resp.StatusCode(), resp.String())
		}, 3*ttl, time.Second)
	})
}

func TestAPIKeyRoleCap(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")

	compose.JWT(compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token)

	requestKey := func(t *testing.T, req *resty.Request, name string, role authorizer.Role) *resty.Response {
		t.Helper()

		resp, err := req.SetBody(&requests.CreateAPIKey{Name: name, ExpiresAt: -1, OptRole: role}).Post("/api/namespaces/api-key")
		require.NoError(t, err)

		return resp
	}

	t.Run("an owner's key created without a role is an administrator key", func(t *testing.T) {
		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "defaulted", ExpiresAt: -1})
		assert.Equal(t, authorizer.RoleAdministrator, key.Role)
		assert.Equal(t, authorizer.RoleAdministrator, listedAPIKey(t, compose, key.Name).Role)

		resp, err := withAPIKey(t, compose, key.Key).Delete("/api/namespaces/" + ShellHubNamespace)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("no key is created as owner", func(t *testing.T) {
		resp := requestKey(t, compose.R(t.Context()), "owner", authorizer.RoleOwner)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("an operator creates no key", func(t *testing.T) {
		operator := newMember(t, compose, "operator", authorizer.RoleOperator)

		resp := requestKey(t, asBearer(t, compose, operator.Token), "operator", "")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("a key keeps its role after its creator is demoted", func(t *testing.T) {
		creator := newMember(t, compose, "demoted", authorizer.RoleAdministrator)

		key := new(responses.CreateAPIKey)
		resp := requestKey(t, asBearer(t, compose, creator.Token).SetResult(key), "demoted", "")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		require.Equal(t, authorizer.RoleAdministrator, key.Role)

		resp, err := compose.R(t.Context()).
			SetBody(map[string]string{"role": authorizer.RoleOperator.String()}).
			Patch("/api/namespaces/" + ShellHubNamespace + "/members/" + creator.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.Equal(t, authorizer.RoleAdministrator, listedAPIKey(t, compose, key.Name).Role)

		resp, err = withAPIKey(t, compose, key.Key).Get("/api/access-policies")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})
}

func listedAPIKey(t *testing.T, compose *environment.DockerCompose, name string) models.APIKey {
	t.Helper()

	keys := []models.APIKey{}
	resp, err := compose.R(t.Context()).SetResult(&keys).Get("/api/namespaces/api-key")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	index := slices.IndexFunc(keys, func(k models.APIKey) bool { return k.Name == name })
	require.GreaterOrEqual(t, index, 0, "the key %s is not listed", name)

	return keys[index]
}

func withAPIKey(t *testing.T, compose *environment.DockerCompose, plaintext string) *resty.Request {
	t.Helper()

	return compose.Anonymous(t.Context()).SetHeader("X-API-Key", plaintext)
}
