package guard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func TestEvaluateSubject(t *testing.T) {
	mock := &mocks.Store{}

	ctx := context.TODO()

	userOwner := &models.User{
		ID: "userOwnerID",
		UserData: models.UserData{
			Name:     "userOwner",
			Email:    "userOwner@userOwner.com",
			Username: "usernameOwner",
		},
	}

	userObserver := &models.User{
		ID: "userObserverID",
		UserData: models.UserData{
			Name:     "userObserver",
			Email:    "userObserver@userObserver.com",
			Username: "usernameObserver",
		},
	}
	userOperator := &models.User{
		ID: "userOperatorID",
		UserData: models.UserData{
			Name:     "userOperator",
			Email:    "userOperator@userOperator.com",
			Username: "usernameOperator",
		},
	}

	userAdministrator := &models.User{
		ID: "userAdministratorID",
		UserData: models.UserData{
			Name:     "userAdministrator",
			Email:    "userAdministrator@userAdministrator.com",
			Username: "usernameAdministrator",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    userOwner.ID,
		TenantID: "tenantID",
		Members: []models.Member{
			{
				ID:   userOwner.ID,
				Role: authorizer.MemberRoleOwner,
			},
			{
				ID:   userObserver.ID,
				Role: authorizer.MemberRoleObserver,
			},
			{
				ID:   userOperator.ID,
				Role: authorizer.MemberRoleOperator,
			},
			{
				ID:   userAdministrator.ID,
				Role: authorizer.MemberRoleAdministrator,
			},
		},
	}
	passiveRoleOperator := authorizer.MemberRoleOperator
	passiveRoleObserver := authorizer.MemberRoleObserver
	passiveRoleAdministrator := authorizer.MemberRoleAdministrator
	passiveRoleOwner := authorizer.MemberRoleOwner

	cases := []struct {
		description   string
		tenantID      string
		activeID      string
		passiveRole   string
		requiredMocks func()
		expected      bool
	}{
		{
			description: "EvaluateSubject successes when active user is a operator and passive role is observer",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveRole: passiveRoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a operator and passive role is operator",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveRole: passiveRoleOperator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a operator and passive role is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveRole: passiveRoleAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject successes when active user is a administrator and passive role is observer",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveRole: passiveRoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject success when active user is a administrator and passive role is operator",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveRole: passiveRoleOperator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a administrator and passive role is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveRole: passiveRoleAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a administrator and passive role is owner",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveRole: passiveRoleOwner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive role is observer",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveRole: passiveRoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive role is operator",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveRole: passiveRoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive role is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveRole: passiveRoleAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive role is owner",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveRole: passiveRoleOwner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			ok := EvaluateSubject(ctx, store.Store(mock), tc.tenantID, tc.activeID, tc.passiveRole)
			assert.Equal(t, tc.expected, ok)
		})
	}

	mock.AssertExpectations(t)
}

func TestEvaluatePermission(t *testing.T) {
	mock := &mocks.Store{}

	cases := []struct {
		description   string
		role          string
		actions       []int
		requiredMocks func()
		expected      bool
	}{
		{
			description: "CheckPermission success when user is the observer",
			role:        authorizer.MemberRoleObserver,
			actions: []int{
				authorizer.Actions.Device.Connect,

				authorizer.Actions.Session.Details,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the operator",
			role:        authorizer.MemberRoleOperator,
			actions: []int{
				authorizer.Actions.Device.Accept,
				authorizer.Actions.Device.Reject,
				authorizer.Actions.Device.Connect,
				authorizer.Actions.Device.Rename,

				authorizer.Actions.Session.Details,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the administrator",
			role:        authorizer.MemberRoleAdministrator,
			actions: []int{
				authorizer.Actions.Device.Accept,
				authorizer.Actions.Device.Reject,
				authorizer.Actions.Device.Remove,
				authorizer.Actions.Device.Connect,
				authorizer.Actions.Device.Rename,

				authorizer.Actions.Session.Play,
				authorizer.Actions.Session.Close,
				authorizer.Actions.Session.Remove,
				authorizer.Actions.Session.Details,

				authorizer.Actions.Firewall.Create,
				authorizer.Actions.Firewall.Edit,
				authorizer.Actions.Firewall.Remove,

				authorizer.Actions.PublicKey.Create,
				authorizer.Actions.PublicKey.Edit,
				authorizer.Actions.PublicKey.Remove,

				authorizer.Actions.Namespace.Rename,
				authorizer.Actions.Namespace.AddMember,
				authorizer.Actions.Namespace.RemoveMember,
				authorizer.Actions.Namespace.EditMember,
				authorizer.Actions.Namespace.EnableSessionRecord,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the owner",
			role:        authorizer.MemberRoleOwner,
			actions: []int{
				authorizer.Actions.Device.Accept,
				authorizer.Actions.Device.Reject,
				authorizer.Actions.Device.Remove,
				authorizer.Actions.Device.Connect,
				authorizer.Actions.Device.Rename,

				authorizer.Actions.Session.Play,
				authorizer.Actions.Session.Close,
				authorizer.Actions.Session.Remove,
				authorizer.Actions.Session.Details,

				authorizer.Actions.Firewall.Create,
				authorizer.Actions.Firewall.Edit,
				authorizer.Actions.Firewall.Remove,

				authorizer.Actions.PublicKey.Create,
				authorizer.Actions.PublicKey.Edit,
				authorizer.Actions.PublicKey.Remove,

				authorizer.Actions.Namespace.Rename,
				authorizer.Actions.Namespace.AddMember,
				authorizer.Actions.Namespace.RemoveMember,
				authorizer.Actions.Namespace.EditMember,
				authorizer.Actions.Namespace.EnableSessionRecord,
				authorizer.Actions.Namespace.Delete,

				authorizer.Actions.Billing.AddPaymentMethod,
				authorizer.Actions.Billing.UpdatePaymentMethod,
				authorizer.Actions.Billing.RemovePaymentMethod,
				authorizer.Actions.Billing.ChooseDevices,
				authorizer.Actions.Billing.CancelSubscription,
				authorizer.Actions.Billing.CreateSubscription,
				authorizer.Actions.Billing.GetSubscription,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			for _, action := range tc.actions {
				assert.NoError(t, EvaluatePermission(tc.role, action, func() error {
					return nil
				}))
			}
		})
	}

	mock.AssertExpectations(t)
}
