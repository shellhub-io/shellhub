package routes

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo-contrib/v5/echoprometheus"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/envs"
	pkgmiddleware "github.com/shellhub-io/shellhub/pkg/middleware"
	"github.com/shellhub-io/shellhub/pkg/websocket"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/sirupsen/logrus"
)

// DefaultHTTPHandlerConfig is what the router needs beyond the service layer. Every field is
// optional; a zero config yields a router with no reporting and no metrics.
type DefaultHTTPHandlerConfig struct {
	// Reporter represents an instance of [*sentry.Client] that should be proper configured to send error messages
	// from the error handler. If it's nil, the error handler will ignore the Sentry client.
	Reporter *sentry.Client
}

// DefaultHTTPHandler creates an HTTP handler, using [github.com/labstack/echo/v5] package, with the default
// configuration required by ShellHub's services, loading the [github.com/shellhub-io/shellhub/server/api/pkg/gateway] into
// the context, and the service layer. The configuration received controls the error reporter and more.
func DefaultHTTPHandler[S any](service S, cfg *DefaultHTTPHandlerConfig) http.Handler {
	server := echo.New()

	server.Binder = handlers.NewBinder()

	server.Validator = handlers.NewValidator()

	server.HTTPErrorHandler = handlers.NewErrors(cfg.Reporter)

	server.IPExtractor = handlers.RealIPExtractor()

	server.Logger = pkgmiddleware.NewSlogLogger(logrus.NewEntry(logrus.StandardLogger()))

	server.Use(echoMiddleware.RequestIDWithConfig(echoMiddleware.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, id string) {
			c.Request().Header.Set(echo.HeaderXRequestID, id)
		},
	}))
	server.Use(echoMiddleware.Secure())
	server.Use(gateway.WithContext(service))
	server.Use(pkgmiddleware.Log)

	return server
}

// Option installs a cross-cutting concern on the router. Returning an error aborts startup,
// so a misconfigured option fails the process rather than serving without it.
type Option func(e *echo.Echo, handler *Handler) error

// WithReporter sends unhandled errors to Sentry through reporter.
func WithReporter(reporter *sentry.Client) Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.HTTPErrorHandler = handlers.NewErrors(reporter)

		return nil
	}
}

// WithMetrics records request metrics and serves them at /metrics.
func WithMetrics() Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.Use(echoprometheus.NewMiddleware("api"))
		e.GET("/metrics", echoprometheus.NewHandler())

		return nil
	}
}

// WithOpenAPIValidator rejects requests that do not match the published spec, so the spec is
// enforced rather than merely documented.
func WithOpenAPIValidator(cfg *routesmiddleware.OpenAPIValidatorConfig) Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.Use(routesmiddleware.OpenAPIValidator(cfg))

		return nil
	}
}

// WithAuthentication authenticates every request in-process, replacing the
// per-route authentication subrequest the edge proxy used to issue.
//
// It is an option rather than a default so handler tests can keep exercising a
// route's authorization with an identity injected straight into the request
// headers. Every production entrypoint must pass it;
// [TestRouterRejectsUncredentialedRequests] guards that the wiring stays in place.
func WithAuthentication(authn *routesmiddleware.Authenticator) Option {
	return func(e *echo.Echo, handler *Handler) error {
		handler.WithAuthenticator(authn)
		e.Use(authn.Middleware)

		return nil
	}
}

