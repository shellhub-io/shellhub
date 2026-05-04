package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type mcpContextKey string

const (
	mcpKeyUserID   mcpContextKey = "mcp_user_id"
	mcpKeyTenantID mcpContextKey = "mcp_tenant_id"
	mcpKeyRole     mcpContextKey = "mcp_role"
)

// SetupMCPRoutes mounts the MCP Streamable HTTP server at /mcp.
func SetupMCPRoutes(router *echo.Echo, service services.Service) {
	s := buildMCPServer(service)

	httpCtxFn := func(ctx context.Context, r *http.Request) context.Context {
		bearerToken := r.Header.Get("Authorization")
		claims, err := jwttoken.ClaimsFromBearerToken(service.PublicKey(), bearerToken)
		if err != nil {
			return ctx
		}

		userClaims, ok := claims.(*authorizer.UserClaims)
		if !ok {
			return ctx
		}

		role := authorizer.RoleFromString("")
		if userClaims.TenantID != "" {
			if roleStr, err := service.GetUserRole(ctx, userClaims.TenantID, userClaims.ID); err == nil {
				role = authorizer.RoleFromString(roleStr)
			}
		}

		ctx = context.WithValue(ctx, mcpKeyUserID, userClaims.ID)
		ctx = context.WithValue(ctx, mcpKeyTenantID, userClaims.TenantID)
		ctx = context.WithValue(ctx, mcpKeyRole, role)

		// Mirror identity into the keys gateway.{Tenant,Username,ID}FromContext
		// looks up so service methods that derive scope from the context (e.g.
		// GetDevice, GetSession) automatically scope queries to this tenant
		// instead of running unscoped — which would expose other namespaces.
		ctx = context.WithValue(ctx, "tenant", userClaims.TenantID)   //nolint:staticcheck // SA1029: matches gateway.TenantFromContext key
		ctx = context.WithValue(ctx, "ID", userClaims.ID)             //nolint:staticcheck // SA1029: matches gateway.IDFromContext key
		ctx = context.WithValue(ctx, "username", userClaims.Username) //nolint:staticcheck // SA1029: matches gateway.UsernameFromContext key

		return ctx
	}

	streamable := mcpserver.NewStreamableHTTPServer(s,
		mcpserver.WithHTTPContextFunc(httpCtxFn),
	)

	router.Any("/mcp", echo.WrapHandler(streamable))
	router.Any("/mcp/*", echo.WrapHandler(streamable))
}

func buildMCPServer(svc services.Service) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("shellhub", "1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	addDeviceTools(s, svc)
	addSessionTools(s, svc)
	addNamespaceTools(s, svc)

	return s
}

// --- helpers ---

func mcpForbidden() *mcp.CallToolResult {
	return mcp.NewToolResultError("forbidden: insufficient permissions")
}

func mcpUnauth() *mcp.CallToolResult {
	return mcp.NewToolResultError("unauthorized: missing or invalid token")
}

func roleFromCtx(ctx context.Context) (authorizer.Role, bool) {
	r, ok := ctx.Value(mcpKeyRole).(authorizer.Role)

	return r, ok
}

func tenantFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(mcpKeyTenantID).(string)

	return v
}

func userIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(mcpKeyUserID).(string)

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

