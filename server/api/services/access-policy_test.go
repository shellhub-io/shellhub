package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "user1"
		deviceID = "device1"
	)

	device := &models.Device{UID: deviceID, Name: "web-01", TenantID: tenantID, Taggable: models.Taggable{TagIDs: []string{"tag-web"}}}

	namespaceWith := func(role authorizer.Role) *models.Namespace {
		return &models.Namespace{
			TenantID: tenantID,
			Members:  []models.Member{{ID: userID, Role: role}},
		}
	}

	cases := []struct {
		description     string
		login           string
		sourceIP        string
		requireMocks    func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedAllowed bool
		expectedReauth  bool
		expectedReason  models.DenialReason
		expectedPolicy  string
		expectedLogin   string
		expectedErr     bool
	}{
		{
			description: "denies when the device cannot be resolved",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedAllowed: false,
			expectedErr:     true,
		},
		{
			description: "denies when the user is not a member of the namespace",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{TenantID: tenantID}, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNotAMember,
			expectedErr:     false,
		},
		{
			description: "denies a role without the connect permission before reading any policy",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleObserver), nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonRoleCannotConnect,
			expectedErr:     false,
		},
		{
			description: "grants a service account although its role holds no permissions",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{
						TenantID: tenantID,
						Members:  []models.Member{{ID: userID, Role: authorizer.RoleService, Type: models.UserTypeService}},
					}, nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "fails closed when the policy store errors",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return(nil, 0, errors.New("boom", "store", 0)).Once()
			},
			expectedAllowed: false,
			expectedErr:     true,
		},
		{
			description: "denies by default when there are no policies",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{}, 0, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "grants when an all-members policy grants the login",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOperator), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "denies when the login is outside the policy's login list",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"deploy"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "grants when the login is explicitly listed",
			login:       "deploy",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"deploy"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "denies when the role subject does not match the user's role",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOperator), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectRole, Value: "administrator"},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "grants when the role subject matches the user's role",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleAdministrator), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectRole, Value: "administrator"},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "grants when the user subject matches the user id",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "denies when the filter selects a different device by tag",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{"tag-db"}}},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "grants when the tag filter selects the device",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{"tag-web"}}},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "grants and flags re-auth when the matched policy requires it",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"*"},
							RequireReauth: true,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  true,
			expectedErr:     false,
		},
		{
			description: "requires re-auth when a narrower allow adds it despite a broad no-reauth allow",
			login:       "deploy",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"deploy"},
							Action:        models.PolicyActionAllow,
							RequireReauth: true,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  true,
			expectedErr:     false,
		},
		{
			description: "requires re-auth regardless of the allow order",
			login:       "deploy",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"deploy"},
							Action:        models.PolicyActionAllow,
							RequireReauth: true,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  true,
			expectedErr:     false,
		},
		{
			description: "does not require re-auth for a login the reauth allow does not cover",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"deploy"},
							Action:        models.PolicyActionAllow,
							RequireReauth: true,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  false,
			expectedErr:     false,
		},
		{
			description: "does not flag re-auth for a service account even when the policy requires it",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{
						TenantID: tenantID,
						Members:  []models.Member{{ID: userID, Role: authorizer.RoleObserver, Type: models.UserTypeService}},
					}, nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"*"},
							RequireReauth: true,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  false,
			expectedErr:     false,
		},
		{
			description: "skips a policy with a broken hostname regexp and stays default-deny",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{Hostname: "["},
							Logins:  []string{"*"},
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "deny wins over a matching allow",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Name:    "block root",
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Action:  models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonDeniedByPolicy,
			expectedPolicy:  "block root",
			expectedErr:     false,
		},
		{
			description: "deny does not fire for a login outside its list; allow still grants",
			login:       "teste",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Action:  models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "deny with a wildcard login blocks every login",
			login:       "anything",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Name:    "block everything",
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonDeniedByPolicy,
			expectedPolicy:  "block everything",
			expectedErr:     false,
		},
		{
			description: "deny with a broken hostname regexp fails closed",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Name:    "broken deny",
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{Hostname: "["},
							Logins:  []string{"*"},
							Action:  models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonPolicyUnevaluable,
			expectedPolicy:  "broken deny",
			expectedErr:     false,
		},
		{
			description: "denies when only a matching deny policy exists",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Name:    "block root",
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Action:  models.PolicyActionDeny,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonDeniedByPolicy,
			expectedPolicy:  "block root",
			expectedErr:     false,
		},
		{
			description: "a deny-all blocks even a specific allow for the same subject",
			login:       "teste",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Name:    "deny all",
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionDeny,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"teste"},
							Action:  models.PolicyActionAllow,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonDeniedByPolicy,
			expectedPolicy:  "deny all",
			expectedErr:     false,
		},
		{
			description: "allow grants when the client IP is inside the source CIDR",
			login:       "root",
			sourceIP:    "10.1.2.3",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"10.0.0.0/8"},
							Action:   models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "allow does not grant when the client IP is outside the source CIDR",
			login:       "root",
			sourceIP:    "192.168.1.1",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"10.0.0.0/8"},
							Action:   models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
		{
			description: "empty source IP matches any client IP",
			login:       "root",
			sourceIP:    "203.0.113.9",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "deny fires when the client IP is inside the deny source CIDR",
			login:       "root",
			sourceIP:    "203.0.113.9",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Name:     "block office range",
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"203.0.113.0/24"},
							Action:   models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonDeniedByPolicy,
			expectedPolicy:  "block office range",
			expectedErr:     false,
		},
		{
			description: "deny does not fire when the client IP is outside the deny source CIDR",
			login:       "root",
			sourceIP:    "10.0.0.5",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"203.0.113.0/24"},
							Action:   models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "allow grants for a specific host /32",
			login:       "root",
			sourceIP:    "203.0.113.9",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"203.0.113.9/32"},
							Action:   models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "allow grants when the client IP is in any of multiple source CIDRs",
			login:       "root",
			sourceIP:    "192.168.5.5",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"10.0.0.0/8", "192.168.0.0/16"},
							Action:   models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedErr:     false,
		},
		{
			description: "deny with a source IP fails closed on an unparseable client IP",
			login:       "root",
			sourceIP:    "not-an-ip",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionAllow,
						},
						{
							Name:     "block office range",
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"10.0.0.0/8"},
							Action:   models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonPolicyUnevaluable,
			expectedPolicy:  "block office range",
			expectedErr:     false,
		},
		{
			description: "allow with an invalid source CIDR is skipped and stays default-deny",
			login:       "root",
			sourceIP:    "10.0.0.1",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"garbage"},
							Action:   models.PolicyActionAllow,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  models.ReasonNoGrant,
			expectedLogin:   "root",
			expectedErr:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			decision, err := service.Authorize(ctx, tenantID, userID, deviceID, tc.login, tc.sourceIP)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAllowed, decision.Allowed)
				require.Equal(t, tc.expectedReauth, decision.RequireReauth)
				require.Equal(t, tc.expectedReason, decision.Reason)
				require.Equal(t, tc.expectedPolicy, decision.PolicyName)
				require.Equal(t, tc.expectedLogin, decision.Login)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestNormalizeSourceIPs(t *testing.T) {
	cases := []struct {
		description string
		in          []string
		expected    []string
	}{
		{
			description: "nil yields empty",
			in:          nil,
			expected:    []string{},
		},
		{
			description: "a bare IPv4 becomes a /32 host route",
			in:          []string{"203.0.113.5"},
			expected:    []string{"203.0.113.5/32"},
		},
		{
			description: "a bare IPv6 becomes a /128 host route",
			in:          []string{"2001:db8::1"},
			expected:    []string{"2001:db8::1/128"},
		},
		{
			description: "an existing CIDR passes through unchanged",
			in:          []string{"10.0.0.0/8"},
			expected:    []string{"10.0.0.0/8"},
		},
		{
			description: "a mix of bare IP and CIDR normalizes only the bare IP",
			in:          []string{"10.0.0.0/8", "192.168.1.1"},
			expected:    []string{"10.0.0.0/8", "192.168.1.1/32"},
		},
		{
			description: "surrounding whitespace is trimmed",
			in:          []string{" 1.2.3.4 "},
			expected:    []string{"1.2.3.4/32"},
		},
		{
			description: "empty entries are dropped",
			in:          []string{"", "   ", "1.2.3.4"},
			expected:    []string{"1.2.3.4/32"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.Equal(t, tc.expected, normalizeSourceIPs(tc.in))
		})
	}
}

