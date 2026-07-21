package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
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
		requireMocks    func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedAllowed bool
		expectedStepUp  bool
		expectedErr     bool
	}{
		{
			description: "denies when the device cannot be resolved",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedAllowed: false,
			expectedErr:     true,
		},
		{
			description: "denies when the user is not a member of the namespace",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, _ *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(&models.Namespace{TenantID: tenantID}, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "fails closed when the policy store errors",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{}, 0, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "grants when an all-members policy grants the login",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleObserver), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			expectedErr:     false,
		},
		{
			description: "grants when the login is explicitly listed",
			login:       "deploy",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleObserver), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			expectedErr:     false,
		},
		{
			description: "grants when the role subject matches the user's role",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleAdministrator), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			expectedErr:     false,
		},
		{
			description: "grants when the tag filter selects the device",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			description: "grants and flags step-up when the matched policy requires it",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject:       models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:        models.PublicKeyFilter{},
							Logins:        []string{"*"},
							RequireStepUp: true,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedStepUp:  true,
			expectedErr:     false,
		},
		{
			description: "skips a policy with a broken hostname regexp and stays default-deny",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			expectedErr:     false,
		},
		{
			description: "deny wins over a matching allow",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Effect:  models.PolicyEffectDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "deny does not fire for a login outside its list; allow still grants",
			login:       "teste",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Effect:  models.PolicyEffectDeny,
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "deny with a broken hostname regexp fails closed",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{Hostname: "["},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "denies when only a matching deny policy exists",
			login:       "root",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"root"},
							Effect:  models.PolicyEffectDeny,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "a deny-all blocks even a specific allow for the same subject",
			login:       "teste",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Effect:  models.PolicyEffectDeny,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"teste"},
							Effect:  models.PolicyEffectAllow,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

			decision, err := service.Authorize(ctx, tenantID, userID, &models.Device{UID: deviceID}, tc.login)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAllowed, decision.Allowed)
				require.Equal(t, tc.expectedStepUp, decision.RequireStepUp)
			}

			storeMock.AssertExpectations(t)
		})
	}
}