func addDeviceTools(s *mcpserver.MCPServer, svc services.Service) {
	s.AddTool(
		mcp.NewTool("shellhub_list_devices",
			mcp.WithDescription("List devices in the ShellHub namespace. Filter by status and paginate results."),
			mcp.WithString("status", mcp.Description("Device status filter: accepted, pending, rejected, removed, unused.")),
			mcp.WithInteger("page", mcp.Description("Page number (1-based). Default: 1.")),
			mcp.WithInteger("per_page", mcp.Description("Results per page (max 100). Default: 20.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.DeviceDetails) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			status, _ := args["status"].(string)
			page := intArg(args, "page", 1)
			perPage := intArg(args, "per_page", 20)

			r := &requests.DeviceList{
				TenantID:     tenantFromCtx(ctx),
				DeviceStatus: models.DeviceStatus(status),
				Paginator:    query.Paginator{Page: page, PerPage: perPage},
			}
			r.Paginator.Normalize()

			devices, count, err := svc.ListDevices(ctx, r)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(map[string]any{
				"total":   count,
				"devices": devices,
			})), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_device",
			mcp.WithDescription("Get details of a ShellHub device by its UID."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.DeviceDetails) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			device, err := svc.GetDevice(ctx, models.UID(uid))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(device)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_update_device_status",
			mcp.WithDescription("Accept or reject a device. status must be 'accepted' or 'rejected'."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
			mcp.WithString("status", mcp.Required(), mcp.Description("New status: accepted or rejected.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}

			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			status, _ := args["status"].(string)

			switch status {
			case "accepted":
				if !role.HasPermission(authorizer.DeviceAccept) {
					return mcpForbidden(), nil
				}
			case "rejected":
				if !role.HasPermission(authorizer.DeviceReject) {
					return mcpForbidden(), nil
				}
			default:
				return mcp.NewToolResultError("status must be 'accepted' or 'rejected'"), nil
			}

			r := &requests.DeviceUpdateStatus{
				TenantID: tenantFromCtx(ctx),
				UID:      uid,
				Status:   status,
			}

			if err := svc.UpdateDeviceStatus(ctx, r); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("device %s status updated to %s", uid, status)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_delete_device",
			mcp.WithDescription("Permanently delete a ShellHub device. This action is irreversible."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.DeviceRemove) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			tenant := tenantFromCtx(ctx)

			if err := svc.DeleteDevice(ctx, models.UID(uid), tenant); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("device %s deleted", uid)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_rename_device",
			mcp.WithDescription("Rename a ShellHub device."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Device UID.")),
			mcp.WithString("name", mcp.Required(), mcp.Description("New device name (hostname format).")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.DeviceRename) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			name, _ := args["name"].(string)
			tenant := tenantFromCtx(ctx)

			if err := svc.RenameDevice(ctx, models.UID(uid), name, tenant); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("device %s renamed to %s", uid, name)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_stats",
			mcp.WithDescription("Get ShellHub instance statistics: registered/online/pending/rejected devices and active sessions."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}

			r := &requests.GetStats{TenantID: tenantFromCtx(ctx)}
			stats, err := svc.GetStats(ctx, r)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(stats)), nil
		},
	)
}

// --- Session tools ---

func addSessionTools(s *mcpserver.MCPServer, svc services.Service) {
	s.AddTool(
		mcp.NewTool("shellhub_list_sessions",
			mcp.WithDescription("List SSH sessions in the namespace. Returns active and historical sessions."),
			mcp.WithInteger("page", mcp.Description("Page number (1-based). Default: 1.")),
			mcp.WithInteger("per_page", mcp.Description("Results per page (max 100). Default: 20.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.SessionDetails) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			page := intArg(args, "page", 1)
			perPage := intArg(args, "per_page", 20)

			r := &requests.ListSessions{
				TenantID:  tenantFromCtx(ctx),
				Paginator: query.Paginator{Page: page, PerPage: perPage},
			}
			r.Paginator.Normalize()

			sessions, count, err := svc.ListSessions(ctx, r)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(map[string]any{
				"total":    count,
				"sessions": sessions,
			})), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_session",
			mcp.WithDescription("Get details of a specific SSH session by its UID."),
			mcp.WithString("uid", mcp.Required(), mcp.Description("Session UID.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			role, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}
			if !role.HasPermission(authorizer.SessionDetails) {
				return mcpForbidden(), nil
			}

			args := req.GetArguments()
			uid, _ := args["uid"].(string)
			session, err := svc.GetSession(ctx, models.UID(uid))
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(session)), nil
		},
	)
}

// --- Namespace tools ---

func addNamespaceTools(s *mcpserver.MCPServer, svc services.Service) {
	s.AddTool(
		mcp.NewTool("shellhub_list_namespaces",
			mcp.WithDescription("List ShellHub namespaces accessible to the authenticated user."),
			mcp.WithInteger("page", mcp.Description("Page number (1-based). Default: 1.")),
			mcp.WithInteger("per_page", mcp.Description("Results per page. Default: 20.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}

			args := req.GetArguments()
			page := intArg(args, "page", 1)
			perPage := intArg(args, "per_page", 20)

			r := &requests.NamespaceList{
				UserID:    userIDFromCtx(ctx),
				TenantID:  tenantFromCtx(ctx),
				Paginator: query.Paginator{Page: page, PerPage: perPage},
			}
			r.Paginator.Normalize()

			namespaces, count, err := svc.ListNamespaces(ctx, r)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(map[string]any{
				"total":      count,
				"namespaces": namespaces,
			})), nil
		},
	)

	s.AddTool(
		mcp.NewTool("shellhub_get_namespace",
			mcp.WithDescription("Get details and settings of a ShellHub namespace by its tenant ID."),
			mcp.WithString("tenant_id", mcp.Required(), mcp.Description("Namespace tenant ID (UUID).")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, ok := roleFromCtx(ctx)
			if !ok {
				return mcpUnauth(), nil
			}

			args := req.GetArguments()
			tenantID, _ := args["tenant_id"].(string)

			// GetNamespace performs no membership check — only fetch the
			// namespace the caller already belongs to.
			if tenantID != tenantFromCtx(ctx) {
				return mcpForbidden(), nil
			}

			namespace, err := svc.GetNamespace(ctx, tenantID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultText(toJSON(namespace)), nil
		},
	)
}
