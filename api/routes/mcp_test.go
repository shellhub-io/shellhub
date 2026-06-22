package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mcpToolCall builds a tools/call JSON-RPC body for the given tool and args.
func mcpToolCall(name, args string) string {
	return `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"` + name + `","arguments":` + args + `}}`
}

const mcpCallerTenant = "00000000-0000-4000-0000-000000000000"

// mcpCall posts a JSON-RPC message to the MCP endpoint, simulating the headers
// nginx injects for an API-key caller (X-Tenant-ID and X-Role, no X-ID).
func mcpCall(t *testing.T, router http.Handler, tenant, role, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// The stateless session manager only validates the format of the session
	// ID, not its existence -- a well-formed value stands in for a client that
	// has already completed the initialize handshake.
	req.Header.Set("Mcp-Session-Id", "mcp-session-00000000-0000-4000-8000-000000000000")
	if tenant != "" {
		req.Header.Set("X-Tenant-ID", tenant)
	}
	if role != "" {
		req.Header.Set("X-Role", role)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

// mcpToolResult decodes a tools/call response into the tool's first text
// content and its isError flag.
func mcpToolResult(t *testing.T, rec *httptest.ResponseRecorder) (string, bool) {
	t.Helper()

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var env struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env), rec.Body.String())
	require.NotEmpty(t, env.Result.Content)

	return env.Result.Content[0].Text, env.Result.IsError
}

// TestMCPGetNamespaceDefaultsToCaller ensures shellhub_get_namespace reads the
// caller's own namespace when no tenant_id argument is supplied -- the MCP
// credential is scoped to a single namespace.
func TestMCPGetNamespaceDefaultsToCaller(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("GetNamespace", gomock.Anything, mcpCallerTenant).
		Return(&models.Namespace{TenantID: mcpCallerTenant, Name: "dev"}, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"shellhub_get_namespace","arguments":{}}}`)

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, mcpCallerTenant)
	mock.AssertExpectations(t)
}

// TestMCPGetNamespaceRejectsOtherTenant ensures the caller cannot read another
// tenant's namespace by passing an explicit tenant_id. The service must never
// be reached.
func TestMCPGetNamespaceRejectsOtherTenant(t *testing.T) {
	mock := mocks.NewMockService(t)

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"shellhub_get_namespace","arguments":{"tenant_id":"11111111-1111-4000-0000-000000000000"}}}`)

	text, isErr := mcpToolResult(t, rec)
	assert.True(t, isErr)
	assert.Contains(t, text, "forbidden")
	mock.AssertNotCalled(t, "GetNamespace", gomock.Anything, gomock.Anything)
}

// TestMCPDoesNotExposeListNamespaces locks in parity with the REST API: listing
// namespaces is an account-level action blocked for API keys (BlockAPIKey on
// GetNamespaceList), so the MCP server must not expose it as a tool.
func TestMCPDoesNotExposeListNamespaces(t *testing.T) {
	mock := mocks.NewMockService(t)

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var env struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env), rec.Body.String())
	require.NotEmpty(t, env.Result.Tools)

	for _, tool := range env.Result.Tools {
		assert.NotEqual(t, "shellhub_list_namespaces", tool.Name)
	}
}

// --- device tools ---

