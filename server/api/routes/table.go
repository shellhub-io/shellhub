package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

const publicAPIPrefix = "/api"

type accessKind int

const (
	accessUnstated accessKind = iota
	accessAnonymous
	accessDevice
	accessSelf
	accessMember
	accessInHandler
	accessPermission
)

type access struct {
	kind       accessKind
	permission authorizer.Permission
}

var (
	anonymous = access{kind: accessAnonymous}
	device    = access{kind: accessDevice}
	self      = access{kind: accessSelf}
	member    = access{kind: accessMember}
	inHandler = access{kind: accessInHandler}
)

func requires(permission authorizer.Permission) access {
	return access{kind: accessPermission, permission: permission}
}

type route struct {
	method      string
	path        string
	endpoint    gateway.Endpoint
	access      access
	unbounded   bool
	noAPIKey    bool
	tenantParam string
	scoped      bool
}

func (r route) handler() echo.HandlerFunc {
	return r.endpoint(gateway.Claims{Anonymous: r.access.kind == accessAnonymous, Unbounded: r.unbounded})
}

func (r route) middleware() []echo.MiddlewareFunc {
	var chain []echo.MiddlewareFunc

	if r.noAPIKey {
		chain = append(chain, routesmiddleware.BlockAPIKey)
	}

	if r.tenantParam != "" {
		chain = append(chain, routesmiddleware.RequiresTenant(r.tenantParam))
	}

	if r.access.kind == accessPermission {
		chain = append(chain, routesmiddleware.RequiresPermission(r.access.permission))
	}

	if r.scoped {
		chain = append(chain, routesmiddleware.Authorize)
	}

	return chain
}

func mount(group *echo.Group, table []route) {
	for _, r := range table {
		group.Add(r.method, r.path, r.handler(), r.middleware()...)
	}
}

func allow(authn *routesmiddleware.Authenticator, table []route) {
	for _, r := range table {
		switch r.access.kind {
		case accessAnonymous:
			authn.AllowAnonymous(r.method, publicAPIPrefix+r.path)
		case accessDevice:
			authn.AllowDevice(r.method, publicAPIPrefix+r.path)
		case accessUnstated, accessSelf, accessMember, accessInHandler, accessPermission:
		}
	}
}

