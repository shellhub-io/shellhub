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
		sourceIP        string
		requireMocks    func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedAllowed bool
		expectedReauth  bool
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
			description: "grants and flags re-auth when the matched policy requires it",
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
							RequireReauth: true,
						},
					}, 1, nil).Once()
			},
			expectedAllowed: true,
			expectedReauth:  true,
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
							Action:  models.PolicyActionAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{},
							Logins:  []string{"*"},
							Action:  models.PolicyActionDeny,
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
							Action:  models.PolicyActionAllow,
						},
						{
							Subject: models.PolicySubject{Type: models.PolicySubjectUser, Value: userID},
							Filter:  models.PublicKeyFilter{Hostname: "["},
							Logins:  []string{"*"},
							Action:  models.PolicyActionDeny,
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
							Action:  models.PolicyActionDeny,
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
			expectedErr:     false,
		},
		{
			description: "allow grants when the client IP is inside the source CIDR",
			login:       "root",
			sourceIP:    "10.1.2.3",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
			expectedErr:     false,
		},
		{
			description: "empty source IP matches any client IP",
			login:       "root",
			sourceIP:    "203.0.113.9",
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
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "deny does not fire when the client IP is outside the deny source CIDR",
			login:       "root",
			sourceIP:    "10.0.0.5",
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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
							Action:  models.PolicyActionAllow,
						},
						{
							Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
							Filter:   models.PublicKeyFilter{},
							Logins:   []string{"*"},
							SourceIP: []string{"10.0.0.0/8"},
							Action:   models.PolicyActionDeny,
						},
					}, 2, nil).Once()
			},
			expectedAllowed: false,
			expectedErr:     false,
		},
		{
			description: "allow with an invalid source CIDR is skipped and stays default-deny",
			login:       "root",
			sourceIP:    "10.0.0.1",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, deviceID).
					Return(device, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
					Return(namespaceWith(authorizer.RoleOwner), nil).Once()
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
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

			decision, err := service.Authorize(ctx, tenantID, userID, &models.Device{UID: deviceID}, tc.login, tc.sourceIP)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAllowed, decision.Allowed)
				require.Equal(t, tc.expectedReauth, decision.RequireReauth)
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
	const saID = "00000000-0000-0000-0000-00000000000a"

	cases := []struct {
		description string
		subject     models.PolicySubject
		userID      string
		role        authorizer.Role
		userType    models.UserType
		expected    bool
	}{
		{
			description: "all-members matches a human",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			userID:      "human-id",
			role:        authorizer.RoleObserver,
			userType:    models.UserTypeHuman,
			expected:    true,
		},
		{
			description: "all-members does NOT match a service account (footgun)",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			userID:      saID,
			role:        authorizer.RoleService,
			userType:    models.UserTypeService,
			expected:    false,
		},
		{
			description: "a human role subject does not match a service account",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "observer"},
			userID:      saID,
			role:        authorizer.RoleService,
			userType:    models.UserTypeService,
			expected:    false,
		},
		{
			description: "role=service matches a service account",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "service"},
			userID:      saID,
			role:        authorizer.RoleService,
			userType:    models.UserTypeService,
			expected:    true,
		},
		{
			description: "role=service does not match a human observer",
			subject:     models.PolicySubject{Type: models.PolicySubjectRole, Value: "service"},
			userID:      "human-id",
			role:        authorizer.RoleObserver,
			userType:    models.UserTypeHuman,
			expected:    false,
		},
		{
			description: "user subject matches a service account by id",
			subject:     models.PolicySubject{Type: models.PolicySubjectUser, Value: saID},
			userID:      saID,
			role:        authorizer.RoleService,
			userType:    models.UserTypeService,
			expected:    true,
		},
		{
			description: "empty type is treated as human for all-members",
			subject:     models.PolicySubject{Type: models.PolicySubjectAllMembers},
			userID:      "legacy-id",
			role:        authorizer.RoleObserver,
			userType:    "",
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.Equal(t, tc.expected, subjectMatches(tc.subject, tc.userID, tc.role, tc.userType))
		})
	}
}