// TestMCPListDevices ensures status and pagination args reach the REST list
// handler and the response is wrapped as {total, devices}.
func TestMCPListDevices(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("ListDevices", gomock.Anything, gomock.MatchedBy(func(r *requests.DeviceList) bool {
			return r.TenantID == mcpCallerTenant &&
				r.DeviceStatus == models.DeviceStatus("accepted") &&
				r.Paginator.Page == 2 && r.Paginator.PerPage == 50
		})).
		Return([]models.Device{{UID: "uid1"}}, 1, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_list_devices", `{"status":"accepted","page":2,"per_page":50}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, `"total": 1`)
	assert.Contains(t, text, "uid1")
	mock.AssertExpectations(t)
}

// TestMCPGetDevice ensures the uid arg becomes the path parameter.
func TestMCPGetDevice(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("GetDevice", gomock.Anything, models.UID("uid1")).
		Return(&models.Device{UID: "uid1", Name: "dev1"}, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_get_device", `{"uid":"uid1"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "dev1")
	mock.AssertExpectations(t)
}

// TestMCPUpdateDeviceStatus ensures uid and status become path parameters and
// an owner (who has DeviceAccept) succeeds.
func TestMCPUpdateDeviceStatus(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("UpdateDeviceStatus", gomock.Anything, gomock.MatchedBy(func(r *requests.DeviceUpdateStatus) bool {
			return r.UID == "uid1" && r.Status == "accepted"
		})).
		Return(nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_update_device_status", `{"uid":"uid1","status":"accepted"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "status updated")
	mock.AssertExpectations(t)
}

// TestMCPUpdateDeviceStatusForbidden is the crux of the parity work: an
// observer lacks DeviceAccept, so RequiresPermission on the route returns 403
// before the handler runs. The tool no longer checks permissions itself -- the
// middleware does. The service must never be reached.
func TestMCPUpdateDeviceStatusForbidden(t *testing.T) {
	mock := mocks.NewMockService(t)

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleObserver.String(),
		mcpToolCall("shellhub_update_device_status", `{"uid":"uid1","status":"accepted"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.True(t, isErr)
	assert.Contains(t, text, "forbidden")
	mock.AssertNotCalled(t, "UpdateDeviceStatus", gomock.Anything, gomock.Anything)
}

// TestMCPDeleteDevice ensures the uid arg becomes the path parameter.
func TestMCPDeleteDevice(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("DeleteDevice", gomock.Anything, models.UID("uid1"), mcpCallerTenant).
		Return(nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_delete_device", `{"uid":"uid1"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "deleted")
	mock.AssertExpectations(t)
}

// TestMCPRenameDevice ensures the name arg is sent as the JSON body.
func TestMCPRenameDevice(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("RenameDevice", gomock.Anything, models.UID("uid1"), "newname", mcpCallerTenant).
		Return(nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_rename_device", `{"uid":"uid1","name":"newname"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "renamed")
	mock.AssertExpectations(t)
}

// TestMCPGetStats ensures the tool reaches the stats handler scoped to the
// caller's tenant.
func TestMCPGetStats(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("GetStats", gomock.Anything, gomock.MatchedBy(func(r *requests.GetStats) bool {
			return r.TenantID == mcpCallerTenant
		})).
		Return(&models.Stats{RegisteredDevices: 7}, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_get_stats", `{}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "registered_devices")
	mock.AssertExpectations(t)
}

// --- session tools ---

// TestMCPListSessions ensures pagination reaches the list handler and the
// response is wrapped as {total, sessions}.
func TestMCPListSessions(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("ListSessions", gomock.Anything, gomock.MatchedBy(func(r *requests.ListSessions) bool {
			return r.TenantID == mcpCallerTenant && r.Paginator.Page == 1 && r.Paginator.PerPage == 20
		})).
		Return([]models.Session{{UID: "sess1"}}, 1, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_list_sessions", `{}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, `"total": 1`)
	assert.Contains(t, text, "sess1")
	mock.AssertExpectations(t)
}

// TestMCPGetSession ensures the uid arg becomes the path parameter.
func TestMCPGetSession(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("GetSession", gomock.Anything, models.UID("sess1")).
		Return(&models.Session{UID: "sess1"}, nil).
		Once()

	rec := mcpCall(t, NewRouter(mock), mcpCallerTenant, authorizer.RoleOwner.String(),
		mcpToolCall("shellhub_get_session", `{"uid":"sess1"}`))

	text, isErr := mcpToolResult(t, rec)
	assert.False(t, isErr)
	assert.Contains(t, text, "sess1")
	mock.AssertExpectations(t)
}
