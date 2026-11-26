package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	pkgmiddleware "github.com/shellhub-io/shellhub/pkg/middleware"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/sirupsen/logrus"
)

type mcpCallerKey struct{}

type mcpCaller struct {
	tenantID string
	apiKey   string
}

const mcpSessionIdleTTL = 10 * time.Minute

// SetupMCPRoutes mounts the MCP Streamable HTTP server at /mcp and /mcp/*. It serves only a caller
// that authenticated with a namespace API key, and each tool call it replays carries that key and no
// other credential. A request with no credential, or one the authenticator cannot honour, gets 401
// from the authenticator; one that authenticated any other way gets 403. Both are refused before
// the transport sees the request, so a refused caller never opens a session. A session that receives
// no request for mcpSessionIdleTTL is removed within half that again, by a sweeper goroutine that
// runs for the life of the process. An open GET stream does not count as a request. Each later
// request carrying the removed session's ID still runs, but without the client info and other
// per-session state the old session held. The transport writes only its warnings and errors,
// through the standard logrus logger.
func SetupMCPRoutes(router *echo.Echo) {
	streamable := newMCPHandler(router, mcpSessionIdleTTL, logrus.NewEntry(logrus.StandardLogger()))

	router.Any("/mcp", echo.WrapHandler(streamable), routesmiddleware.RequiresAPIKey)
	router.Any("/mcp/*", echo.WrapHandler(streamable), routesmiddleware.RequiresAPIKey)
}

func newMCPHandler(router http.Handler, idleTTL time.Duration, log *logrus.Entry, opts ...mcpserver.ServerOption) *mcpserver.StreamableHTTPServer {
	s := buildMCPServer(router, opts...)

	httpCtxFn := func(ctx context.Context, r *http.Request) context.Context {
		caller, _ := identity.FromContext(r.Context())
		if caller.TenantID == "" || caller.Role == authorizer.RoleInvalid {
			return ctx
		}

		return context.WithValue(ctx, mcpCallerKey{}, mcpCaller{tenantID: caller.TenantID, apiKey: caller.APIKey})
	}

	return mcpserver.NewStreamableHTTPServer(s,
		mcpserver.WithHTTPContextFunc(httpCtxFn),
		mcpserver.WithSessionIdleTTL(idleTTL),
		mcpserver.WithStreamableHTTPLogger(newMCPLogger(log)),
	)
}

func newMCPLogger(log *logrus.Entry) *slog.Logger {
	return slog.New(warnAndAbove{pkgmiddleware.NewSlogLogger(log).Handler()})
}

type warnAndAbove struct {
	slog.Handler
}

func (h warnAndAbove) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelWarn && h.Handler.Enabled(ctx, level)
}

func (h warnAndAbove) WithAttrs(attrs []slog.Attr) slog.Handler {
	return warnAndAbove{h.Handler.WithAttrs(attrs)}
}

func (h warnAndAbove) WithGroup(name string) slog.Handler {
	return warnAndAbove{h.Handler.WithGroup(name)}
}

func buildMCPServer(router http.Handler, opts ...mcpserver.ServerOption) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("shellhub", "1.0.0",
		append([]mcpserver.ServerOption{mcpserver.WithToolCapabilities(true)}, opts...)...,
	)

	addDeviceTools(s, router)
	addSessionTools(s, router)
	addNamespaceTools(s, router)

	return s
}

func mcpAPICall(ctx context.Context, router http.Handler, method, target string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, target, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if caller, ok := ctx.Value(mcpCallerKey{}).(mcpCaller); ok && caller.apiKey != "" {
		req.Header.Set("X-API-Key", caller.apiKey)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

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
	caller, _ := ctx.Value(mcpCallerKey{}).(mcpCaller)

	return caller.tenantID
}

func intArg(args map[string]any, key string, def int) int {
	v, _ := args[key].(float64)
	if v <= 0 {
		return def
	}

	return int(v)
}

func toJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to encode the result: %v", err)
	}

	return string(b)
}

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
			if name == "" {
				return mcp.NewToolResultError("name must not be empty"), nil
			}

			body, _ := json.Marshal(map[string]string{"name": name})
			rec := mcpAPICall(ctx, router, http.MethodPut,
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

func addNamespaceTools(s *mcpserver.MCPServer, router http.Handler) {
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

			rec := mcpAPICall(ctx, router, http.MethodGet, "/api/namespaces/"+url.PathEscape(tenantID), nil)

			return mcpAPIResult(rec), nil
		},
	)
}
