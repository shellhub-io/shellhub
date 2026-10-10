package main

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnterpriseAdmin drives the enterprise admin API as an instance administrator and as the
// callers it must turn away: the instance API keys an administrator mints and when they stop
// authenticating, the guard that keeps every other user off /admin/api, the user and namespace
// exports, the instance-wide statistics, and the token that signs an administrator in as another
// user. The cases share one stack; each creates the users, namespaces and devices it reads.
func TestEnterpriseAdmin(t *testing.T) {
	compose := newEnterpriseEnvironment(t, context.Background(), environment.New(t, run))

	t.Run("instance API keys", func(t *testing.T) { testInstanceAPIKeys(t, compose) })
	t.Run("access guard", func(t *testing.T) { testAdminAccessGuard(t, compose) })
	t.Run("exports", func(t *testing.T) { testAdminExports(t, compose) })
	t.Run("stats", func(t *testing.T) { testAdminStats(t, compose) })
	t.Run("auth token", func(t *testing.T) { testAdminAuthToken(t, compose) })
}

func testInstanceAPIKeys(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("an instance key authenticates on the admin routes", func(t *testing.T) {
		key := createInstanceAPIKey(t, compose.R(t.Context()), "automation")

		for _, path := range []string{"/admin/api/users", "/admin/api/namespaces", "/admin/api/devices"} {
			requireAnswer(t, withAPIKey(t, compose, key.ID), http.MethodGet, path, http.StatusOK)
		}

		requireAnswer(t, withAPIKey(t, compose, models.InstanceAPIKeyPrefix+uuid.Generate()), http.MethodGet, "/admin/api/users", http.StatusUnauthorized)
	})

	t.Run("a user who is not an instance administrator cannot create an instance key", func(t *testing.T) {
		compose.NewUser(t, "plain", "plain@shellhub.io", ShellHubPassword)
		user := compose.AuthUser(t, "plain", ShellHubPassword)

		resp, err := asBearer(t, compose, user.Token).
			SetBody(map[string]any{"name": "unprivileged", "expires_at": 30}).
			Post("/admin/api/instance-api-keys")
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		assert.NotContains(t, instanceAPIKeyNames(t, compose), "unprivileged")
	})

	t.Run("an instance key stops authenticating once its creator is no longer an administrator", func(t *testing.T) {
		compose.NewAdmin(t, "keeper", "keeper@shellhub.io", ShellHubPassword)
		keeper := compose.AuthUser(t, "keeper", ShellHubPassword)

		key := createInstanceAPIKey(t, asBearer(t, compose, keeper.Token), "keeper-automation")
		require.Equal(t, keeper.ID, key.CreatedBy)
		requireAnswer(t, withAPIKey(t, compose, key.ID), http.MethodGet, "/admin/api/users", http.StatusOK)

		resp, err := compose.R(t.Context()).
			SetBody(map[string]any{"admin": false}).
			Put("/admin/api/users/" + keeper.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		requireAnswer(t, withAPIKey(t, compose, key.ID), http.MethodGet, "/admin/api/users", http.StatusUnauthorized)
	})

	t.Run("an expired instance key is refused", func(t *testing.T) {
		key := createInstanceAPIKey(t, compose.R(t.Context()), "expiring")
		requireAnswer(t, withAPIKey(t, compose, key.ID), http.MethodGet, "/admin/api/users", http.StatusOK)

		compose.ExpireInstanceAPIKey(t, key.Name)

		requireAnswer(t, withAPIKey(t, compose, key.ID), http.MethodGet, "/admin/api/users", http.StatusUnauthorized)
	})
}

func testAdminAccessGuard(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	compose.NewUser(t, "outsider", "outsider@shellhub.io", ShellHubPassword)
	outsider := compose.AuthUser(t, "outsider", ShellHubPassword)

	t.Run("an administrator's token reaches the admin routes", func(t *testing.T) {
		requireAnswer(t, compose.R(t.Context()), http.MethodGet, "/admin/api/users", http.StatusOK)
	})

	t.Run("a token without the administrator flag is refused on the admin routes", func(t *testing.T) {
		for _, path := range []string{"/admin/api/users", "/admin/api/stats", "/admin/api/export/users"} {
			requireAnswer(t, asBearer(t, compose, outsider.Token), http.MethodGet, path, http.StatusForbidden)
		}
	})

	t.Run("a forged administrator header does not let a token without the flag through", func(t *testing.T) {
		requireAnswer(t, asBearer(t, compose, outsider.Token).SetHeader("X-Admin", "true"), http.MethodGet, "/admin/api/users", http.StatusForbidden)
	})
}

func testAdminExports(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	compose.NewUser(t, "export-alpha", "alpha@export.shellhub.io", ShellHubPassword)
	compose.NewUser(t, "export-beta", "beta@export.shellhub.io", ShellHubPassword)

	tenant := uuid.Generate()
	compose.NewNamespace(t, "export-alpha", "export-space", tenant, "")

	alpha := adminUser(t, compose, "export-alpha")
	beta := adminUser(t, compose, "export-beta")

	t.Run("the user export lists every user with the namespaces each owns", func(t *testing.T) {
		header, rows := exportCSV(t, compose, "/admin/api/export/users", "users.csv", "")

		assert.Equal(t, []string{"Name", "email", "username", "ID", "namespaces"}, header)
		assert.ElementsMatch(t, usernames(adminUsers(t, compose)), column(rows, 2))
		assert.Contains(t, rows, []string{alpha.Name, alpha.Email, alpha.Username, alpha.ID, "1"})
		assert.Contains(t, rows, []string{beta.Name, beta.Email, beta.Username, beta.ID, "0"})
	})

	t.Run("a filter narrows the user export to the users it matches", func(t *testing.T) {
		header, rows := exportCSV(t, compose, "/admin/api/export/users", "users.csv", propertyFilter(t, "username", "contains", "export-"))

		assert.Equal(t, []string{"Name", "email", "username", "ID", "namespaces"}, header)
		assert.ElementsMatch(t, [][]string{
			{alpha.Name, alpha.Email, alpha.Username, alpha.ID, "1"},
			{beta.Name, beta.Email, beta.Username, beta.ID, "0"},
		}, rows)
	})

	t.Run("a user export no user matches answers 204 without a file", func(t *testing.T) {
		requireEmptyExport(t, compose, "/admin/api/export/users", propertyFilter(t, "username", "eq", "nobody"))
	})

	t.Run("the namespace export lists every namespace with its owner", func(t *testing.T) {
		namespace := adminNamespace(t, compose, tenant)

		header, rows := exportCSV(t, compose, "/admin/api/export/namespaces", "namespaces.csv", "")

		assert.Equal(t, []string{"Namespace", "Owner Name", "Owner Username", "Email", "tenantID", "Type"}, header)
		assert.ElementsMatch(t, tenants(adminNamespaces(t, compose)), column(rows, 4))
		assert.Contains(t, rows, []string{"export-space", alpha.Name, alpha.Username, alpha.Email, tenant, string(namespace.Type)})
	})

	t.Run("a namespace export no namespace matches answers 204 without a file", func(t *testing.T) {
		requireEmptyExport(t, compose, "/admin/api/export/namespaces", propertyFilter(t, "name", "eq", "nowhere"))
	})
}

type adminStats struct {
	models.Stats
	RegisteredUsers int64 `json:"registered_users"`
}

func testAdminStats(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("the stats count users, devices and sessions across every namespace", func(t *testing.T) {
		ctx := t.Context()
		before := readAdminStats(t, compose)

		_, connected := startAcceptedAgent(t, ctx, compose)
		conn, session := openSession(t, ctx, compose, connected, registerDeviceKey(t, ctx, compose), "")
		requireSessionActive(t, ctx, compose, session.UID, true)

		compose.NewUser(t, "stats-owner", "stats-owner@shellhub.io", ShellHubPassword)
		tenant := uuid.Generate()
		compose.NewNamespace(t, "stats-owner", "stats-space", tenant, "")
		owner := compose.AuthUser(t, "stats-owner", ShellHubPassword)

		enrollPendingDevice(t, compose, "stats-pending", "02:00:00:00:24:01")
		enrollPendingDeviceAs(t, compose, owner, "stats-other-pending", "02:00:00:00:24:02")

		accepted := enrollPendingDeviceAs(t, compose, owner, "stats-accepted", "02:00:00:00:24:03")
		requireAnswer(t, asBearer(t, compose, owner.Token), http.MethodPatch, "/api/devices/"+accepted.UID+"/"+string(environment.DeviceActionAccept), http.StatusOK)

		rejected := enrollPendingDeviceAs(t, compose, owner, "stats-rejected", "02:00:00:00:24:04")
		requireAnswer(t, asBearer(t, compose, owner.Token), http.MethodPatch, "/api/devices/"+rejected.UID+"/"+string(environment.DeviceActionReject), http.StatusOK)

		after := readAdminStats(t, compose)

		assert.Equal(t, before.RegisteredUsers+1, after.RegisteredUsers, "registered users")
		assert.Equal(t, before.PendingDevices+2, after.PendingDevices, "pending devices")
		assert.Equal(t, before.RegisteredDevices+2, after.RegisteredDevices, "registered devices")
		assert.Equal(t, before.RejectedDevices+1, after.RejectedDevices, "rejected devices")
		assert.Equal(t, before.OnlineDevices+2, after.OnlineDevices, "online devices")
		assert.Equal(t, before.ActiveSessions+1, after.ActiveSessions, "active sessions")

		require.NoError(t, conn.Close())
		requireSessionActive(t, ctx, compose, session.UID, false)

		assert.Equal(t, before.ActiveSessions, readAdminStats(t, compose).ActiveSessions, "active sessions once the session ended")
	})
}

func testAdminAuthToken(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("an administrator gets a token that signs in as the user", func(t *testing.T) {
		compose.NewUser(t, "impersonated", "impersonated@shellhub.io", ShellHubPassword)
		tenant := uuid.Generate()
		compose.NewNamespace(t, "impersonated", "impersonated-space", tenant, "")
		target := adminUser(t, compose, "impersonated")

		token := struct {
			Token string `json:"token"`
		}{}

		resp, err := compose.R(t.Context()).SetResult(&token).Get("/admin/api/auth/token/" + target.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		claims, err := jwttoken.ClaimsFromBearerToken(compose.APIPublicKey(t), token.Token)
		require.NoError(t, err)
		userClaims, ok := claims.(*authorizer.UserClaims)
		require.True(t, ok, "the token carries %T, not user claims", claims)
		assert.Equal(t, target.ID, userClaims.ID)
		assert.Equal(t, tenant, userClaims.TenantID)

		signedIn := new(models.UserAuthResponse)

		resp, err = asBearer(t, compose, token.Token).SetResult(signedIn).Get("/api/auth/user")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, target.ID, signedIn.ID)
		assert.Equal(t, "impersonated", signedIn.User)
		require.NotNil(t, signedIn.Tenant)
		assert.Equal(t, tenant, *signedIn.Tenant)
	})

	t.Run("a token for a user that does not exist is refused as not found", func(t *testing.T) {
		requireAnswer(t, compose.R(t.Context()), http.MethodGet, "/admin/api/auth/token/"+uuid.Generate(), http.StatusNotFound)
	})
}

func createInstanceAPIKey(t *testing.T, req *resty.Request, name string) *responses.CreateInstanceAPIKey {
	t.Helper()

	key := new(responses.CreateInstanceAPIKey)

	resp, err := req.
		SetBody(map[string]any{"name": name, "expires_at": 30}).
		SetResult(key).
		Post("/admin/api/instance-api-keys")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.True(t, strings.HasPrefix(key.ID, models.InstanceAPIKeyPrefix), "the plaintext key carries the instance key prefix")

	return key
}

func instanceAPIKeyNames(t *testing.T, compose *environment.DockerCompose) []string {
	t.Helper()

	keys := []models.InstanceAPIKey{}

	resp, err := compose.R(t.Context()).SetQueryParam("per_page", "100").SetResult(&keys).Get("/admin/api/instance-api-keys")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	total, err := strconv.Atoi(resp.Header().Get("X-Total-Count"))
	require.NoError(t, err)
	require.Len(t, keys, total, "the instance holds more instance API keys than one page lists")

	names := make([]string, 0, len(keys))
	for _, key := range keys {
		names = append(names, key.Name)
	}

	return names
}

func requireAnswer(t *testing.T, req *resty.Request, method, path string, status int) {
	t.Helper()

	resp, err := req.Execute(method, path)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), "%s %s: %s", method, path, resp.String())
}

