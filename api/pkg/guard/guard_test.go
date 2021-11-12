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
				Type: authorizer.MemberTypeOwner,
			},
			{
				ID:   userObserver.ID,
				Type: authorizer.MemberTypeObserver,
			},
			{
				ID:   userOperator.ID,
				Type: authorizer.MemberTypeOperator,
			},
			{
				ID:   userAdministrator.ID,
				Type: authorizer.MemberTypeAdministrator,
			},
		},
	}
	passiveTypeOperator := authorizer.MemberTypeOperator
	passiveTypeObserver := authorizer.MemberTypeObserver
	passiveTypeAdministrator := authorizer.MemberTypeAdministrator
	passiveTypeOwner := authorizer.MemberTypeOwner

	cases := []struct {
		description   string
		tenantID      string
		activeID      string
		passiveType   string
		requiredMocks func()
		expected      bool
	}{
		{
			description: "EvaluateSubject successes when active user is a operator and passive type is observer",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveType: passiveTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a operator and passive type is operator",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveType: passiveTypeOperator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a operator and passive type is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userOperator.ID,
			passiveType: passiveTypeAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject successes when active user is a administrator and passive type is observer",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveType: passiveTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject success when active user is a administrator and passive type is operator",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveType: passiveTypeOperator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a administrator and passive type is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveType: passiveTypeAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a administrator and passive type is owner",
			tenantID:    namespace.TenantID,
			activeID:    userAdministrator.ID,
			passiveType: passiveTypeOwner,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: false,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive type is observer",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveType: passiveTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive type is operator",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveType: passiveTypeObserver,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive type is administrator",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveType: passiveTypeAdministrator,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluateSubject fails when active user is a owner and passive type is owner",
			tenantID:    namespace.TenantID,
			activeID:    userOwner.ID,
			passiveType: passiveTypeOwner,
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
			ok := EvaluateSubject(ctx, store.Store(mock), tc.tenantID, tc.activeID, tc.passiveType)
			assert.Equal(t, tc.expected, ok)
		})
	}

	mock.AssertExpectations(t)
}

func TestEvaluatePermission(t *testing.T) {
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
				Type: authorizer.MemberTypeOwner,
			},
			{
				ID:   userObserver.ID,
				Type: authorizer.MemberTypeObserver,
			},
			{
				ID:   userOperator.ID,
				Type: authorizer.MemberTypeOperator,
			},
			{
				ID:   userAdministrator.ID,
				Type: authorizer.MemberTypeAdministrator,
			},
		},
	}

	cases := []struct {
		description   string
		tenantID      string
		userID        string
		actions       []int
		requiredMocks func()
		expected      bool
	}{
		{
			description: "EvaluatePermission success when user is the observer",
			tenantID:    namespace.TenantID,
			userID:      userObserver.ID,
			actions: []int{
				authorizer.Actions.Device.Connect,

				authorizer.Actions.Session.Details,

				authorizer.Actions.Namespace.Create,
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userObserver.ID, false).Return(userObserver, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluatePermission success when user is the operator",
			tenantID:    namespace.TenantID,
			userID:      userOperator.ID,
			actions: []int{
				authorizer.Actions.Device.Accept,
				authorizer.Actions.Device.Reject,
				authorizer.Actions.Device.Connect,
				authorizer.Actions.Device.Rename,

				authorizer.Actions.Session.Details,

				authorizer.Actions.Namespace.Create,
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOperator.ID, false).Return(userOperator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluatePermission success when user is the administrator",
			tenantID:    namespace.TenantID,
			userID:      userAdministrator.ID,
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

				authorizer.Actions.Namespace.Create,
				authorizer.Actions.Namespace.Rename,
				authorizer.Actions.Namespace.AddMember,
				authorizer.Actions.Namespace.RemoveMember,
				authorizer.Actions.Namespace.EditMember,
				authorizer.Actions.Namespace.EnableSessionRecord,
				authorizer.Actions.Namespace.Delete,
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userAdministrator.ID, false).Return(userAdministrator, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
		{
			description: "EvaluatePermission success when user is the owner",
			tenantID:    namespace.TenantID,
			userID:      userOwner.ID,
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

				authorizer.Actions.Namespace.Create,
				authorizer.Actions.Namespace.Rename,
				authorizer.Actions.Namespace.AddMember,
				authorizer.Actions.Namespace.RemoveMember,
				authorizer.Actions.Namespace.EditMember,
				authorizer.Actions.Namespace.EnableSessionRecord,
				authorizer.Actions.Namespace.Delete,
			},
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, userOwner.ID, false).Return(userOwner, 0, nil)
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil)
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			for _, action := range tc.actions {
				assert.True(t, EvaluatePermission(ctx, store.Store(mock), tc.tenantID, tc.userID, action))
			}
		})
	}

	mock.AssertExpectations(t)
}
