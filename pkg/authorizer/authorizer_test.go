package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateRole(t *testing.T) {
	cases := []struct {
		name string
		exec func(t *testing.T)
	}{
		{
			name: "Fail when the first role is not great than the second one",
			exec: func(t *testing.T) {
				t.Helper()
				assert.False(t, CheckRole(MemberRoleAdministrator, MemberRoleOwner))
			},
		},
		{
			name: "Fail when a role is not valid",
			exec: func(t *testing.T) {
				t.Helper()
				assert.False(t, CheckRole("invalidRole", MemberRoleOperator))
			},
		},
		{
			name: "Fail when both roles are equals",
			exec: func(t *testing.T) {
				t.Helper()
				assert.False(t, CheckRole(MemberRoleOperator, MemberRoleOperator))
			},
		},
		{
			name: "Success when the first role is great than the second one",
			exec: func(t *testing.T) {
				t.Helper()
				assert.True(t, CheckRole(MemberRoleAdministrator, MemberRoleOperator))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.exec)
	}
}

func TestCheckPermission(t *testing.T) {
	cases := []struct {
		name string
		exec func(t *testing.T)
	}{
		{
			name: "Fail when action is not allowed",
			exec: func(t *testing.T) {
				t.Helper()
				action := Actions.Firewall.Create
				assert.False(t, checkPermission(action, observerPermissions))
			},
		},
		{
			name: "Success action is allowed",
			exec: func(t *testing.T) {
				t.Helper()
				action := Actions.Device.Connect
				assert.True(t, checkPermission(action, observerPermissions))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.exec)
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
				ty := "observer"
				action := Actions.Firewall.Create
				assert.False(t, CheckPermission(ty, action))
			},
		},
		{
			name: "Success when member's role has permission",
			exec: func(t *testing.T) {
				t.Helper()
				ty := "owner"
				action := Actions.Firewall.Create
				assert.True(t, CheckPermission(ty, action))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.exec)
	}
}