func adminUsers(t *testing.T, compose *environment.DockerCompose) []models.User {
	t.Helper()

	users := []models.User{}

	resp, err := compose.R(t.Context()).SetQueryParam("per_page", "100").SetResult(&users).Get("/admin/api/users")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	total, err := strconv.Atoi(resp.Header().Get("X-Total-Count"))
	require.NoError(t, err)
	require.Len(t, users, total, "the instance holds more users than one page lists")

	return users
}

func adminUser(t *testing.T, compose *environment.DockerCompose, username string) models.User {
	t.Helper()

	for _, user := range adminUsers(t, compose) {
		if user.Username == username {
			return user
		}
	}

	require.FailNow(t, "the admin API lists no user "+username)

	return models.User{}
}

func adminNamespace(t *testing.T, compose *environment.DockerCompose, tenant string) models.Namespace {
	t.Helper()

	namespace := models.Namespace{}

	resp, err := compose.R(t.Context()).SetResult(&namespace).Get("/admin/api/namespaces/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return namespace
}

func adminNamespaces(t *testing.T, compose *environment.DockerCompose) []models.Namespace {
	t.Helper()

	namespaces := []models.Namespace{}

	resp, err := compose.R(t.Context()).SetQueryParam("per_page", "100").SetResult(&namespaces).Get("/admin/api/namespaces")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	total, err := strconv.Atoi(resp.Header().Get("X-Total-Count"))
	require.NoError(t, err)
	require.Len(t, namespaces, total, "the instance holds more namespaces than one page lists")

	return namespaces
}

func tenants(namespaces []models.Namespace) []string {
	ids := make([]string, 0, len(namespaces))
	for _, namespace := range namespaces {
		ids = append(ids, namespace.TenantID)
	}

	return ids
}

func usernames(users []models.User) []string {
	names := make([]string, 0, len(users))
	for _, user := range users {
		names = append(names, user.Username)
	}

	return names
}

func column(rows [][]string, index int) []string {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		values = append(values, row[index])
	}

	return values
}

