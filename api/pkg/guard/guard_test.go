package guard

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateSubject(t *testing.T) {
	mock := &mocks.Store{}

	memberOperator := models.Member{
		ID:       "memberOperatorID",
		Username: "memberOperatorUsername",
		Role:     authorizer.MemberRoleOperator,
	}

	memberAdministrator := models.Member{
		ID:       "memberAdministratorID",
		Username: "memberAdministratorUsername",
		Role:     authorizer.MemberRoleAdministrator,
	}

	memberOwner := models.Member{
		ID:       "memberOwnerID",
		Username: "memberOwnerUsername",
		Role:     authorizer.MemberRoleOwner,
	}

	passiveRoleOperator := authorizer.MemberRoleOperator
	passiveRoleObserver := authorizer.MemberRoleObserver
	passiveRoleAdministrator := authorizer.MemberRoleAdministrator
	passiveRoleOwner := authorizer.MemberRoleOwner

	cases := []struct {
		description  string
		memberActive models.Member
		rolePassive  string
		expected     bool
	}{
		{
			description:  "CheckRole successes when active user is a operator and passive role is observer",
			memberActive: memberOperator,
			rolePassive:  passiveRoleObserver,
			expected:     true,
		},
		{
			description:  "CheckRole fails when active user is a operator and passive role is operator",
			memberActive: memberOperator,
			rolePassive:  passiveRoleOperator,
			expected:     false,
		},
		{
			description:  "CheckRole fails when active user is a operator and passive role is administrator",
			memberActive: memberOperator,
			rolePassive:  passiveRoleAdministrator,
			expected:     false,
		},
		{
			description:  "CheckRole successes when active user is a administrator and passive role is observer",
			memberActive: memberAdministrator,
			rolePassive:  passiveRoleObserver,
			expected:     true,
		},
		{
			description:  "CheckRole success when active user is a administrator and passive role is operator",
			memberActive: memberAdministrator,
			rolePassive:  passiveRoleOperator,
			expected:     true,
		},
		{
			description:  "CheckRole fails when active user is a administrator and passive role is administrator",
			memberActive: memberAdministrator,
			rolePassive:  passiveRoleAdministrator,
			expected:     false,
		},
		{
			description:  "CheckRole fails when active user is a administrator and passive role is owner",
			memberActive: memberAdministrator,
			rolePassive:  passiveRoleOwner,
			expected:     false,
		},
		{
			description:  "CheckRole fails when active user is a owner and passive role is observer",
			memberActive: memberOwner,
			rolePassive:  passiveRoleObserver,
			expected:     true,
		},
		{
			description:  "CheckRole fails when active user is a owner and passive role is operator",
			memberActive: memberOwner,
			rolePassive:  passiveRoleObserver,
			expected:     true,
		},
		{
			description:  "CheckRole fails when active user is a owner and passive role is administrator",
			memberActive: memberOwner,
			rolePassive:  passiveRoleAdministrator,
			expected:     true,
		},
		{
			description:  "CheckRole fails when active user is a owner and passive role is owner",
			memberActive: memberOwner,
			rolePassive:  passiveRoleOwner,
			expected:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ok := CheckRole(tc.memberActive.Role, tc.rolePassive)
			assert.Equal(t, tc.expected, ok)
		})
	}

	mock.AssertExpectations(t)
}

func TestEvaluateNamespace(t *testing.T) {
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

	cases := []struct {
		description string
		id          string
		namespace   *models.Namespace
		expected    bool
	}{
		{
			description: "Fails when user is not inside the namespace",
			id:          "invalidUserID",
			namespace:   namespace,
			expected:    false,
		},
		{
			description: "Success find the user inside the namespace 1",
			id:          userObserver.ID,
			namespace:   namespace,
			expected:    true,
		},
		{
			description: "Success find the user inside the namespace 2",
			id:          userOperator.ID,
			namespace:   namespace,
			expected:    true,
		},
		{
			description: "Success find the user inside the namespace 3",
			id:          userAdministrator.ID,
			namespace:   namespace,
			expected:    true,
		},
		{
			description: "Success find the user inside the namespace 4",
			id:          userOwner.ID,
			namespace:   namespace,
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			_, ok := CheckMember(tc.namespace, tc.id)

			assert.Equal(t, tc.expected, ok)
		})
	}
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

				authorizer.Actions.Device.CreateTag,
				authorizer.Actions.Device.UpdateTag,
				authorizer.Actions.Device.RemoveTag,
				authorizer.Actions.Device.RenameTag,
				authorizer.Actions.Device.DeleteTag,

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

				authorizer.Actions.Device.CreateTag,
				authorizer.Actions.Device.UpdateTag,
				authorizer.Actions.Device.RemoveTag,
				authorizer.Actions.Device.RenameTag,
				authorizer.Actions.Device.DeleteTag,

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

				authorizer.Actions.Device.CreateTag,
				authorizer.Actions.Device.UpdateTag,
				authorizer.Actions.Device.RemoveTag,
				authorizer.Actions.Device.RenameTag,
				authorizer.Actions.Device.DeleteTag,

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
