package main

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secondUsername  = "other"
	secondEmail     = "other@shellhub.io"
	secondPassword  = "other-password"
	secondNamespace = "otherspace"
	secondTenant    = "11111111-1111-4111-8111-111111111111"
)

func seedTwoNamespaces(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace)
	compose.NewUser(t, secondUsername, secondEmail, secondPassword)
	compose.NewNamespace(t, secondUsername, secondNamespace, secondTenant)
}

func TestRequestFieldsKeepTheirDeclaredSource(t *testing.T) {
	t.Parallel()

	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	seedTwoNamespaces(t, compose)

	caller := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)
	other := compose.AuthUser(t, secondUsername, secondPassword)

	compose.JWT(caller.Token)
	key := compose.CreateAPIKey(t, &requests.CreateAPIKey{
		Name:      "automation",
		ExpiresAt: -1,
		OptRole:   authorizer.RoleAdministrator,
	})

	compose.JWT(other.Token)
	before := compose.GetNamespace(t, secondTenant)

	client := resty.New().SetBaseURL(compose.BaseURL()).SetAllowGetMethodPayload(true)
	asOwner := func(t *testing.T) *resty.Request {
		t.Helper()

		return client.R().SetContext(t.Context()).SetAuthToken(caller.Token)
	}
	asAPIKey := func(t *testing.T) *resty.Request {
		t.Helper()

		return client.R().SetContext(t.Context()).SetHeader("X-API-Key", key.Key)
	}

	cases := []struct {
		description string
		request     func(*testing.T) *resty.Request
		method      string
		path        string
		body        map[string]any
		status      int
	}{
		{
			description: "PUT /api/namespaces/:tenant reads the tenant from the path",
			request:     asOwner,
			method:      http.MethodPut,
			path:        "/api/namespaces/" + ShellHubNamespace,
			body:        map[string]any{"Tenant": secondTenant, "name": "renamed", "settings": map[string]any{"session_record": false}},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "PUT /api/users/security/:tenant reads the tenant from the path",
			request:     asOwner,
			method:      http.MethodPut,
			path:        "/api/users/security/" + ShellHubNamespace,
			body:        map[string]any{"Tenant": secondTenant, "session_record": false},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "GET /api/namespaces/:tenant/members reads the tenant from the path",
			request:     asOwner,
			method:      http.MethodGet,
			path:        "/api/namespaces/" + ShellHubNamespace + "/members",
			body:        map[string]any{"TenantID": secondTenant},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "GET /api/namespaces/:tenant reads the tenant from the path for an API key",
			request:     asAPIKey,
			method:      http.MethodGet,
			path:        "/api/namespaces/" + ShellHubNamespace,
			body:        map[string]any{"Tenant": secondTenant},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "POST /api/namespaces/:tenant/members reads the user from the identity",
			request:     asOwner,
			method:      http.MethodPost,
			path:        "/api/namespaces/" + ShellHubNamespace + "/members",
			body:        map[string]any{"UserID": other.ID, "email": "member@shellhub.io", "role": "administrator"},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "GET /api/namespaces/:tenant/invitations reads the user from the identity",
			request:     asAPIKey,
			method:      http.MethodGet,
			path:        "/api/namespaces/" + ShellHubNamespace + "/invitations",
			body:        map[string]any{"UserID": other.ID},
			status:      http.StatusUnprocessableEntity,
		},
		{
			description: "POST /api/namespaces/:tenant/members refuses an API key before binding",
			request:     asAPIKey,
			method:      http.MethodPost,
			path:        "/api/namespaces/" + ShellHubNamespace + "/members",
			body:        map[string]any{"email": "member@shellhub.io", "role": "administrator"},
			status:      http.StatusForbidden,
		},
		{
			description: "DELETE /api/namespaces/:tenant reads the tenant from the path",
			request:     asOwner,
			method:      http.MethodDelete,
			path:        "/api/namespaces/" + ShellHubNamespace,
			body:        map[string]any{"Tenant": secondTenant},
			status:      http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			resp, err := tc.request(t).SetBody(tc.body).Execute(tc.method, tc.path)
			require.NoError(t, err)

			assert.Equal(t, tc.status, resp.StatusCode(), resp.String())
			assert.NotContains(t, resp.String(), secondNamespace)
		})
	}

	compose.JWT(other.Token)
	after := compose.GetNamespace(t, secondTenant)
	assert.Equal(t, before.Name, after.Name)
	assert.Equal(t, before.Settings, after.Settings)

	compose.JWT(caller.Token)
	assert.Equal(t, ShellHubNamespaceName, compose.GetNamespace(t, ShellHubNamespace).Name)
}

func TestMemberRoutesActOnTheTokenNamespace(t *testing.T) {
	t.Parallel()

	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	seedTwoNamespaces(t, compose)
	compose.NewMember(t, ShellHubUsername, secondNamespace, string(authorizer.RoleAdministrator))

	login := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)

	scoped := new(models.UserAuthResponse)
	resp, err := compose.Anonymous(t.Context()).SetAuthToken(login.Token).SetResult(scoped).Get("/api/auth/token/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotNil(t, scoped.Tenant)
	require.Equal(t, ShellHubNamespace, *scoped.Tenant)

	invite := map[string]any{"email": "member@shellhub.io", "role": "observer"}

	for _, path := range []string{
		"/api/namespaces/" + secondTenant + "/members",
		"/api/namespaces/" + secondTenant + "/invitations/links",
	} {
		t.Run("POST "+path, func(t *testing.T) {
			resp, err := compose.Anonymous(t.Context()).SetAuthToken(scoped.Token).SetBody(invite).Post(path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		})
	}

	compose.JWT(compose.AuthUser(t, secondUsername, secondPassword).Token)
	resp, err = compose.R(t.Context()).Get("/api/namespaces/" + secondTenant + "/invitations")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	assert.NotContains(t, resp.String(), "member@shellhub.io")
}
