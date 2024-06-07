package guard

import (
	"errors"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasAuthority(t *testing.T) {
	cases := []struct {
		description string
		active      string
		passive     string
		expected    bool
	}{
		{
			description: "fails when the first role is not great than the second one",
			active:      RoleAdministrator,
			passive:     RoleOwner,
			expected:    false,
		},
		{
			description: "fails when a role is not valid",
			active:      "invalidRole",
			passive:     RoleOperator,
			expected:    false,
		},
		{
			description: "fails when passive role is owner",
			active:      RoleOwner,
			passive:     RoleOwner,
			expected:    false,
		},
		{
			description: "succeeds when both roles are equals",
			active:      RoleOperator,
			passive:     RoleOperator,
			expected:    true,
		},
		{
			description: "succeeds when the first role is great than the second one",
			active:      RoleAdministrator,
			passive:     RoleOperator,
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expected, HasAuthority(tc.active, tc.passive))
		})
	}
}

func TestEvaluatePermission(t *testing.T) {
	cases := []struct {
		name string
		exec func(t *testing.T)
	}{
		{
			name: "Fails when member's role has no permission",
			exec: func(t *testing.T) {
				t.Helper()

				role := RoleObserver
				action := Actions.Firewall.Create
				assert.Error(t, EvaluatePermission(role, action, nil))
			},
		},
		{
			name: "Success when member's role has permission",
			exec: func(t *testing.T) {
				t.Helper()

				role := RoleOwner
				action := Actions.Firewall.Create
				assert.NoError(t, EvaluatePermission(role, action, func() error {
					return nil
				}))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.exec)
	}
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
				Role: RoleOwner,
			},
			{
				ID:   userObserver.ID,
				Role: RoleObserver,
			},
			{
				ID:   userOperator.ID,
				Role: RoleOperator,
			},
			{
				ID:   userAdministrator.ID,
				Role: RoleAdministrator,
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

func TestCheckPermission(t *testing.T) {
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
			role:        RoleObserver,
			actions: []int{
				Actions.Device.Connect,

				Actions.Session.Details,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the operator",
			role:        RoleOperator,
			actions: []int{
				Actions.Device.Accept,
				Actions.Device.Reject,
				Actions.Device.Connect,
				Actions.Device.Rename,
				Actions.Device.Update,

				Actions.Device.CreateTag,
				Actions.Device.UpdateTag,
				Actions.Device.RemoveTag,
				Actions.Device.RenameTag,
				Actions.Device.DeleteTag,

				Actions.Session.Details,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the administrator",
			role:        RoleAdministrator,
			actions: []int{
				Actions.Device.Accept,
				Actions.Device.Reject,
				Actions.Device.Remove,
				Actions.Device.Connect,
				Actions.Device.Rename,
				Actions.Device.Update,

				Actions.Device.CreateTag,
				Actions.Device.UpdateTag,
				Actions.Device.RemoveTag,
				Actions.Device.RenameTag,
				Actions.Device.DeleteTag,

				Actions.Session.Play,
				Actions.Session.Close,
				Actions.Session.Remove,
				Actions.Session.Details,

				Actions.Firewall.Create,
				Actions.Firewall.Edit,
				Actions.Firewall.Remove,

				Actions.PublicKey.Create,
				Actions.PublicKey.Edit,
				Actions.PublicKey.Remove,

				Actions.Namespace.Update,
				Actions.Namespace.AddMember,
				Actions.Namespace.RemoveMember,
				Actions.Namespace.EditMember,
				Actions.Namespace.EnableSessionRecord,
			},
			requiredMocks: func() {
			},
			expected: true,
		},
		{
			description: "CheckPermission success when user is the owner",
			role:        RoleOwner,
			actions: []int{
				Actions.Device.Accept,
				Actions.Device.Reject,
				Actions.Device.Remove,
				Actions.Device.Connect,
				Actions.Device.Rename,
				Actions.Device.Update,

				Actions.Device.CreateTag,
				Actions.Device.UpdateTag,
				Actions.Device.RemoveTag,
				Actions.Device.RenameTag,
				Actions.Device.DeleteTag,

				Actions.Session.Play,
				Actions.Session.Close,
				Actions.Session.Remove,
				Actions.Session.Details,

				Actions.Firewall.Create,
				Actions.Firewall.Edit,
				Actions.Firewall.Remove,

				Actions.PublicKey.Create,
				Actions.PublicKey.Edit,
				Actions.PublicKey.Remove,

				Actions.Namespace.Update,
				Actions.Namespace.AddMember,
				Actions.Namespace.RemoveMember,
				Actions.Namespace.EditMember,
				Actions.Namespace.EnableSessionRecord,
				Actions.Namespace.Delete,

				Actions.Billing.AddPaymentMethod,
				Actions.Billing.UpdatePaymentMethod,
				Actions.Billing.RemovePaymentMethod,
				Actions.Billing.ChooseDevices,
				Actions.Billing.CancelSubscription,
				Actions.Billing.CreateSubscription,
				Actions.Billing.GetSubscription,
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

func ExampleHasAuthority_observer_and_observer() {
	// If members have the same role, they cannot act over each other.
	active := RoleObserver
	passive := RoleObserver
	fmt.Println(HasAuthority(active, passive))
	// Output: true
}

func ExampleHasAuthority_operator_and_observer() {
	// If active member has a great roles, it can act over passive one.
	active := RoleOperator
	passive := RoleObserver
	fmt.Println(HasAuthority(active, passive))
	// Output: true
}

func ExampleHasAuthority_owner_and_observer() {
	// If active member is owner, it can act over everyone.
	active := RoleOwner
	passive := RoleObserver
	fmt.Println(HasAuthority(active, passive))
	// Output: true
}

func ExampleEvaluatePermission_callback() {
	// RoleObserver can connect to device.
	err := EvaluatePermission(RoleObserver, Actions.Device.Connect, func() error {
		return errors.New("something went wrong")
	})
	fmt.Println(err)
	// Output: something went wrong
}

func ExampleEvaluatePermission_no_callback() {
	// RoleObserver cannot accept a device, so Forbidden is returned from EvaluatePermission.
	err := EvaluatePermission(RoleObserver, Actions.Device.Accept, func() error {
		// As RoleObserver has no permission to accept a device, this function will never be called.
		return errors.New("something went wrong")
	})
	fmt.Println(err)
	// Output: access forbidden
}

func ExampleGetRoleCode() {
	fmt.Println(GetRoleCode(RoleObserver))
	fmt.Println(GetRoleCode(RoleOperator))
	fmt.Println(GetRoleCode(RoleAdministrator))
	fmt.Println(GetRoleCode(""))
	fmt.Println(GetRoleCode("developer"))
	// Output:
	// 1
	// 2
	// 3
	// -1
	// -1
}