func TestSubjectMatches(t *testing.T) {
	const (
		saID  = "00000000-0000-0000-0000-00000000000a"
		keyID = "c629572a-b643-4301-90fe-4572b00d007e"
	)

	cases := []struct {
		description string
		subject     models.PolicySubject
		principal   models.Principal
		role        authorizer.Role
		expected    bool
	}{
		{
			description: "all-members matches a person",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			principal:   models.Principal{Kind: models.PrincipalUser, ID: "human-id"},
			role:        authorizer.RoleObserver,
			expected:    true,
		},
		{
			description: "all-members does NOT match a service account (footgun)",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			principal:   models.Principal{Kind: models.PrincipalService, ID: saID},
			role:        authorizer.RoleService,
			expected:    false,
		},
		{
			description: "all-members does NOT match an API key, which is not a member",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			principal:   models.Principal{Kind: models.PrincipalAPIKey, ID: keyID},
			role:        authorizer.RoleAdministrator,
			expected:    false,
		},
		{
			description: "a human role subject does not match a service account",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "observer"},
			principal:   models.Principal{Kind: models.PrincipalService, ID: saID},
			role:        authorizer.RoleService,
			expected:    false,
		},
		{
			description: "a role subject never matches an API key, whatever role the key holds",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "administrator"},
			principal:   models.Principal{Kind: models.PrincipalAPIKey, ID: keyID},
			role:        authorizer.RoleAdministrator,
			expected:    false,
		},
		{
			description: "role=service matches a service account",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "service"},
			principal:   models.Principal{Kind: models.PrincipalService, ID: saID},
			role:        authorizer.RoleService,
			expected:    false,
		},
		{
			description: "role=service does not match a human observer",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "service"},
			principal:   models.Principal{Kind: models.PrincipalUser, ID: "human-id"},
			role:        authorizer.RoleObserver,
			expected:    false,
		},
		{
			description: "user subject matches a service account by id",
			subject:     models.PolicySubject{Type: models.PolicySubjectUser, Value: saID},
			principal:   models.Principal{Kind: models.PrincipalService, ID: saID},
			role:        authorizer.RoleService,
			expected:    true,
		},
		{
			description: "an api-key subject matches that key",
			subject:     models.PolicySubject{Type: models.PolicySubjectAPIKey, Value: keyID},
			principal:   models.Principal{Kind: models.PrincipalAPIKey, ID: keyID},
			role:        authorizer.RoleAdministrator,
			expected:    true,
		},
		{
			description: "an api-key subject matches nothing else",
			subject:     models.PolicySubject{Type: models.PolicySubjectAPIKey, Value: keyID},
			principal:   models.Principal{Kind: models.PrincipalAPIKey, ID: "another-key"},
			role:        authorizer.RoleAdministrator,
			expected:    false,
		},
		{
			description: "a user subject does not match a key that happens to share the id",
			subject:     models.PolicySubject{Type: models.PolicySubjectUser, Value: keyID},
			principal:   models.Principal{Kind: models.PrincipalAPIKey, ID: keyID},
			role:        authorizer.RoleAdministrator,
			expected:    false,
		},
		{
			description: "empty kind is treated as a person for all-members",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			principal:   models.Principal{Kind: models.PrincipalUser, ID: "legacy-id"},
			role:        authorizer.RoleObserver,
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.Equal(t, tc.expected, subjectMatches(tc.subject, tc.principal, tc.role))
		})
	}
}