func exportCSV(t *testing.T, compose *environment.DockerCompose, path, filename, filter string) ([]string, [][]string) {
	t.Helper()

	req := compose.R(t.Context())
	if filter != "" {
		req.SetQueryParam("filter", filter)
	}

	resp, err := req.Get(path)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	assert.Equal(t, "attachment; filename="+filename, resp.Header().Get("Content-Disposition"))

	records, err := csv.NewReader(strings.NewReader(resp.String())).ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, records, "the export has no header row")

	return records[0], records[1:]
}

func requireEmptyExport(t *testing.T, compose *environment.DockerCompose, path, filter string) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetQueryParam("filter", filter).Get(path)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
	assert.Empty(t, resp.Header().Get("Content-Disposition"))
	assert.Empty(t, resp.String())
}

func readAdminStats(t *testing.T, compose *environment.DockerCompose) adminStats {
	t.Helper()

	stats := adminStats{}

	resp, err := compose.R(t.Context()).SetResult(&stats).Get("/admin/api/stats")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return stats
}

func enrollPendingDeviceAs(t *testing.T, compose *environment.DockerCompose, owner *models.UserAuthResponse, hostname, mac string) models.Device {
	t.Helper()

	require.NotNil(t, owner.Tenant, "the owner's token carries no namespace")

	req := newDeviceAuthRequest(t, hostname, mac)
	req.TenantID = *owner.Tenant

	device := models.Device{}

	resp, err := asBearer(t, compose, owner.Token).SetResult(&device).Get("/api/devices/" + authDevice(t, compose, req).UID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Equal(t, models.DeviceStatusPending, device.Status)

	t.Cleanup(func() {
		resp, err := compose.Anonymous(context.WithoutCancel(t.Context())).SetAuthToken(owner.Token).Delete("/api/devices/" + device.UID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	return device
}
