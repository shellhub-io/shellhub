package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
)

const publicAPIPrefix = "/api"

func (h *Handler) publicRoutes() []routetable.Route {
	table := []routetable.Route{
		{Method: http.MethodGet, Path: HealthCheckURL, Endpoint: gateway.None(h.EvaluateHealth), Access: routetable.Anonymous, Unbounded: true},

		{Method: http.MethodGet, Path: AuthLocalUserURLV2, Endpoint: gateway.Legacy(h.RenewUserToken), Access: routetable.OwnAccount},
		{Method: http.MethodGet, Path: AuthUserTokenPublicURL, Endpoint: gateway.Legacy(h.RenewUserToken), Access: routetable.OwnAccount},
		{Method: http.MethodPost, Path: AuthDeviceURL, Endpoint: gateway.Legacy(h.AuthDevice), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AuthDeviceChallengeURL, Endpoint: gateway.Legacy(h.CreateDeviceAuthChallenge), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AuthDeviceURLV2, Endpoint: gateway.Legacy(h.AuthDevice), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: EnrollmentCallbackURL, Endpoint: gateway.Legacy(h.EnrollmentCallback), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AuthLocalUserURL, Endpoint: gateway.Legacy(h.AuthLocalUser), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AuthLocalUserURLV2, Endpoint: gateway.Legacy(h.AuthLocalUser), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AuthPublicKeyURL, Endpoint: gateway.Legacy(h.AuthPublicKey), Access: routetable.DeviceToken},

		{Method: http.MethodPost, Path: CreateAPIKeyURL, Endpoint: gateway.Legacy(h.CreateAPIKey), Access: routetable.Requires(authorizer.APIKeyCreate), NoAPIKey: true},
		{Method: http.MethodGet, Path: ListAPIKeysURL, Endpoint: gateway.Legacy(h.ListAPIKeys), Access: routetable.NamespaceMember, NoAPIKey: true},
		{Method: http.MethodPatch, Path: UpdateAPIKeyURL, Endpoint: gateway.None(h.UpdateAPIKey), Access: routetable.Requires(authorizer.APIKeyUpdate)},
		{Method: http.MethodDelete, Path: DeleteAPIKeyURL, Endpoint: gateway.Legacy(h.DeleteAPIKey), Access: routetable.Requires(authorizer.APIKeyDelete), NoAPIKey: true},
		{Method: http.MethodPost, Path: CreateAPIKeySSHIdentityURL, Endpoint: gateway.Legacy(h.CreateAPIKeySSHIdentity), Access: routetable.Requires(authorizer.SSHIdentityManage)},
		{Method: http.MethodGet, Path: ListAPIKeySSHIdentitiesURL, Endpoint: gateway.Legacy(h.ListAPIKeySSHIdentities), Access: routetable.Requires(authorizer.SSHIdentityManage)},

		{Method: http.MethodPost, Path: CreateProvisioningKeyURL, Endpoint: gateway.Legacy(h.CreateProvisioningKey), Access: routetable.Requires(authorizer.ProvisioningKeyCreate), NoAPIKey: true},
		{Method: http.MethodGet, Path: ListProvisioningKeysURL, Endpoint: gateway.Legacy(h.ListProvisioningKeys), Access: routetable.Requires(authorizer.ProvisioningKeyList), NoAPIKey: true},
		{Method: http.MethodPatch, Path: UpdateProvisioningKeyURL, Endpoint: gateway.Legacy(h.UpdateProvisioningKey), Access: routetable.Requires(authorizer.ProvisioningKeyUpdate), NoAPIKey: true},
		{Method: http.MethodGet, Path: RevealProvisioningKeyURL, Endpoint: gateway.Legacy(h.RevealProvisioningKey), Access: routetable.Requires(authorizer.ProvisioningKeyReveal), NoAPIKey: true},
		{Method: http.MethodGet, Path: HistoryProvisioningKeyURL, Endpoint: gateway.Legacy(h.HistoryProvisioningKey), Access: routetable.Requires(authorizer.ProvisioningKeyList), NoAPIKey: true},

		{Method: http.MethodPatch, Path: URLUpdateUser, Endpoint: gateway.Legacy(h.UpdateUser), Access: routetable.OwnAccount},
		{Method: http.MethodPatch, Path: URLDeprecatedUpdateUser, Endpoint: gateway.Legacy(h.UpdateUser), Access: routetable.OwnAccount},
		{Method: http.MethodPatch, Path: URLDeprecatedUpdateUserPassword, Endpoint: gateway.Legacy(h.UpdateUserPassword), Access: routetable.OwnAccount},
		{Method: http.MethodDelete, Path: URLRevokeUserTokens, Endpoint: gateway.Legacy(h.RevokeUserTokens), Access: routetable.OwnAccount},
		{Method: http.MethodPut, Path: URLRecordOnboardingSurvey, Endpoint: gateway.Legacy(h.RecordOnboardingSurvey), Access: routetable.OwnAccount},

		{Method: http.MethodPost, Path: RegisterUserURL, Endpoint: gateway.Legacy(h.RegisterUser), Access: routetable.Anonymous},
		{Method: http.MethodGet, Path: URLResolveInvitation, Endpoint: gateway.Legacy(h.ResolveInvitation), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: URLGenerateInvitationLink, Endpoint: gateway.One(h.GenerateInvitationLink), Access: routetable.Requires(authorizer.NamespaceAddMember)},
		{Method: http.MethodPatch, Path: URLAcceptInvite, Endpoint: gateway.Legacy(h.AcceptInvite), Access: routetable.OwnAccount},
		{Method: http.MethodGet, Path: URLUserMembershipInvitationList, Endpoint: gateway.Legacy(h.GetUserMembershipInvitationList), Access: routetable.OwnAccount},
		{Method: http.MethodGet, Path: URLNamespaceMembershipInvitationList, Endpoint: gateway.Legacy(h.GetNamespaceMembershipInvitationList), Access: routetable.Requires(authorizer.NamespaceEditMember)},
		{Method: http.MethodDelete, Path: URLCancelMembershipInvitation, Endpoint: gateway.None(h.CancelMembershipInvitation), Access: routetable.Requires(authorizer.NamespaceRemoveMember)},

		{Method: http.MethodGet, Path: GetDeviceListURL, Endpoint: gateway.Legacy(h.GetDeviceList), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: GetDeviceURL, Endpoint: gateway.One(h.GetDevice), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: DeviceLocationsURL, Endpoint: gateway.Legacy(h.GetDeviceLocations), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: ResolveDeviceURL, Endpoint: gateway.Legacy(h.ResolveDevice), Access: routetable.NamespaceMember},
		{Method: http.MethodPut, Path: UpdateDevice, Endpoint: gateway.Legacy(h.UpdateDevice), Access: routetable.Requires(authorizer.DeviceUpdate)},
		{Method: http.MethodPatch, Path: RenameDeviceURL, Endpoint: gateway.Legacy(h.RenameDevice), Access: routetable.Requires(authorizer.DeviceRename)},
		{Method: http.MethodGet, Path: TenantRegistrationsURL, Endpoint: gateway.Legacy(h.TenantRegistrations), Access: routetable.Requires(authorizer.DeviceStatusUpdate)},
		{Method: http.MethodPatch, Path: UpdateDeviceStatusURL, Endpoint: gateway.Legacy(h.UpdateDeviceStatus), Access: routetable.Requires(authorizer.DeviceStatusUpdate)},

		{Method: http.MethodPost, Path: CreateDeviceLoginCodeURL, Endpoint: gateway.Legacy(h.CreateDeviceLoginCode), Access: routetable.DeviceToken},
		{Method: http.MethodGet, Path: GetDeviceAuthStatusURL, Endpoint: gateway.Legacy(h.GetDeviceAuthStatus), Access: routetable.DeviceToken},
		{Method: http.MethodGet, Path: ResolveDeviceLoginCodeURL, Endpoint: gateway.Legacy(h.ResolveDeviceLoginCode), Access: routetable.PathCode},

		{Method: http.MethodPost, Path: CreateDevicePairingURL, Endpoint: gateway.Legacy(h.CreateDevicePairing), Access: routetable.Anonymous},
		{Method: http.MethodGet, Path: GetDevicePairingStatusURL, Endpoint: gateway.Legacy(h.GetDevicePairingStatus), Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: AcceptDevicePairingURL, Endpoint: gateway.Legacy(h.AcceptDevicePairing), Access: routetable.PathCode},

		{Method: http.MethodGet, Path: GetSSHApprovalURL, Endpoint: gateway.Legacy(h.GetSSHApproval), Access: routetable.PathCode},
		{Method: http.MethodPost, Path: ConfirmSSHApprovalURL, Endpoint: gateway.Legacy(h.ConfirmSSHApproval), Access: routetable.PathCode},
		{Method: http.MethodPost, Path: RejectSSHApprovalURL, Endpoint: gateway.Legacy(h.RejectSSHApproval), Access: routetable.PathCode},
		{Method: http.MethodDelete, Path: DeleteDeviceURL, Endpoint: gateway.Legacy(h.DeleteDevice), Access: routetable.Requires(authorizer.DeviceRemove)},
		{Method: http.MethodDelete, Path: MakeTeamDeviceURL, Endpoint: gateway.Legacy(h.MakeTeamDevice), Access: routetable.Requires(authorizer.ProvisioningKeyCreate)},
		{Method: http.MethodPut, Path: SetDeviceCustomFieldURL, Endpoint: gateway.Legacy(h.SetDeviceCustomField), Access: routetable.Requires(authorizer.DeviceCustomFieldUpdate)},
		{Method: http.MethodDelete, Path: DeleteDeviceCustomFieldURL, Endpoint: gateway.Legacy(h.DeleteDeviceCustomField), Access: routetable.Requires(authorizer.DeviceCustomFieldUpdate)},

		{Method: http.MethodGet, Path: URLGetTags, Endpoint: gateway.Legacy(h.GetTags), Access: routetable.NamespaceMember},
		{Method: http.MethodPost, Path: URLCreateTag, Endpoint: gateway.Legacy(h.CreateTag), Access: routetable.Requires(authorizer.TagCreate)},
		{Method: http.MethodPatch, Path: URLUpdateTag, Endpoint: gateway.Legacy(h.UpdateTag), Access: routetable.Requires(authorizer.TagUpdate)},
		{Method: http.MethodDelete, Path: URLDeleteTag, Endpoint: gateway.Legacy(h.DeleteTag), Access: routetable.Requires(authorizer.TagDelete)},
		{Method: http.MethodPost, Path: URLPushTagToDevice, Endpoint: gateway.Legacy(h.PushTagToDevice), Access: routetable.Requires(authorizer.TagCreate)},
		{Method: http.MethodDelete, Path: URLPullTagFromDevice, Endpoint: gateway.Legacy(h.PullTagFromDevice), Access: routetable.Requires(authorizer.TagDelete)},

		{Method: http.MethodGet, Path: URLOldGetTags, Endpoint: gateway.Legacy(h.GetTags), Access: routetable.NamespaceMember},
		{Method: http.MethodPost, Path: URLOldCreateTag, Endpoint: gateway.Legacy(h.CreateTag), Access: routetable.Requires(authorizer.TagCreate)},
		{Method: http.MethodPatch, Path: URLOldUpdateTag, Endpoint: gateway.Legacy(h.UpdateTag), Access: routetable.Requires(authorizer.TagUpdate)},
		{Method: http.MethodDelete, Path: URLOldDeleteTag, Endpoint: gateway.Legacy(h.DeleteTag), Access: routetable.Requires(authorizer.TagDelete)},
		{Method: http.MethodPost, Path: URLOldPushTagToDevice, Endpoint: gateway.Legacy(h.PushTagToDevice), Access: routetable.Requires(authorizer.TagCreate)},
		{Method: http.MethodDelete, Path: URLOldPullTagFromDevice, Endpoint: gateway.Legacy(h.PullTagFromDevice), Access: routetable.Requires(authorizer.TagDelete)},

		{Method: http.MethodGet, Path: GetSessionsURL, Endpoint: gateway.Legacy(h.GetSessionList), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: GetSessionURL, Endpoint: gateway.Legacy(h.GetSession), Access: routetable.NamespaceMember},

		{Method: http.MethodGet, Path: GetStatsURL, Endpoint: gateway.One(h.GetStats), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: GetSystemInfoURL, Endpoint: gateway.Legacy(h.GetSystemInfo), Access: routetable.Anonymous},
		{Method: http.MethodGet, Path: GetSystemDownloadInstallScriptURL, Endpoint: gateway.Legacy(h.GetSystemDownloadInstallScript), Access: routetable.Anonymous},

		{Method: http.MethodPost, Path: CreatePublicKeyURL, Endpoint: gateway.Legacy(h.CreatePublicKey), Access: routetable.Requires(authorizer.PublicKeyCreate)},
		{Method: http.MethodGet, Path: GetPublicKeysURL, Endpoint: gateway.Legacy(h.GetPublicKeys), Access: routetable.NamespaceMember},
		{Method: http.MethodPut, Path: UpdatePublicKeyURL, Endpoint: gateway.Legacy(h.UpdatePublicKey), Access: routetable.Requires(authorizer.PublicKeyEdit)},
		{Method: http.MethodDelete, Path: DeletePublicKeyURL, Endpoint: gateway.Legacy(h.DeletePublicKey), Access: routetable.Requires(authorizer.PublicKeyRemove)},

		{Method: http.MethodGet, Path: GetNamespaceURL, Endpoint: gateway.One(h.GetNamespace), Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: ListNamespaceURL, Endpoint: gateway.Legacy(h.GetNamespaceList), Access: routetable.OwnAccount},
		{Method: http.MethodPut, Path: EditNamespaceURL, Endpoint: gateway.One(h.EditNamespace), Access: routetable.Requires(authorizer.NamespaceUpdate)},
		{Method: http.MethodDelete, Path: DeleteNamespaceURL, Endpoint: gateway.None(h.DeleteNamespace), Access: routetable.Requires(authorizer.NamespaceDelete)},

		{Method: http.MethodGet, Path: ListNamespaceMembersURL, Endpoint: gateway.List(h.ListNamespaceMembers), Access: routetable.NamespaceMember},
		{Method: http.MethodPost, Path: AddNamespaceMemberURL, Endpoint: gateway.One(h.AddNamespaceMember), Access: routetable.Requires(authorizer.NamespaceAddMember)},
		{Method: http.MethodPatch, Path: EditNamespaceMemberURL, Endpoint: gateway.None(h.EditNamespaceMember), Access: routetable.Requires(authorizer.NamespaceEditMember)},
		{Method: http.MethodDelete, Path: RemoveNamespaceMemberURL, Endpoint: gateway.One(h.RemoveNamespaceMember), Access: routetable.Requires(authorizer.NamespaceRemoveMember)},
		{Method: http.MethodDelete, Path: LeaveNamespaceURL, Endpoint: gateway.Legacy(h.LeaveNamespace), Access: routetable.OwnAccount},

		{Method: http.MethodPut, Path: EditSessionRecordStatusURL, Endpoint: gateway.None(h.EditSessionRecordStatus), Access: routetable.Requires(authorizer.NamespaceEnableSessionRecord)},
		{Method: http.MethodPost, Path: UpgradeNamespaceURL, Endpoint: gateway.None(h.UpgradeNamespace), Access: routetable.Requires(authorizer.NamespaceUpdate)},

		{Method: http.MethodGet, Path: ListAccessPoliciesURL, Endpoint: gateway.Legacy(h.ListAccessPolicies), Access: routetable.Requires(authorizer.AccessPolicyManage)},
		{Method: http.MethodPost, Path: CreateAccessPolicyURL, Endpoint: gateway.Legacy(h.CreateAccessPolicy), Access: routetable.Requires(authorizer.AccessPolicyManage)},
		{Method: http.MethodGet, Path: GetAccessPolicyURL, Endpoint: gateway.Legacy(h.GetAccessPolicy), Access: routetable.Requires(authorizer.AccessPolicyManage)},
		{Method: http.MethodPut, Path: UpdateAccessPolicyURL, Endpoint: gateway.Legacy(h.UpdateAccessPolicy), Access: routetable.Requires(authorizer.AccessPolicyManage)},
		{Method: http.MethodDelete, Path: DeleteAccessPolicyURL, Endpoint: gateway.Legacy(h.DeleteAccessPolicy), Access: routetable.Requires(authorizer.AccessPolicyManage)},

		{Method: http.MethodGet, Path: ListSSHIdentitiesURL, Endpoint: gateway.Legacy(h.ListSSHIdentities), Access: routetable.OwnAccount},
		{Method: http.MethodPost, Path: CreateSSHIdentityURL, Endpoint: gateway.Legacy(h.CreateSSHIdentity), Access: routetable.Requires(authorizer.SSHIdentityAdd)},
		{Method: http.MethodPatch, Path: UpdateSSHIdentityURL, Endpoint: gateway.Legacy(h.UpdateSSHIdentity), Access: routetable.Requires(authorizer.SSHIdentityAdd, authorizer.SSHIdentityManage)},
		{Method: http.MethodDelete, Path: DeleteSSHIdentityURL, Endpoint: gateway.Legacy(h.DeleteSSHIdentity), Access: routetable.Requires(authorizer.SSHIdentityAdd, authorizer.SSHIdentityManage)},

		{Method: http.MethodPost, Path: WebReauthURL, Endpoint: gateway.Legacy(h.WebReauthVerify), Access: routetable.OwnAccount},
	}

	if h.edition.IsEnterpriseOrCloud() {
		table = append(table, routetable.Route{Method: http.MethodPost, Path: CreateNamespaceURL, Endpoint: gateway.Legacy(h.CreateNamespace), Access: routetable.OwnAccount})
	}

	if !h.edition.IsCloud() {
		table = append(table, routetable.Route{Method: http.MethodPost, Path: SetupEndpoint, Endpoint: gateway.Legacy(h.Setup), Access: routetable.Anonymous})
	}

	return table
}
