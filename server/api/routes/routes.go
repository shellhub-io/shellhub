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

	publicAPI := gateway.MountOn(router, router.Group(publicAPIPrefix))

	gateway.GET(publicAPI, HealthCheckURL, gateway.None(handler.EvaluateHealth),
		gateway.Unbounded("the health check reports on the instance, which belongs to no namespace"),
		gateway.Anonymous("the health check is what a load balancer asks before any credential exists"))

	gateway.GET(publicAPI, AuthLocalUserURLV2, gateway.Handler(handler.CreateUserToken), // TODO: method POST
		gateway.PermissionInHandler("the service mints a token only for a namespace the caller already belongs to, which is a membership question and not a role one"))
	gateway.GET(publicAPI, AuthUserTokenPublicURL, gateway.Handler(handler.CreateUserToken), gateway.NoAPIKey(), // TODO: method POST
		gateway.PermissionInHandler("the service mints a token only for a namespace the caller already belongs to, which is a membership question and not a role one"))
	gateway.POST(publicAPI, AuthDeviceURL, gateway.Handler(handler.AuthDevice),
		gateway.Anonymous("a device authenticates with its own credentials, and holds none before this call answers"))
	gateway.POST(publicAPI, AuthDeviceURLV2, gateway.Handler(handler.AuthDevice),
		gateway.Anonymous("a device authenticates with its own credentials, and holds none before this call answers"))
	gateway.POST(publicAPI, EnrollmentCallbackURL, gateway.Handler(handler.EnrollmentCallback),
		gateway.Anonymous("the enrollment provider calls back with the token it was issued, not with a ShellHub credential"))
	gateway.POST(publicAPI, AuthLocalUserURL, gateway.Handler(handler.AuthLocalUser),
		gateway.Anonymous("signing in is what produces a credential, so it cannot demand one"))
	gateway.POST(publicAPI, AuthLocalUserURLV2, gateway.Handler(handler.AuthLocalUser),
		gateway.Anonymous("signing in is what produces a credential, so it cannot demand one"))
	gateway.POST(publicAPI, AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey),
		gateway.NoPermission("the agent authenticates with its own device token, which carries no role, and the service signs with the private key the fingerprint names rather than reading the namespace the token is scoped to"))

	gateway.POST(publicAPI, CreateAPIKeyURL, gateway.Handler(handler.CreateAPIKey), gateway.NoAPIKey(), gateway.Requires(authorizer.APIKeyCreate))
	gateway.GET(publicAPI, ListAPIKeysURL, gateway.List(handler.ListAPIKeys), gateway.Accepts(services.APIKeyQuery), gateway.NoAPIKey(),
		gateway.NoPermission("the listing is bounded to the namespace the caller is scoped to, and names no key's secret"))
	gateway.PATCH(publicAPI, UpdateAPIKeyURL, gateway.Handler(handler.UpdateAPIKey), gateway.NoAPIKey(), gateway.Requires(authorizer.APIKeyUpdate))
	gateway.DELETE(publicAPI, DeleteAPIKeyURL, gateway.Handler(handler.DeleteAPIKey), gateway.NoAPIKey(), gateway.Requires(authorizer.APIKeyDelete))

	gateway.POST(publicAPI, CreateInstallKeyURL, gateway.Handler(handler.CreateInstallKey), gateway.NoAPIKey(), gateway.Requires(authorizer.InstallKeyCreate))
	gateway.GET(publicAPI, ListInstallKeysURL, gateway.List(handler.ListInstallKeys), gateway.Accepts(services.InstallKeyQuery), gateway.NoAPIKey(), gateway.Requires(authorizer.InstallKeyList))
	gateway.PATCH(publicAPI, UpdateInstallKeyURL, gateway.Handler(handler.UpdateInstallKey), gateway.NoAPIKey(), gateway.Requires(authorizer.InstallKeyUpdate))
	gateway.GET(publicAPI, RevealInstallKeyURL, gateway.Handler(handler.RevealInstallKey), gateway.NoAPIKey(), gateway.Requires(authorizer.InstallKeyReveal))
	gateway.GET(publicAPI, HistoryInstallKeyURL, gateway.List(handler.HistoryInstallKey), gateway.Accepts(services.InstallKeyEventQuery), gateway.NoAPIKey(), gateway.Requires(authorizer.InstallKeyList))

	gateway.PATCH(publicAPI, URLUpdateUser, gateway.Handler(handler.UpdateUser), gateway.NoAPIKey(),
		gateway.NoPermission("the caller edits their own account, which is theirs to edit whatever role they hold anywhere"))
	gateway.PATCH(publicAPI, URLDeprecatedUpdateUser, gateway.Handler(handler.UpdateUser), gateway.NoAPIKey(), // WARN: DEPRECATED.
		gateway.NoPermission("the caller edits their own account, which is theirs to edit whatever role they hold anywhere"))
	gateway.PATCH(publicAPI, URLDeprecatedUpdateUserPassword, gateway.Handler(handler.UpdateUserPassword), gateway.NoAPIKey(), // WARN: DEPRECATED.
		gateway.NoPermission("the caller changes their own password, which is theirs to change whatever role they hold anywhere"))

	gateway.POST(publicAPI, RegisterUserURL, gateway.Handler(handler.RegisterUser),
		gateway.Anonymous("registering is what creates the person a credential would name"))
	gateway.GET(publicAPI, URLResolveInvitation, gateway.Handler(handler.ResolveInvitation),
		gateway.Anonymous("an invitee follows the link before holding an account, and the signed invitation is the credential"))
	gateway.POST(publicAPI, URLGenerateInvitationLink, gateway.Handler(handler.GenerateInvitationLink), gateway.NoAPIKey(), gateway.Requires(authorizer.NamespaceAddMember))
	gateway.PATCH(publicAPI, URLAcceptInvite, gateway.Handler(handler.AcceptInvite), gateway.NoAPIKey(),
		gateway.NoPermission("accepting is what grants the role, so the invitee holds none in the namespace yet"))
	gateway.GET(publicAPI, URLUserMembershipInvitationList, gateway.List(handler.GetUserMembershipInvitationList), gateway.Accepts(services.MembershipInvitationQuery),
		gateway.Unbounded("an invitation is addressed to a person across namespaces, and the invitee may belong to none yet"),
		gateway.NoPermission("the listing answers which invitations name the caller, who holds no role in the namespaces inviting them"))
	gateway.GET(publicAPI, URLNamespaceMembershipInvitationList, gateway.List(handler.GetNamespaceMembershipInvitationList), gateway.Accepts(services.MembershipInvitationQuery), gateway.Requires(authorizer.NamespaceEditMember))
	gateway.DELETE(publicAPI, URLCancelMembershipInvitation, gateway.Handler(handler.CancelMembershipInvitation), gateway.Requires(authorizer.NamespaceRemoveMember))

	gateway.GET(publicAPI, GetDeviceListURL, gateway.List(handler.GetDeviceList), gateway.Accepts(services.DeviceQuery), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("every namespace member reads the namespace's devices, and the scope the request is bounded to is what limits the answer"))
	gateway.GET(publicAPI, GetDeviceURL, gateway.One(handler.GetDevice), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("every namespace member reads the namespace's devices, and the scope the request is bounded to is what limits the answer"))
	gateway.GET(publicAPI, ResolveDeviceURL, gateway.Handler(handler.ResolveDevice), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("every namespace member reads the namespace's devices, and the scope the request is bounded to is what limits the answer"))
	gateway.PUT(publicAPI, UpdateDevice, gateway.Handler(handler.UpdateDevice), gateway.Requires(authorizer.DeviceUpdate))
	gateway.PATCH(publicAPI, RenameDeviceURL, gateway.Handler(handler.RenameDevice), gateway.Requires(authorizer.DeviceRename))
	gateway.PATCH(publicAPI, UpdateDeviceStatusURL, gateway.Handler(handler.UpdateDeviceStatus), gateway.Requires(authorizer.DeviceAccept)) // TODO: DeviceWrite

	gateway.POST(publicAPI, CreateDeviceLoginCodeURL, gateway.Handler(handler.CreateDeviceLoginCode),
		gateway.NoPermission("the caller is a device authenticating with its own token, which names no member and carries no role"))
	gateway.GET(publicAPI, GetDeviceAuthStatusURL, gateway.Handler(handler.GetDeviceAuthStatus),
		gateway.NoPermission("the caller is a device polling its own login code, which names no member and carries no role"))
	gateway.GET(publicAPI, ResolveDeviceLoginCodeURL, gateway.Handler(handler.ResolveDeviceLoginCode), gateway.NoAPIKey(),
		gateway.PermissionInHandler("the device's namespace comes from the code in the path and not from the caller's session, so the service is where membership in it is checked — and a code that names a pairing rather than a device answers before there is a namespace to be a member of"))

	gateway.POST(publicAPI, CreateDevicePairingURL, gateway.Handler(handler.CreateDevicePairing),
		gateway.Anonymous("an unpaired agent asks for a pairing code with no namespace behind it yet"))
	gateway.GET(publicAPI, GetDevicePairingStatusURL, gateway.Handler(handler.GetDevicePairingStatus),
		gateway.Anonymous("the agent polls its own pairing code, which is the only secret the call needs"))
	gateway.POST(publicAPI, AcceptDevicePairingURL, gateway.Handler(handler.AcceptDevicePairing), gateway.NoAPIKey(),
		gateway.PermissionInHandler("the namespace being paired into comes from the code in the path and not from the caller's session, so the service is where the role can be checked against it"))
	gateway.POST(publicAPI, PrepareDevicePairingURL, gateway.Handler(handler.PrepareDevicePairing), gateway.NoAPIKey(), gateway.Requires(authorizer.DeviceAccept))

	gateway.GET(publicAPI, GetSSHApprovalURL, gateway.Handler(handler.GetSSHApproval), gateway.NoAPIKey(),
		gateway.PermissionInHandler("the approval names the namespace, not the caller's session, so the service is where membership in it is checked"))
	gateway.POST(publicAPI, ConfirmSSHApprovalURL, gateway.Handler(handler.ConfirmSSHApproval), gateway.NoAPIKey(),
		gateway.PermissionInHandler("the approval names the namespace, not the caller's session, so SessionApprove and SSHIdentityAdd are checked in the service against the role the caller holds there"))
	gateway.POST(publicAPI, RejectSSHApprovalURL, gateway.Handler(handler.RejectSSHApproval), gateway.NoAPIKey(),
		gateway.PermissionInHandler("the approval names the namespace, not the caller's session, so SessionApprove is checked in the service against the role the caller holds there"))
	gateway.DELETE(publicAPI, DeleteDeviceURL, gateway.Handler(handler.DeleteDevice), gateway.Requires(authorizer.DeviceRemove))
	gateway.PUT(publicAPI, SetDeviceCustomFieldURL, gateway.Handler(handler.SetDeviceCustomField), gateway.Requires(authorizer.DeviceCustomFieldUpdate))
	gateway.DELETE(publicAPI, DeleteDeviceCustomFieldURL, gateway.Handler(handler.DeleteDeviceCustomField), gateway.Requires(authorizer.DeviceCustomFieldUpdate))

	gateway.GET(publicAPI, URLGetTags, gateway.List(handler.GetTags), gateway.Accepts(services.TagQuery),
		gateway.NoPermission("every namespace member reads the namespace's tags, and the scope the request is bounded to is what limits the answer"))
	gateway.POST(publicAPI, URLCreateTag, gateway.Handler(handler.CreateTag), gateway.Requires(authorizer.TagCreate))
	gateway.PATCH(publicAPI, URLUpdateTag, gateway.Handler(handler.UpdateTag), gateway.Requires(authorizer.TagUpdate))
	gateway.DELETE(publicAPI, URLDeleteTag, gateway.Handler(handler.DeleteTag), gateway.Requires(authorizer.TagDelete))
	gateway.POST(publicAPI, URLPushTagToDevice, gateway.Handler(handler.PushTagToDevice), gateway.Requires(authorizer.TagCreate))
	gateway.DELETE(publicAPI, URLPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), gateway.Requires(authorizer.TagDelete))

	gateway.GET(publicAPI, URLOldGetTags, gateway.List(handler.GetTags), gateway.Accepts(services.TagQuery),
		gateway.NoPermission("every namespace member reads the namespace's tags, and the scope the request is bounded to is what limits the answer"))
	gateway.POST(publicAPI, URLOldCreateTag, gateway.Handler(handler.CreateTag), gateway.Requires(authorizer.TagCreate))
	gateway.PATCH(publicAPI, URLOldUpdateTag, gateway.Handler(handler.UpdateTag), gateway.Requires(authorizer.TagUpdate))
	gateway.DELETE(publicAPI, URLOldDeleteTag, gateway.Handler(handler.DeleteTag), gateway.Requires(authorizer.TagDelete))
	gateway.POST(publicAPI, URLOldPushTagToDevice, gateway.Handler(handler.PushTagToDevice), gateway.Requires(authorizer.TagCreate))
	gateway.DELETE(publicAPI, URLOldPullTagFromDevice, gateway.Handler(handler.PullTagFromDevice), gateway.Requires(authorizer.TagDelete))

	gateway.GET(publicAPI, GetSessionsURL, gateway.List(handler.GetSessionList), gateway.Accepts(services.SessionQuery), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("every namespace member reads the namespace's sessions, and the scope the request is bounded to is what limits the answer"))
	gateway.GET(publicAPI, GetSessionURL, gateway.Handler(handler.GetSession), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("every namespace member reads the namespace's sessions, and the scope the request is bounded to is what limits the answer"))

	gateway.GET(publicAPI, GetStatsURL, gateway.Handler(handler.GetStats), gateway.Guard(routesmiddleware.Authorize),
		gateway.NoPermission("the counts describe the namespace the request is bounded to, and name nothing a member cannot already read"))
	gateway.GET(publicAPI, GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo),
		gateway.Anonymous("the instance describes itself to a browser that has not signed in yet"))
	gateway.GET(publicAPI, GetSystemDownloadInstallScriptURL, gateway.Handler(handler.GetSystemDownloadInstallScript),
		gateway.Anonymous("the install script is fetched by a shell on a machine that holds no credential"))

	gateway.POST(publicAPI, CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey), gateway.Requires(authorizer.PublicKeyCreate))
	gateway.GET(publicAPI, GetPublicKeysURL, gateway.List(handler.GetPublicKeys), gateway.Accepts(services.PublicKeyQuery),
		gateway.NoPermission("every namespace member reads the namespace's public keys, and the scope the request is bounded to is what limits the answer"))
	gateway.PUT(publicAPI, UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey), gateway.Requires(authorizer.PublicKeyEdit))
	gateway.DELETE(publicAPI, DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey), gateway.Requires(authorizer.PublicKeyRemove))

	if envs.IsEnterpriseOrCloud() {
		gateway.POST(publicAPI, CreateNamespaceURL, gateway.Handler(handler.CreateNamespace), gateway.NoAPIKey(),
			gateway.NoPermission("creating a namespace is prior to holding a role in one, and the caller becomes its owner"))
	}
	gateway.GET(publicAPI, GetNamespaceURL, gateway.Handler(handler.GetNamespace), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)),
		gateway.PermissionInHandler("the handler refuses a user the namespace does not list as a member, which is a membership question and not a role one, and the tenant guard is what bounds a credential naming no user"))
	gateway.GET(publicAPI, ListNamespaceURL, gateway.List(handler.GetNamespaceList), gateway.Accepts(services.NamespaceQuery), gateway.NoAPIKey(),
		gateway.Unbounded("the list answers which namespaces the caller belongs to, and the caller may have selected none"),
		gateway.NoPermission("the answer is about the caller rather than about any one namespace, so there is no namespace to hold a role in"))
	gateway.PUT(publicAPI, EditNamespaceURL, gateway.Handler(handler.EditNamespace), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)), gateway.Requires(authorizer.NamespaceUpdate))
	gateway.DELETE(publicAPI, DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)), gateway.Requires(authorizer.NamespaceDelete))

	gateway.GET(publicAPI, ListNamespaceMembersURL, gateway.List(handler.ListNamespaceMembers), gateway.Accepts(services.MemberQuery), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)),
		gateway.NoPermission("every namespace member reads who else belongs, and the scope the request is bounded to is what limits the answer"))
	gateway.POST(publicAPI, AddNamespaceMemberURL, gateway.Handler(handler.AddNamespaceMember), gateway.Requires(authorizer.NamespaceAddMember))
	gateway.PATCH(publicAPI, EditNamespaceMemberURL, gateway.Handler(handler.EditNamespaceMember), gateway.Requires(authorizer.NamespaceEditMember))
	gateway.DELETE(publicAPI, RemoveNamespaceMemberURL, gateway.Handler(handler.RemoveNamespaceMember), gateway.Requires(authorizer.NamespaceRemoveMember))
	gateway.DELETE(publicAPI, LeaveNamespaceURL, gateway.Handler(handler.LeaveNamespace), gateway.NoAPIKey(),
		gateway.NoPermission("leaving removes the caller's own membership, which needs no authority over anyone else"))

	gateway.PUT(publicAPI, EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)), gateway.Requires(authorizer.NamespaceEnableSessionRecord))
	gateway.PUT(publicAPI, EditSSHAccessModeURL, gateway.Handler(handler.EditSSHAccessMode), gateway.Guard(routesmiddleware.RequiresTenant(ParamNamespaceTenant)), gateway.Requires(authorizer.NamespaceUpdate))

	gateway.GET(publicAPI, ListAccessPoliciesURL, gateway.List(handler.ListAccessPolicies), gateway.Accepts(services.AccessPolicyQuery), gateway.Requires(authorizer.AccessPolicyManage))
	gateway.POST(publicAPI, CreateAccessPolicyURL, gateway.Handler(handler.CreateAccessPolicy), gateway.Requires(authorizer.AccessPolicyManage))
	gateway.GET(publicAPI, GetAccessPolicyURL, gateway.Handler(handler.GetAccessPolicy), gateway.Requires(authorizer.AccessPolicyManage))
	gateway.PUT(publicAPI, UpdateAccessPolicyURL, gateway.Handler(handler.UpdateAccessPolicy), gateway.Requires(authorizer.AccessPolicyManage))
	gateway.DELETE(publicAPI, DeleteAccessPolicyURL, gateway.Handler(handler.DeleteAccessPolicy), gateway.Requires(authorizer.AccessPolicyManage))

	gateway.GET(publicAPI, ListSSHIdentitiesURL, gateway.List(handler.ListSSHIdentities), gateway.Accepts(services.SSHIdentityQuery),
		gateway.PermissionInHandler("the listing narrows to the caller's own identities without SSHIdentityManage, so the permission widens the answer rather than admitting the call"))
	gateway.POST(publicAPI, CreateSSHIdentityURL, gateway.Handler(handler.CreateSSHIdentity), gateway.Requires(authorizer.SSHIdentityAdd))
	gateway.PATCH(publicAPI, UpdateSSHIdentityURL, gateway.Handler(handler.UpdateSSHIdentity), gateway.Requires(authorizer.SSHIdentityAdd))
	gateway.DELETE(publicAPI, DeleteSSHIdentityURL, gateway.Handler(handler.DeleteSSHIdentity),
		gateway.RequiresAny(authorizer.SSHIdentityAdd, authorizer.SSHIdentityManage))

	gateway.POST(publicAPI, WebReauthURL, gateway.Handler(handler.WebReauthVerify),
		gateway.NoPermission("the caller re-proves they are themselves, which is about their own credential and not about a role"))

	gateway.GET(publicAPI, ListServiceAccountsURL, gateway.List(handler.ListServiceAccounts), gateway.Accepts(services.ServiceAccountQuery), gateway.Requires(authorizer.NamespaceAddMember))
	gateway.POST(publicAPI, CreateServiceAccountURL, gateway.Handler(handler.CreateServiceAccount), gateway.Requires(authorizer.NamespaceAddMember))
	gateway.DELETE(publicAPI, DeleteServiceAccountURL, gateway.Handler(handler.DeleteServiceAccount), gateway.Requires(authorizer.NamespaceAddMember))

	if !envs.IsCloud() {
		gateway.POST(publicAPI, SetupEndpoint, gateway.Handler(handler.Setup),
			gateway.Anonymous("the first administrator is created before anyone can hold a credential"))
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