func (h *Handler) publicRoutes(edition envs.Edition) []route {
	table := []route{
		{method: http.MethodGet, path: HealthCheckURL, endpoint: gateway.None(h.EvaluateHealth), access: anonymous, unbounded: true},

		{method: http.MethodGet, path: AuthLocalUserURLV2, endpoint: gateway.Legacy(h.CreateUserToken), access: inHandler},
		{method: http.MethodGet, path: AuthUserTokenPublicURL, endpoint: gateway.Legacy(h.CreateUserToken), access: inHandler, noAPIKey: true},
		{method: http.MethodPost, path: AuthDeviceURL, endpoint: gateway.Legacy(h.AuthDevice), access: anonymous},
		{method: http.MethodPost, path: AuthDeviceChallengeURL, endpoint: gateway.Legacy(h.CreateDeviceAuthChallenge), access: anonymous},
		{method: http.MethodPost, path: AuthDeviceURLV2, endpoint: gateway.Legacy(h.AuthDevice), access: anonymous},
		{method: http.MethodPost, path: EnrollmentCallbackURL, endpoint: gateway.Legacy(h.EnrollmentCallback), access: anonymous},
		{method: http.MethodPost, path: AuthLocalUserURL, endpoint: gateway.Legacy(h.AuthLocalUser), access: anonymous},
		{method: http.MethodPost, path: AuthLocalUserURLV2, endpoint: gateway.Legacy(h.AuthLocalUser), access: anonymous},
		{method: http.MethodPost, path: AuthPublicKeyURL, endpoint: gateway.Legacy(h.AuthPublicKey), access: device},

		{method: http.MethodPost, path: CreateAPIKeyURL, endpoint: gateway.Legacy(h.CreateAPIKey), access: requires(authorizer.APIKeyCreate), noAPIKey: true},
		{method: http.MethodGet, path: ListAPIKeysURL, endpoint: gateway.Legacy(h.ListAPIKeys), access: member, noAPIKey: true},
		{method: http.MethodPatch, path: UpdateAPIKeyURL, endpoint: gateway.Legacy(h.UpdateAPIKey), access: requires(authorizer.APIKeyUpdate), noAPIKey: true},
		{method: http.MethodDelete, path: DeleteAPIKeyURL, endpoint: gateway.Legacy(h.DeleteAPIKey), access: requires(authorizer.APIKeyDelete), noAPIKey: true},
		{method: http.MethodPost, path: CreateAPIKeySSHIdentityURL, endpoint: gateway.Legacy(h.CreateAPIKeySSHIdentity), access: requires(authorizer.SSHIdentityManage)},
		{method: http.MethodGet, path: ListAPIKeySSHIdentitiesURL, endpoint: gateway.Legacy(h.ListAPIKeySSHIdentities), access: requires(authorizer.SSHIdentityManage)},

		{method: http.MethodPost, path: CreateProvisioningKeyURL, endpoint: gateway.Legacy(h.CreateProvisioningKey), access: requires(authorizer.ProvisioningKeyCreate), noAPIKey: true},
		{method: http.MethodGet, path: ListProvisioningKeysURL, endpoint: gateway.Legacy(h.ListProvisioningKeys), access: requires(authorizer.ProvisioningKeyList), noAPIKey: true},
		{method: http.MethodPatch, path: UpdateProvisioningKeyURL, endpoint: gateway.Legacy(h.UpdateProvisioningKey), access: requires(authorizer.ProvisioningKeyUpdate), noAPIKey: true},
		{method: http.MethodGet, path: RevealProvisioningKeyURL, endpoint: gateway.Legacy(h.RevealProvisioningKey), access: requires(authorizer.ProvisioningKeyReveal), noAPIKey: true},
		{method: http.MethodGet, path: HistoryProvisioningKeyURL, endpoint: gateway.Legacy(h.HistoryProvisioningKey), access: requires(authorizer.ProvisioningKeyList), noAPIKey: true},

		{method: http.MethodPatch, path: URLUpdateUser, endpoint: gateway.Legacy(h.UpdateUser), access: self, noAPIKey: true},
		{method: http.MethodPatch, path: URLDeprecatedUpdateUser, endpoint: gateway.Legacy(h.UpdateUser), access: self, noAPIKey: true},
		{method: http.MethodPatch, path: URLDeprecatedUpdateUserPassword, endpoint: gateway.Legacy(h.UpdateUserPassword), access: self, noAPIKey: true},
		{method: http.MethodDelete, path: URLRevokeUserTokens, endpoint: gateway.Legacy(h.RevokeUserTokens), access: self, noAPIKey: true},
		{method: http.MethodPut, path: URLRecordOnboardingSurvey, endpoint: gateway.Legacy(h.RecordOnboardingSurvey), access: self, noAPIKey: true},

		{method: http.MethodPost, path: RegisterUserURL, endpoint: gateway.Legacy(h.RegisterUser), access: anonymous},
		{method: http.MethodGet, path: URLResolveInvitation, endpoint: gateway.Legacy(h.ResolveInvitation), access: anonymous},
		{method: http.MethodPost, path: URLGenerateInvitationLink, endpoint: gateway.Legacy(h.GenerateInvitationLink), access: requires(authorizer.NamespaceAddMember), noAPIKey: true},
		{method: http.MethodPatch, path: URLAcceptInvite, endpoint: gateway.Legacy(h.AcceptInvite), access: self, noAPIKey: true},
		{method: http.MethodGet, path: URLUserMembershipInvitationList, endpoint: gateway.Legacy(h.GetUserMembershipInvitationList), access: self},
		{method: http.MethodGet, path: URLNamespaceMembershipInvitationList, endpoint: gateway.Legacy(h.GetNamespaceMembershipInvitationList), access: requires(authorizer.NamespaceEditMember)},
		{method: http.MethodDelete, path: URLCancelMembershipInvitation, endpoint: gateway.Legacy(h.CancelMembershipInvitation), access: requires(authorizer.NamespaceRemoveMember)},

		{method: http.MethodGet, path: GetDeviceListURL, endpoint: gateway.Legacy(h.GetDeviceList), access: member, scoped: true},
		{method: http.MethodGet, path: GetDeviceURL, endpoint: gateway.One(h.GetDevice), access: member, scoped: true},
		{method: http.MethodGet, path: ResolveDeviceURL, endpoint: gateway.Legacy(h.ResolveDevice), access: member, scoped: true},
		{method: http.MethodPut, path: UpdateDevice, endpoint: gateway.Legacy(h.UpdateDevice), access: requires(authorizer.DeviceUpdate)},
		{method: http.MethodPatch, path: RenameDeviceURL, endpoint: gateway.Legacy(h.RenameDevice), access: requires(authorizer.DeviceRename)},
		{method: http.MethodGet, path: TenantRegistrationsURL, endpoint: gateway.Legacy(h.TenantRegistrations), access: requires(authorizer.DeviceStatusUpdate)},
		{method: http.MethodPatch, path: UpdateDeviceStatusURL, endpoint: gateway.Legacy(h.UpdateDeviceStatus), access: requires(authorizer.DeviceStatusUpdate)},

		{method: http.MethodPost, path: CreateDeviceLoginCodeURL, endpoint: gateway.Legacy(h.CreateDeviceLoginCode), access: device},
		{method: http.MethodGet, path: GetDeviceAuthStatusURL, endpoint: gateway.Legacy(h.GetDeviceAuthStatus), access: device},
		{method: http.MethodGet, path: ResolveDeviceLoginCodeURL, endpoint: gateway.Legacy(h.ResolveDeviceLoginCode), access: inHandler, noAPIKey: true},

		{method: http.MethodPost, path: CreateDevicePairingURL, endpoint: gateway.Legacy(h.CreateDevicePairing), access: anonymous},
		{method: http.MethodGet, path: GetDevicePairingStatusURL, endpoint: gateway.Legacy(h.GetDevicePairingStatus), access: anonymous},
		{method: http.MethodPost, path: AcceptDevicePairingURL, endpoint: gateway.Legacy(h.AcceptDevicePairing), access: inHandler, noAPIKey: true},

		{method: http.MethodGet, path: GetSSHApprovalURL, endpoint: gateway.Legacy(h.GetSSHApproval), access: inHandler, noAPIKey: true},
		{method: http.MethodPost, path: ConfirmSSHApprovalURL, endpoint: gateway.Legacy(h.ConfirmSSHApproval), access: inHandler, noAPIKey: true},
		{method: http.MethodPost, path: RejectSSHApprovalURL, endpoint: gateway.Legacy(h.RejectSSHApproval), access: inHandler, noAPIKey: true},
		{method: http.MethodDelete, path: DeleteDeviceURL, endpoint: gateway.Legacy(h.DeleteDevice), access: requires(authorizer.DeviceRemove)},
		{method: http.MethodDelete, path: MakeTeamDeviceURL, endpoint: gateway.Legacy(h.MakeTeamDevice), access: requires(authorizer.ProvisioningKeyCreate)},
		{method: http.MethodPut, path: SetDeviceCustomFieldURL, endpoint: gateway.Legacy(h.SetDeviceCustomField), access: requires(authorizer.DeviceCustomFieldUpdate)},
		{method: http.MethodDelete, path: DeleteDeviceCustomFieldURL, endpoint: gateway.Legacy(h.DeleteDeviceCustomField), access: requires(authorizer.DeviceCustomFieldUpdate)},

		{method: http.MethodGet, path: URLGetTags, endpoint: gateway.Legacy(h.GetTags), access: member},
		{method: http.MethodPost, path: URLCreateTag, endpoint: gateway.Legacy(h.CreateTag), access: requires(authorizer.TagCreate)},
		{method: http.MethodPatch, path: URLUpdateTag, endpoint: gateway.Legacy(h.UpdateTag), access: requires(authorizer.TagUpdate)},
		{method: http.MethodDelete, path: URLDeleteTag, endpoint: gateway.Legacy(h.DeleteTag), access: requires(authorizer.TagDelete)},
		{method: http.MethodPost, path: URLPushTagToDevice, endpoint: gateway.Legacy(h.PushTagToDevice), access: requires(authorizer.TagCreate)},
		{method: http.MethodDelete, path: URLPullTagFromDevice, endpoint: gateway.Legacy(h.PullTagFromDevice), access: requires(authorizer.TagDelete)},

		{method: http.MethodGet, path: URLOldGetTags, endpoint: gateway.Legacy(h.GetTags), access: member, tenantParam: ParamNamespaceTenant},
		{method: http.MethodPost, path: URLOldCreateTag, endpoint: gateway.Legacy(h.CreateTag), access: requires(authorizer.TagCreate)},
		{method: http.MethodPatch, path: URLOldUpdateTag, endpoint: gateway.Legacy(h.UpdateTag), access: requires(authorizer.TagUpdate)},
		{method: http.MethodDelete, path: URLOldDeleteTag, endpoint: gateway.Legacy(h.DeleteTag), access: requires(authorizer.TagDelete)},
		{method: http.MethodPost, path: URLOldPushTagToDevice, endpoint: gateway.Legacy(h.PushTagToDevice), access: requires(authorizer.TagCreate)},
		{method: http.MethodDelete, path: URLOldPullTagFromDevice, endpoint: gateway.Legacy(h.PullTagFromDevice), access: requires(authorizer.TagDelete)},

		{method: http.MethodGet, path: GetSessionsURL, endpoint: gateway.Legacy(h.GetSessionList), access: member, scoped: true},
		{method: http.MethodGet, path: GetSessionURL, endpoint: gateway.Legacy(h.GetSession), access: member, scoped: true},

		{method: http.MethodGet, path: GetStatsURL, endpoint: gateway.Legacy(h.GetStats), access: member, scoped: true},
		{method: http.MethodGet, path: GetSystemInfoURL, endpoint: gateway.Legacy(h.GetSystemInfo), access: anonymous},
		{method: http.MethodGet, path: GetSystemDownloadInstallScriptURL, endpoint: gateway.Legacy(h.GetSystemDownloadInstallScript), access: anonymous},

		{method: http.MethodPost, path: CreatePublicKeyURL, endpoint: gateway.Legacy(h.CreatePublicKey), access: requires(authorizer.PublicKeyCreate)},
		{method: http.MethodGet, path: GetPublicKeysURL, endpoint: gateway.Legacy(h.GetPublicKeys), access: member},
		{method: http.MethodPut, path: UpdatePublicKeyURL, endpoint: gateway.Legacy(h.UpdatePublicKey), access: requires(authorizer.PublicKeyEdit)},
		{method: http.MethodDelete, path: DeletePublicKeyURL, endpoint: gateway.Legacy(h.DeletePublicKey), access: requires(authorizer.PublicKeyRemove)},

		{method: http.MethodGet, path: GetNamespaceURL, endpoint: gateway.Legacy(h.GetNamespace), access: inHandler, tenantParam: ParamNamespaceTenant},
		{method: http.MethodGet, path: ListNamespaceURL, endpoint: gateway.Legacy(h.GetNamespaceList), access: self, noAPIKey: true},
		{method: http.MethodPut, path: EditNamespaceURL, endpoint: gateway.Legacy(h.EditNamespace), access: requires(authorizer.NamespaceUpdate), tenantParam: ParamNamespaceTenant},
		{method: http.MethodDelete, path: DeleteNamespaceURL, endpoint: gateway.Legacy(h.DeleteNamespace), access: requires(authorizer.NamespaceDelete), tenantParam: ParamNamespaceTenant},

		{method: http.MethodGet, path: ListNamespaceMembersURL, endpoint: gateway.Legacy(h.ListNamespaceMembers), access: member, tenantParam: ParamNamespaceTenant},
		{method: http.MethodPost, path: AddNamespaceMemberURL, endpoint: gateway.Legacy(h.AddNamespaceMember), access: requires(authorizer.NamespaceAddMember)},
		{method: http.MethodPatch, path: EditNamespaceMemberURL, endpoint: gateway.Legacy(h.EditNamespaceMember), access: requires(authorizer.NamespaceEditMember)},
		{method: http.MethodDelete, path: RemoveNamespaceMemberURL, endpoint: gateway.Legacy(h.RemoveNamespaceMember), access: requires(authorizer.NamespaceRemoveMember)},
		{method: http.MethodDelete, path: LeaveNamespaceURL, endpoint: gateway.Legacy(h.LeaveNamespace), access: self, noAPIKey: true},

		{method: http.MethodPut, path: EditSessionRecordStatusURL, endpoint: gateway.Legacy(h.EditSessionRecordStatus), access: requires(authorizer.NamespaceEnableSessionRecord), tenantParam: ParamNamespaceTenant},
		{method: http.MethodPost, path: UpgradeNamespaceURL, endpoint: gateway.Legacy(h.UpgradeNamespace), access: requires(authorizer.NamespaceUpdate), noAPIKey: true, tenantParam: ParamNamespaceTenant},

		{method: http.MethodGet, path: ListAccessPoliciesURL, endpoint: gateway.Legacy(h.ListAccessPolicies), access: requires(authorizer.AccessPolicyManage)},
		{method: http.MethodPost, path: CreateAccessPolicyURL, endpoint: gateway.Legacy(h.CreateAccessPolicy), access: requires(authorizer.AccessPolicyManage)},
		{method: http.MethodGet, path: GetAccessPolicyURL, endpoint: gateway.Legacy(h.GetAccessPolicy), access: requires(authorizer.AccessPolicyManage)},
		{method: http.MethodPut, path: UpdateAccessPolicyURL, endpoint: gateway.Legacy(h.UpdateAccessPolicy), access: requires(authorizer.AccessPolicyManage)},
		{method: http.MethodDelete, path: DeleteAccessPolicyURL, endpoint: gateway.Legacy(h.DeleteAccessPolicy), access: requires(authorizer.AccessPolicyManage)},

		{method: http.MethodGet, path: ListSSHIdentitiesURL, endpoint: gateway.Legacy(h.ListSSHIdentities), access: inHandler},
		{method: http.MethodPost, path: CreateSSHIdentityURL, endpoint: gateway.Legacy(h.CreateSSHIdentity), access: requires(authorizer.SSHIdentityAdd)},
		{method: http.MethodPatch, path: UpdateSSHIdentityURL, endpoint: gateway.Legacy(h.UpdateSSHIdentity), access: requires(authorizer.SSHIdentityAdd)},
		{method: http.MethodDelete, path: DeleteSSHIdentityURL, endpoint: gateway.Legacy(h.DeleteSSHIdentity), access: inHandler},

		{method: http.MethodPost, path: WebReauthURL, endpoint: gateway.Legacy(h.WebReauthVerify), access: self},
	}

	if edition.IsEnterpriseOrCloud() {
		table = append(table, route{method: http.MethodPost, path: CreateNamespaceURL, endpoint: gateway.Legacy(h.CreateNamespace), access: self, noAPIKey: true})
	}

	if !edition.IsCloud() {
		table = append(table, route{method: http.MethodPost, path: SetupEndpoint, endpoint: gateway.Legacy(h.Setup), access: anonymous})
	}

	return table
}