// NewRouter builds the API router over service, applying each option in turn.
func NewRouter(service services.Service, opts ...Option) *echo.Echo {
	router, ok := DefaultHTTPHandler(service, new(DefaultHTTPHandlerConfig)).(*echo.Echo)
	if !ok {
		return nil
	}

	handler := NewHandler(service, websocket.NewGorillaWebSocketUpgrader())
	for _, opt := range opts {
		if err := opt(router, handler); err != nil {
			return nil
		}
	}

	publicAPI := router.Group("/api")
	publicAPI.GET(HealthCheckURL,
		gateway.None(handler.EvaluateHealth,
			gateway.Unbounded("the health check reports on the instance, which belongs to no namespace"),
			gateway.Anonymous("the health check is what a load balancer asks before any credential exists")))

	publicAPI.GET(AuthLocalUserURLV2, gateway.Handler(handler.CreateUserToken))                                   // TODO: method POST
	publicAPI.GET(AuthUserTokenPublicURL, gateway.Handler(handler.CreateUserToken), routesmiddleware.BlockAPIKey) // TODO: method POST
	publicAPI.POST(AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(EnrollmentCallbackURL, gateway.Handler(handler.EnrollmentCallback))
	publicAPI.POST(AuthLocalUserURL, gateway.Handler(handler.AuthLocalUser))
	publicAPI.POST(AuthLocalUserURLV2, gateway.Handler(handler.AuthLocalUser))
	publicAPI.POST(AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))

	publicAPI.POST(CreateAPIKeyURL, gateway.Handler(handler.CreateAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyCreate))
	publicAPI.GET(ListAPIKeysURL, gateway.Handler(handler.ListAPIKeys), routesmiddleware.BlockAPIKey)
	publicAPI.PATCH(UpdateAPIKeyURL, gateway.Handler(handler.UpdateAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyUpdate))
	publicAPI.DELETE(DeleteAPIKeyURL, gateway.Handler(handler.DeleteAPIKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.APIKeyDelete))

	publicAPI.POST(CreateInstallKeyURL, gateway.Handler(handler.CreateInstallKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.InstallKeyCreate))
	publicAPI.GET(ListInstallKeysURL, gateway.Handler(handler.ListInstallKeys), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.InstallKeyList))
	publicAPI.PATCH(UpdateInstallKeyURL, gateway.Handler(handler.UpdateInstallKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.InstallKeyUpdate))
	publicAPI.GET(RevealInstallKeyURL, gateway.Handler(handler.RevealInstallKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.InstallKeyReveal))
	publicAPI.GET(HistoryInstallKeyURL, gateway.Handler(handler.HistoryInstallKey), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.InstallKeyList))

	publicAPI.PATCH(URLUpdateUser, gateway.Handler(handler.UpdateUser), routesmiddleware.BlockAPIKey)
	publicAPI.PATCH(URLDeprecatedUpdateUser, gateway.Handler(handler.UpdateUser), routesmiddleware.BlockAPIKey)                 // WARN: DEPRECATED.
	publicAPI.PATCH(URLDeprecatedUpdateUserPassword, gateway.Handler(handler.UpdateUserPassword), routesmiddleware.BlockAPIKey) // WARN: DEPRECATED.

	publicAPI.POST(RegisterUserURL, gateway.Handler(handler.RegisterUser))
	publicAPI.GET(URLResolveInvitation, gateway.Handler(handler.ResolveInvitation))
	publicAPI.POST(URLGenerateInvitationLink, gateway.Handler(handler.GenerateInvitationLink), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.PATCH(URLAcceptInvite, gateway.Handler(handler.AcceptInvite), routesmiddleware.BlockAPIKey)
	publicAPI.GET(URLUserMembershipInvitationList, gateway.Handler(handler.GetUserMembershipInvitationList))
	publicAPI.GET(URLNamespaceMembershipInvitationList, gateway.Handler(handler.GetNamespaceMembershipInvitationList), routesmiddleware.RequiresPermission(authorizer.NamespaceEditMember))
	publicAPI.DELETE(URLCancelMembershipInvitation, gateway.Handler(handler.CancelMembershipInvitation), routesmiddleware.RequiresPermission(authorizer.NamespaceRemoveMember))

	publicAPI.GET(GetDeviceListURL, routesmiddleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(GetDeviceURL, routesmiddleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.GET(ResolveDeviceURL, routesmiddleware.Authorize(gateway.Handler(handler.ResolveDevice)))
	publicAPI.PUT(UpdateDevice, gateway.Handler(handler.UpdateDevice), routesmiddleware.RequiresPermission(authorizer.DeviceUpdate))
	publicAPI.PATCH(RenameDeviceURL, gateway.Handler(handler.RenameDevice), routesmiddleware.RequiresPermission(authorizer.DeviceRename))
	publicAPI.PATCH(UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus), routesmiddleware.RequiresPermission(authorizer.DeviceAccept)) // TODO: DeviceWrite

	publicAPI.POST(CreateDeviceLoginCodeURL, gateway.Handler(handler.CreateDeviceLoginCode))
	publicAPI.GET(GetDeviceAuthStatusURL, gateway.Handler(handler.GetDeviceAuthStatus))
	publicAPI.GET(ResolveDeviceLoginCodeURL, gateway.Handler(handler.ResolveDeviceLoginCode), routesmiddleware.BlockAPIKey)

	publicAPI.POST(CreateDevicePairingURL, gateway.Handler(handler.CreateDevicePairing))
	publicAPI.GET(GetDevicePairingStatusURL, gateway.Handler(handler.GetDevicePairingStatus))
	publicAPI.POST(AcceptDevicePairingURL, gateway.Handler(handler.AcceptDevicePairing), routesmiddleware.BlockAPIKey)
	publicAPI.POST(PrepareDevicePairingURL, gateway.Handler(handler.PrepareDevicePairing), routesmiddleware.BlockAPIKey, routesmiddleware.RequiresPermission(authorizer.DeviceAccept))

	publicAPI.GET(GetSSHApprovalURL, gateway.Handler(handler.GetSSHApproval), routesmiddleware.BlockAPIKey)
	publicAPI.POST(ConfirmSSHApprovalURL, gateway.Handler(handler.ConfirmSSHApproval), routesmiddleware.BlockAPIKey)
	publicAPI.POST(RejectSSHApprovalURL, gateway.Handler(handler.RejectSSHApproval), routesmiddleware.BlockAPIKey)
	publicAPI.DELETE(DeleteDeviceURL, gateway.Handler(handler.DeleteDevice), routesmiddleware.RequiresPermission(authorizer.DeviceRemove))
	publicAPI.PUT(SetDeviceCustomFieldURL, gateway.Handler(handler.SetDeviceCustomField), routesmiddleware.RequiresPermission(authorizer.DeviceCustomFieldUpdate))
	publicAPI.DELETE(DeleteDeviceCustomFieldURL, gateway.Handler(handler.DeleteDeviceCustomField), routesmiddleware.RequiresPermission(authorizer.DeviceCustomFieldUpdate))

	publicAPI.GET(URLGetTags, gateway.Handler(handler.GetTags))
	publicAPI.POST(URLCreateTag, gateway.Handler(handler.CreateTag), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.PATCH(URLUpdateTag, gateway.Handler(handler.UpdateTag), routesmiddleware.RequiresPermission(authorizer.TagUpdate))
	publicAPI.DELETE(URLDeleteTag, gateway.Handler(handler.DeleteTag), routesmiddleware.RequiresPermission(authorizer.TagDelete))
	publicAPI.POST(URLPushTagToDevice, gateway.Handler(handler.PushTagToDevice), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.DELETE(URLPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), routesmiddleware.RequiresPermission(authorizer.TagDelete))

	publicAPI.GET(URLOldGetTags, gateway.Handler(handler.GetTags))
	publicAPI.POST(URLOldCreateTag, gateway.Handler(handler.CreateTag), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.PATCH(URLOldUpdateTag, gateway.Handler(handler.UpdateTag), routesmiddleware.RequiresPermission(authorizer.TagUpdate))
	publicAPI.DELETE(URLOldDeleteTag, gateway.Handler(handler.DeleteTag), routesmiddleware.RequiresPermission(authorizer.TagDelete))
	publicAPI.POST(URLOldPushTagToDevice, gateway.Handler(handler.PushTagToDevice), routesmiddleware.RequiresPermission(authorizer.TagCreate))
	publicAPI.DELETE(URLOldPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), routesmiddleware.RequiresPermission(authorizer.TagDelete))

	publicAPI.GET(GetSessionsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(GetSessionURL, routesmiddleware.Authorize(gateway.Handler(handler.GetSession)))

	publicAPI.GET(GetStatsURL, routesmiddleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))
	publicAPI.GET(GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript))

	publicAPI.POST(CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyCreate))
	publicAPI.GET(GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.PUT(UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyEdit))
	publicAPI.DELETE(DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), routesmiddleware.RequiresPermission(authorizer.PublicKeyRemove))

	if envs.IsEnterpriseOrCloud() {
		publicAPI.POST(CreateNamespaceURL, gateway.Handler(handler.CreateNamespace), routesmiddleware.BlockAPIKey)
	}
	publicAPI.GET(GetNamespaceURL, gateway.Handler(handler.GetNamespace), routesmiddleware.RequiresTenant(ParamNamespaceTenant))
	publicAPI.GET(ListNamespaceURL, gateway.Handler(handler.GetNamespaceList), routesmiddleware.BlockAPIKey)
	publicAPI.PUT(EditNamespaceURL, gateway.Handler(handler.EditNamespace), routesmiddleware.RequiresTenant(ParamNamespaceTenant), routesmiddleware.RequiresPermission(authorizer.NamespaceUpdate))
	publicAPI.DELETE(DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace), routesmiddleware.RequiresTenant(ParamNamespaceTenant), routesmiddleware.RequiresPermission(authorizer.NamespaceDelete))

	publicAPI.GET(ListNamespaceMembersURL, gateway.Handler(handler.ListNamespaceMembers), routesmiddleware.RequiresTenant(ParamNamespaceTenant))
	publicAPI.POST(AddNamespaceMemberURL, gateway.Handler(handler.AddNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.PATCH(EditNamespaceMemberURL, gateway.Handler(handler.EditNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceEditMember))
	publicAPI.DELETE(RemoveNamespaceMemberURL, gateway.Handler(handler.RemoveNamespaceMember), routesmiddleware.RequiresPermission(authorizer.NamespaceRemoveMember))
	publicAPI.DELETE(LeaveNamespaceURL, gateway.Handler(handler.LeaveNamespace), routesmiddleware.BlockAPIKey)

	publicAPI.PUT(EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus), routesmiddleware.RequiresTenant(ParamNamespaceTenant), routesmiddleware.RequiresPermission(authorizer.NamespaceEnableSessionRecord))
	publicAPI.PUT(EditSSHAccessModeURL, gateway.Handler(handler.EditSSHAccessMode), routesmiddleware.RequiresTenant(ParamNamespaceTenant), routesmiddleware.RequiresPermission(authorizer.NamespaceUpdate))

	publicAPI.GET(ListAccessPoliciesURL, gateway.Handler(handler.ListAccessPolicies), routesmiddleware.RequiresPermission(authorizer.AccessPolicyManage))
	publicAPI.POST(CreateAccessPolicyURL, gateway.Handler(handler.CreateAccessPolicy), routesmiddleware.RequiresPermission(authorizer.AccessPolicyManage))
	publicAPI.GET(GetAccessPolicyURL, gateway.Handler(handler.GetAccessPolicy), routesmiddleware.RequiresPermission(authorizer.AccessPolicyManage))
	publicAPI.PUT(UpdateAccessPolicyURL, gateway.Handler(handler.UpdateAccessPolicy), routesmiddleware.RequiresPermission(authorizer.AccessPolicyManage))
	publicAPI.DELETE(DeleteAccessPolicyURL, gateway.Handler(handler.DeleteAccessPolicy), routesmiddleware.RequiresPermission(authorizer.AccessPolicyManage))

	publicAPI.GET(ListSSHIdentitiesURL, gateway.Handler(handler.ListSSHIdentities))
	publicAPI.POST(CreateSSHIdentityURL, gateway.Handler(handler.CreateSSHIdentity), routesmiddleware.RequiresPermission(authorizer.SSHIdentityAdd))
	publicAPI.PATCH(UpdateSSHIdentityURL, gateway.Handler(handler.UpdateSSHIdentity), routesmiddleware.RequiresPermission(authorizer.SSHIdentityAdd))
	publicAPI.DELETE(DeleteSSHIdentityURL, gateway.Handler(handler.DeleteSSHIdentity))

	publicAPI.POST(WebReauthURL, gateway.Handler(handler.WebReauthVerify))

	publicAPI.GET(ListServiceAccountsURL, gateway.Handler(handler.ListServiceAccounts), routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.POST(CreateServiceAccountURL, gateway.Handler(handler.CreateServiceAccount), routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))
	publicAPI.DELETE(DeleteServiceAccountURL, gateway.Handler(handler.DeleteServiceAccount), routesmiddleware.RequiresPermission(authorizer.NamespaceAddMember))

	if !envs.IsCloud() {
		publicAPI.POST(SetupEndpoint, gateway.Handler(handler.Setup))
	}

	SetupMCPRoutes(router)

	registerInternalMetrics(router, handler.authn)

	if handler.authn != nil {
		registerAnonymousRoutes(handler.authn)
	}

	if err := applyExtensions(router, handler.authn, service); err != nil {
		logrus.WithError(err).Error("failed to apply route extensions")
	}

	router.Pre(echoMiddleware.Rewrite(map[string]string{
		"/api/containers":   "/api/devices?connector=true",
		"/api/containers?*": "/api/devices?connector=true&$1",
		"/api/containers/*": "/api/devices/$1",
	}))

	router.Pre(echoMiddleware.Rewrite(rootAliases))

	return router
}