func TestStricterReauthPeriod(t *testing.T) {
	ptr := func(n int) *int { return &n }

	cases := []struct {
		description string
		a, b        *int
		expected    *int
	}{
		{"nil (every session) wins over a concrete window", nil, ptr(3600), nil},
		{"a concrete window loses to nil in either position", ptr(3600), nil, nil},
		{"both nil stays every session", nil, nil, nil},
		{"the shorter window wins", ptr(3600), ptr(60), ptr(60)},
		{"order does not matter between concrete windows", ptr(60), ptr(3600), ptr(60)},
		{"equal windows return that window", ptr(3600), ptr(3600), ptr(3600)},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.Equal(t, tc.expected, stricterReauthPeriod(tc.a, tc.b))
		})
	}
}

func TestCreateAccessPolicyValidatesTheSubject(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID   = "00000000-0000-4000-0000-000000000000"
		memberID   = "11111111-1111-4111-1111-111111111111"
		serviceID  = "22222222-2222-4222-2222-222222222222"
		strangerID = "33333333-3333-4333-3333-333333333333"
	)

	namespace := &models.Namespace{
		TenantID: tenantID,
		Members: []models.Member{
			{ID: memberID, Role: authorizer.RoleOperator, Type: models.UserTypeHuman},
			{ID: serviceID, Role: authorizer.RoleService, Type: models.UserTypeService},
		},
	}

	cases := []struct {
		description string
		subject     requests.AccessPolicySubject
		rejected    bool
	}{
		{"a member of the namespace is accepted", requests.AccessPolicySubject{Type: "user", Value: memberID}, false},
		{"a service account is accepted, since it is a member too", requests.AccessPolicySubject{Type: "user", Value: serviceID}, false},
		{"a user of another namespace is rejected", requests.AccessPolicySubject{Type: "user", Value: strangerID}, true},
		{"a role the authorizer defines is accepted", requests.AccessPolicySubject{Type: "role", Value: "operator"}, false},
		{"owner is accepted, though a member cannot be assigned it", requests.AccessPolicySubject{Type: "role", Value: "owner"}, false},
		{"an invented role is rejected", requests.AccessPolicySubject{Type: "role", Value: "superadmin"}, true},
		{"all-members with no value is accepted", requests.AccessPolicySubject{Type: "all-members"}, false},
		{"all-members carrying a value is rejected", requests.AccessPolicySubject{Type: "all-members", Value: memberID}, true},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
				Return(namespace, nil).Once()
			storeMock.On("APIKeyList", ctx, mock.Anything).
				Return([]models.APIKey{}, 0, nil).Maybe()

			if !tc.rejected {
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Maybe()
				storeMock.On("AccessPolicyCreate", ctx, mock.Anything).Return("policy-1", nil).Once()
				storeMock.On("AccessPolicyResolve", ctx, mock.Anything, store.AccessPolicyIDResolver, "policy-1").
					Return(&models.AccessPolicy{ID: "policy-1"}, nil).Once()
			}

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

			_, err := service.CreateAccessPolicy(ctx, &requests.AccessPolicyCreate{
				TenantID: tenantID,
				Name:     "rule",
				Subject:  tc.subject,
				Logins:   []string{"root"},
			})

			if tc.rejected {
				require.ErrorIs(t, err, ErrAccessPolicyInvalidField)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUpdateAccessPolicyValidatesTheSubject(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID   = "00000000-0000-4000-0000-000000000000"
		memberID   = "11111111-1111-4111-1111-111111111111"
		strangerID = "33333333-3333-4333-3333-333333333333"
		policyID   = "policy-1"
	)

	namespace := &models.Namespace{
		TenantID: tenantID,
		Members:  []models.Member{{ID: memberID, Role: authorizer.RoleOperator, Type: models.UserTypeHuman}},
	}

	storeMock := storemock.NewMockStore(t)
	storeMock.On("AccessPolicyResolve", ctx, mock.Anything, store.AccessPolicyIDResolver, policyID).
		Return(&models.AccessPolicy{ID: policyID}, nil).Once()
	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
		Return(namespace, nil).Once()
	storeMock.On("APIKeyList", ctx, mock.Anything).Return([]models.APIKey{}, 0, nil).Maybe()

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	_, err := service.UpdateAccessPolicy(ctx, &requests.AccessPolicyUpdate{
		AccessPolicyIDParam: requests.AccessPolicyIDParam{ID: policyID},
		TenantID:            tenantID,
		Name:                "rule",
		Subject:             requests.AccessPolicySubject{Type: "user", Value: strangerID},
		Logins:              []string{"root"},
	})

	require.ErrorIs(t, err, ErrAccessPolicyInvalidField)
}

func TestListAccessPoliciesReportsASubjectThatMatchesNobody(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID   = "00000000-0000-4000-0000-000000000000"
		memberID   = "11111111-1111-4111-1111-111111111111"
		departedID = "44444444-4444-4444-4444-444444444444"
	)

	namespace := &models.Namespace{
		TenantID: tenantID,
		Members:  []models.Member{{ID: memberID, Role: authorizer.RoleOperator, Type: models.UserTypeHuman}},
	}

	stored := []models.AccessPolicy{
		{ID: "a", Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: memberID}},
		{ID: "b", Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: departedID}},
		{ID: "c", Subject: models.PolicySubject{Type: models.PolicySubjectRole, Value: "operator"}},
		{ID: "d", Subject: models.PolicySubject{Type: models.PolicySubjectRole, Value: "owner"}},
		{ID: "e", Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers}},
	}

	storeMock := storemock.NewMockStore(t)
	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
		Return(namespace, nil).Once()
	storeMock.On("APIKeyList", ctx, mock.Anything).Return([]models.APIKey{}, 0, nil).Maybe()
	storeMock.On("AccessPolicyList", ctx, mock.Anything).Return(stored, len(stored), nil).Once()

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	policies, err := service.ListAccessPolicies(ctx, tenantID)
	require.NoError(t, err)

	matches := make(map[string]bool, len(policies))
	for _, policy := range policies {
		matches[policy.ID] = policy.SubjectMatches
	}

	require.Equal(t, map[string]bool{
		"a": true,
		"b": false,
		"c": true,
		"d": false,
		"e": true,
	}, matches)
}

