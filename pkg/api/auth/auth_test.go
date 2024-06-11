package auth_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/auth"
	"github.com/stretchr/testify/require"
)

func TestRoleFromString(t *testing.T) {
	cases := []struct {
		description string
		role        string
		expected    auth.Role
	}{
		{
			description: "fails with invalid roles",
			role:        "N/A",
			expected:    auth.RoleInvalid,
		},
		{
			description: "succeeds with 'owner'",
			role:        "owner",
			expected:    auth.RoleOwner,
		},
		{
			description: "succeeds with 'administrator'",
			role:        "administrator",
			expected:    auth.RoleAdministrator,
		},
		{
			description: "succeeds with 'operator'",
			role:        "operator",
			expected:    auth.RoleOperator,
		},
		{
			description: "succeeds with 'observer'",
			role:        "observer",
			expected:    auth.RoleObserver,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expected, auth.RoleFromString(tc.role))
		})
	}
}

func TestRolePermissions(t *testing.T) {
	cases := []struct {
		description string
		role        auth.Role
		expected    []auth.Permission
	}{
		{
			description: "fails with invalid roles",
			role:        auth.RoleInvalid,
			expected:    []auth.Permission{},
		},
		{
			description: "succeeds with 'owner'",
			role:        auth.RoleOwner,
			expected: []auth.Permission{
				auth.DeviceAccept,
				auth.DeviceReject,
				auth.DeviceRemove,
				auth.DeviceConnect,
				auth.DeviceRename,
				auth.DeviceDetails,
				auth.DeviceUpdate,
				auth.DeviceCreateTag,
				auth.DeviceUpdateTag,
				auth.DeviceRemoveTag,
				auth.DeviceRenameTag,
				auth.DeviceDeleteTag,
				auth.SessionPlay,
				auth.SessionClose,
				auth.SessionRemove,
				auth.SessionDetails,
				auth.FirewallCreate,
				auth.FirewallEdit,
				auth.FirewallRemove,
				auth.FirewallAddTag,
				auth.FirewallRemoveTag,
				auth.FirewallUpdateTag,
				auth.PublicKeyCreate,
				auth.PublicKeyEdit,
				auth.PublicKeyRemove,
				auth.PublicKeyAddTag,
				auth.PublicKeyRemoveTag,
				auth.PublicKeyUpdateTag,
				auth.NamespaceUpdate,
				auth.NamespaceAddMember,
				auth.NamespaceRemoveMember,
				auth.NamespaceEditMember,
				auth.NamespaceEnableSessionRecord,
				auth.NamespaceDelete,
				auth.BillingCreateCustomer,
				auth.BillingChooseDevices,
				auth.BillingAddPaymentMethod,
				auth.BillingUpdatePaymentMethod,
				auth.BillingRemovePaymentMethod,
				auth.BillingCancelSubscription,
				auth.BillingCreateSubscription,
				auth.BillingGetSubscription,
				auth.APIKeyCreate,
				auth.APIKeyUpdate,
				auth.APIKeyDelete,
			},
		},
		{
			description: "succeeds with 'administrator'",
			role:        auth.RoleAdministrator,
			expected: []auth.Permission{
				auth.DeviceAccept,
				auth.DeviceReject,
				auth.DeviceRemove,
				auth.DeviceConnect,
				auth.DeviceRename,
				auth.DeviceDetails,
				auth.DeviceUpdate,
				auth.DeviceCreateTag,
				auth.DeviceUpdateTag,
				auth.DeviceRemoveTag,
				auth.DeviceRenameTag,
				auth.DeviceDeleteTag,
				auth.SessionPlay,
				auth.SessionClose,
				auth.SessionRemove,
				auth.SessionDetails,
				auth.FirewallCreate,
				auth.FirewallEdit,
				auth.FirewallRemove,
				auth.FirewallAddTag,
				auth.FirewallRemoveTag,
				auth.FirewallUpdateTag,
				auth.PublicKeyCreate,
				auth.PublicKeyEdit,
				auth.PublicKeyRemove,
				auth.PublicKeyAddTag,
				auth.PublicKeyRemoveTag,
				auth.PublicKeyUpdateTag,
				auth.NamespaceUpdate,
				auth.NamespaceAddMember,
				auth.NamespaceRemoveMember,
				auth.NamespaceEditMember,
				auth.NamespaceEnableSessionRecord,
				auth.APIKeyCreate,
				auth.APIKeyUpdate,
				auth.APIKeyDelete,
			},
		},
		{
			description: "succeeds with 'operator'",
			role:        auth.RoleOperator,
			expected: []auth.Permission{
				auth.DeviceAccept,
				auth.DeviceReject,
				auth.DeviceConnect,
				auth.DeviceRename,
				auth.DeviceDetails,
				auth.DeviceUpdate,
				auth.DeviceCreateTag,
				auth.DeviceUpdateTag,
				auth.DeviceRemoveTag,
				auth.DeviceRenameTag,
				auth.DeviceDeleteTag,
				auth.SessionDetails,
			},
		},
		{
			description: "succeeds with 'observer'",
			role:        auth.RoleObserver,
			expected: []auth.Permission{
				auth.DeviceConnect,
				auth.DeviceDetails,
				auth.SessionDetails,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expected, tc.role.Permissions())
		})
	}
}

func TestRolePreferences(t *testing.T) {
	cases := []struct {
		description string
		role        auth.Role
		greater     []auth.Role
		less        []auth.Role
	}{
		{
			description: auth.RoleInvalid.String(),
			role:        auth.RoleInvalid,
			greater:     []auth.Role{auth.RoleOwner, auth.RoleAdministrator, auth.RoleOperator, auth.RoleObserver},
			less:        []auth.Role{},
		},
		{
			description: auth.RoleOwner.String(),
			role:        auth.RoleOwner,
			greater:     []auth.Role{},
			less:        []auth.Role{auth.RoleAdministrator, auth.RoleOperator, auth.RoleObserver},
		},
		{
			description: auth.RoleAdministrator.String(),
			role:        auth.RoleAdministrator,
			greater:     []auth.Role{auth.RoleOwner},
			less:        []auth.Role{auth.RoleOperator, auth.RoleObserver},
		},
		{
			description: auth.RoleOperator.String(),
			role:        auth.RoleOperator,
			greater:     []auth.Role{auth.RoleOwner, auth.RoleAdministrator},
			less:        []auth.Role{auth.RoleObserver},
		},
		{
			description: auth.RoleObserver.String(),
			role:        auth.RoleObserver,
			greater:     []auth.Role{auth.RoleOwner, auth.RoleAdministrator, auth.RoleOperator},
			less:        []auth.Role{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			for _, r := range tc.greater {
				require.Equal(tt, false, tc.role.GreaterThan(r))
			}

			for _, r := range tc.less {
				require.Equal(tt, true, tc.role.GreaterThan(r))
			}
		})
	}
}
