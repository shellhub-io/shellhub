package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type mcpContextKey string

const (
	mcpKeyTenantID mcpContextKey = "mcp_tenant_id"
	mcpKeyHeaders  mcpContextKey = "mcp_headers"
)

// mcpAuthHeaders are the headers the API gateway injects to identify an API-key
// caller. They are captured from the incoming /mcp request and replayed on the
// in-process API calls the tools make, so the same authentication and
// authorization the gateway performed for the MCP request flows through to the
// REST middleware (BlockAPIKey, RequiresPermission, RequiresTenant). MCP auth
// is API-key-only, so the user-identity headers (X-ID, X-Admin, X-Username)
// never apply and are intentionally omitted.
var mcpAuthHeaders = []string{
	"X-Tenant-ID",
	"X-Role",
	"X-Api-Key",
}

// SetupMCPRoutes mounts the MCP Streamable HTTP server at /mcp.
func SetupMCPRoutes(router *echo.Echo) {
	s := buildMCPServer(router)

	httpCtxFn := func(ctx context.Context, r *http.Request) context.Context {
		tenantID := r.Header.Get("X-Tenant-ID")
		role := authorizer.RoleFromString(r.Header.Get("X-Role"))

		if tenantID == "" || role == authorizer.RoleInvalid {
			return ctx
		}

		ctx = context.WithValue(ctx, mcpKeyTenantID, tenantID)

		// Capture the gateway-injected auth headers so tools can replay them
		// on in-process API calls and inherit the REST middleware.
		headers := http.Header{}
		for _, key := range mcpAuthHeaders {
			if value := r.Header.Get(key); value != "" {
				headers.Set(key, value)
			}
		}
		ctx = context.WithValue(ctx, mcpKeyHeaders, headers)

		return ctx
	}

	streamable := mcpserver.NewStreamableHTTPServer(s,
		mcpserver.WithHTTPContextFunc(httpCtxFn),
	)

	router.Any("/mcp", echo.WrapHandler(streamable))
	router.Any("/mcp/*", echo.WrapHandler(streamable))
}

func buildMCPServer(router http.Handler) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("shellhub", "1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	addDeviceTools(s, router)
	addSessionTools(s, router)
	addNamespaceTools(s, router)

	return s
}

// --- helpers ---

// mcpAPICall replays the caller's request against the API's own Echo router
// in-process, so the full middleware chain (BlockAPIKey, RequiresPermission,
// RequiresTenant) runs exactly as it would for a REST caller. The MCP server
// is an in-binary client of the API, not a shortcut past it -- no network hop,
// but the same authorization path. body may be nil.
func mcpAPICall(ctx context.Context, router http.Handler, method, target string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if headers, ok := ctx.Value(mcpKeyHeaders).(http.Header); ok {
		for key, values := range headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

// mcpAPIResult turns an in-process API response into a tool result: a 2xx
// forwards the JSON body verbatim, anything else maps the status code to a
// fixed message so internal store details never leak to MCP clients.
func mcpAPIResult(rec *httptest.ResponseRecorder) *mcp.CallToolResult {
	if rec.Code >= http.StatusOK && rec.Code < http.StatusMultipleChoices {
		return mcp.NewToolResultText(rec.Body.String())
	}

	switch rec.Code {
	case http.StatusUnauthorized:
		return mcpUnauth()
	case http.StatusForbidden:
		return mcpForbidden()
	case http.StatusNotFound:
		return mcp.NewToolResultError("not found")
	default:
		return mcp.NewToolResultError("internal error")
	}
}

// mcpAPIListResult wraps a paginated list response as {total, <key>: [...]},
// reading the count from the X-Total-Count header the list handlers set and
// forwarding the body array verbatim. Non-2xx responses fall back to the
// shared error mapping.
func mcpAPIListResult(rec *httptest.ResponseRecorder, key string) *mcp.CallToolResult {
	if rec.Code < http.StatusOK || rec.Code >= http.StatusMultipleChoices {
		return mcpAPIResult(rec)
	}

	total, _ := strconv.Atoi(rec.Header().Get("X-Total-Count"))

	return mcp.NewToolResultText(toJSON(map[string]any{
		"total": total,
		key:     json.RawMessage(rec.Body.Bytes()),
	}))
}

// mcpAPIOK returns a fixed success message for write operations whose REST
// response body isn't useful to forward. Non-2xx responses fall back to the
// shared error mapping.
func mcpAPIOK(rec *httptest.ResponseRecorder, msg string) *mcp.CallToolResult {
	if rec.Code >= http.StatusOK && rec.Code < http.StatusMultipleChoices {
		return mcp.NewToolResultText(msg)
	}

	return mcpAPIResult(rec)
}

func mcpForbidden() *mcp.CallToolResult {
	return mcp.NewToolResultError("forbidden: insufficient permissions")
}

func mcpUnauth() *mcp.CallToolResult {
	return mcp.NewToolResultError("unauthorized: missing or invalid token")
}

func tenantFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(mcpKeyTenantID).(string)

	return v
}

func intArg(args map[string]any, key string, def int) int {
	v, _ := args[key].(float64)
	if v <= 0 {
		return def
	}

	return int(v)
}

func toJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")

	return string(b)
}