func TestAccessPolicyReadPathsReportASubjectThatMatchesNobody(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		memberID = "11111111-1111-4111-1111-111111111111"
		policyID = "policy-1"
	)

	namespace := &models.Namespace{
		TenantID: tenantID,
		Members:  []models.Member{{ID: memberID, Role: authorizer.RoleOperator, Type: models.UserTypeHuman}},
	}

	reads := []struct {
		description string
		read        func(*APIService, *storemock.MockStore, string) (*models.AccessPolicy, error)
	}{
		{
			description: "get",
			read: func(service *APIService, storeMock *storemock.MockStore, role string) (*models.AccessPolicy, error) {
				storeMock.On("AccessPolicyResolve", ctx, mock.Anything, store.AccessPolicyIDResolver, policyID).
					Return(storedAccessPolicyWithRole(tenantID, policyID, role), nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).Once()
				storeMock.On("APIKeyList", ctx, mock.Anything).Return([]models.APIKey{}, 0, nil).Maybe()

				return service.GetAccessPolicy(ctx, &requests.AccessPolicyGet{
					AccessPolicyIDParam: requests.AccessPolicyIDParam{ID: policyID},
					TenantID:            tenantID,
				})
			},
		},
		{
			description: "create",
			read: func(service *APIService, storeMock *storemock.MockStore, role string) (*models.AccessPolicy, error) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).Once()
				storeMock.On("APIKeyList", ctx, mock.Anything).Return([]models.APIKey{}, 0, nil).Maybe()
				storeMock.On("Options").Return(new(storemock.MockQueryOptions)).Maybe()
				storeMock.On("AccessPolicyCreate", ctx, mock.Anything).Return(policyID, nil).Once()
				storeMock.On("AccessPolicyResolve", ctx, mock.Anything, store.AccessPolicyIDResolver, policyID).
					Return(storedAccessPolicyWithRole(tenantID, policyID, role), nil).Once()

				return service.CreateAccessPolicy(ctx, &requests.AccessPolicyCreate{
					TenantID: tenantID,
					Name:     "rule",
					Subject:  requests.AccessPolicySubject{Type: "role", Value: role},
					Logins:   []string{"root"},
				})
			},
		},
		{
			description: "update",
			read: func(service *APIService, storeMock *storemock.MockStore, role string) (*models.AccessPolicy, error) {
				storeMock.On("AccessPolicyResolve", ctx, mock.Anything, store.AccessPolicyIDResolver, policyID).
					Return(storedAccessPolicyWithRole(tenantID, policyID, role), nil).Twice()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespace, nil).Once()
				storeMock.On("APIKeyList", ctx, mock.Anything).Return([]models.APIKey{}, 0, nil).Maybe()
				storeMock.On("Options").Return(new(storemock.MockQueryOptions)).Maybe()
				storeMock.On("AccessPolicyUpdate", ctx, mock.Anything).Return(nil).Once()

				return service.UpdateAccessPolicy(ctx, &requests.AccessPolicyUpdate{
					AccessPolicyIDParam: requests.AccessPolicyIDParam{ID: policyID},
					TenantID:            tenantID,
					Name:                "rule",
					Subject:             requests.AccessPolicySubject{Type: "role", Value: role},
					Logins:              []string{"root"},
				})
			},
		},
	}

	subjects := []struct {
		description string
		role        string
		matches     bool
	}{
		{"a role a member holds matches", "operator", true},
		{"a role no member holds matches nobody", "administrator", false},
	}

	for _, read := range reads {
		for _, subject := range subjects {
			t.Run(read.description+": "+subject.description, func(t *testing.T) {
				storeMock := storemock.NewMockStore(t)
				service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

				policy, err := read.read(service, storeMock, subject.role)
				require.NoError(t, err)
				require.Equal(t, subject.matches, policy.SubjectMatches)
			})
		}
	}
}

func storedAccessPolicyWithRole(tenantID, id, role string) *models.AccessPolicy {
	return &models.AccessPolicy{
		ID:       id,
		TenantID: tenantID,
		Subject:  models.PolicySubject{Type: models.PolicySubjectRole, Value: role},
	}
}
