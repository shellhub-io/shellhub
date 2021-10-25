package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateType(t *testing.T) {
	cases := []struct {
		name string
		exec func(t *testing.T)
	}{
		{
			name: "Fail when the first type is not great than the second one",
			exec: func(t *testing.T) {
				assert.False(t, EvaluateType(MemberTypeAdministrator, MemberTypeOwner))
			},
		},
		{
			name: "Fail when a type is not valid",
			exec: func(t *testing.T) {
				assert.False(t, EvaluateType("invalidType", MemberTypeOperator))
			},
		},
		{
			name: "Fail when both types are equals",
			exec: func(t *testing.T) {
				assert.False(t, EvaluateType(MemberTypeOperator, MemberTypeOperator))
			},
		},
		{
			name: "Success when the first type is great than the second one",
			exec: func(t *testing.T) {
				assert.True(t, EvaluateType(MemberTypeAdministrator, MemberTypeOperator))
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
				action := Actions.Firewall.Create
				assert.False(t, checkPermission(action, observerPermissions))
			},
		},
		{
			name: "Success action is allowed",
			exec: func(t *testing.T) {
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
			name: "Fails when member's type has no permission",
			exec: func(t *testing.T) {
				ty := "observer"
				action := Actions.Firewall.Create
				assert.False(t, EvaluatePermission(action, ty))
			},
		},
		{
			name: "Success when member's type has permission",
			exec: func(t *testing.T) {
				ty := "owner"
				action := Actions.Firewall.Create
				assert.True(t, EvaluatePermission(action, ty))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.exec)
	}
}