// --- Device tools ---

func addDeviceTools(s *mcpserver.MCPServer, router http.Handler) {
	s.AddTool(
		mcp.NewTool("shellhub_list_devices",
			mcp.WithDescription("List devices in the ShellHub namespace. Filter by status and paginate results."),
			mcp.WithString("status", mcp.Description("Device status filter: accepted, pending, rejected, removed, unused.")),
			mcp.WithInteger("page", mcp.Description("Page number (1-based). Default: 1.")),
			mcp.WithInteger("per_page", mcp.Description("Results per page (max 100). Default: 20.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			q := url.Values{}
			if status, _ := args["status"].(string); status != "" {
				q.Set("status", status)
			}
			q.Set("page", strconv.Itoa(intArg(args, "page", 1)))
			q.Set("per_page", strconv.Itoa(intArg(args, "per_page", 20)))

			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/devices?"+q.Encode(), nil)

			return mcpAPIListResult(rec, "devices"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_device",
			mcp.WithDescription("Get details of a ShellHub device by its UID."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uid, _ := req.GetArguments()["uid"].(string)

			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/devices/"+url.PathEscape(uid), nil)

			return mcpAPIResult(rec), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_update_device_status",
			mcp.WithDescription("Accept or reject a device. status must be 'accepted' or 'rejected'."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
			mcp.WithString("status", mcp.Required(), mcp.Description("New status: accepted or rejected.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			status, _ := args["status"].(string)

			// The REST route expects the legacy short form in the path param.
			pathStatus := map[string]string{"accepted": "accept", "rejected": "reject"}[status]
			if pathStatus == "" {
				return mcp.NewToolResultError("status must be 'accepted' or 'rejected'"), nil
			}

			rec := mcpAPICall(ctx, router, http.MethodPatch,
				"/api/devices/"+url.PathEscape(uid)+"/"+pathStatus, nil)

			return mcpAPIOK(rec, fmt.Sprintf("device %s status updated to %s", uid, status)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_delete_device",
			mcp.WithDescription("Permanently delete a ShellHub device. This action is irreversible."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uid, _ := req.GetArguments()["uid"].(string)

			rec := mcpAPICall(ctx, router, http.MethodDelete, "/api/devices/"+url.PathEscape(uid), nil)

			return mcpAPIOK(rec, fmt.Sprintf("device %s deleted", uid)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_rename_device",
			mcp.WithDescription("Rename a ShellHub device."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
			mcp.WithString("name", mcp.Required(), mcp.Description("New device name (hostname format).")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			name, _ := args["name"].(string)

			body, _ := json.Marshal(map[string]string{"name": name})
			rec := mcpAPICall(ctx, router, http.MethodPatch,
				"/api/devices/"+url.PathEscape(uid), bytes.NewReader(body))

			return mcpAPIOK(rec, fmt.Sprintf("device %s renamed to %s", uid, name)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_stats",
			mcp.WithDescription("Get ShellHub instance statistics: registered/online/pending/rejected devices and active sessions."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/stats", nil)

			return mcpAPIResult(rec), nil
		},
	)
}

// --- Session tools ---

func addSessionTools(s *mcpserver.MCPServer, router http.Handler) {
	s.AddTool(
		mcp.NewTool("shellhub_list_sessions",
			mcp.WithDescription("List SSH sessions in the namespace. Returns active and historical sessions."),
			mcp.WithInteger("page", mcp.Description("Page number (1-based). Default: 1.")),
			mcp.WithInteger("per_page", mcp.Description("Results per page (max 100). Default: 20.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			q := url.Values{}
			q.Set("page", strconv.Itoa(intArg(args, "page", 1)))
			q.Set("per_page", strconv.Itoa(intArg(args, "per_page", 20)))

			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/sessions?"+q.Encode(), nil)

			return mcpAPIListResult(rec, "sessions"), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_session",
			mcp.WithDescription("Get details of a specific SSH session by its UID."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Session UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			uid, _ := req.GetArguments()["uid"].(string)

			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/sessions/"+url.PathEscape(uid), nil)

			return mcpAPIResult(rec), nil
		},
	)
}

// --- Namespace tools ---

func addNamespaceTools(s *mcpserver.MCPServer, router http.Handler) {
	// No "list namespaces" tool: an API key is scoped to one namespace, and
	// listing namespaces is an account-level action the REST API blocks for
	// API keys (BlockAPIKey on GetNamespaceList). get_namespace reads the
	// caller's own namespace.
	s.AddTool(
		mcp.NewTool("shellhub_get_namespace",
			mcp.WithDescription("Get details and settings of the caller's ShellHub namespace."),
			mcp.WithString("tenant_id", mcp.Description("Namespace tenant ID (UUID). Defaults to the caller's own namespace; any other value is rejected.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			tenantID, _ := args["tenant_id"].(string)
			if tenantID == "" {
				tenantID = tenantFromCtx(ctx)
			}

			// Cross-tenant access is rejected by RequiresTenant on the route;
			// no manual guard needed here.
			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/namespaces/"+url.PathEscape(tenantID), nil)

			return mcpAPIResult(rec), nil
		},
	)
}
