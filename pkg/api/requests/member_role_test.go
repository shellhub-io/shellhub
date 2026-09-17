package requests_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestsRefuseTheServiceRole(t *testing.T) {
	const id = "00000000-0000-4000-0000-000000000000"

	cases := map[string]func(authorizer.Role) any{
		"NamespaceAddMember": func(role authorizer.Role) any {
			return requests.NamespaceAddMember{
				ForwardedHost: "localhost", UserID: id, TenantID: id,
				MemberEmail: "member@test.com", MemberRole: role,
			}
		},
		"NamespaceUpdateMember": func(role authorizer.Role) any {
			return requests.NamespaceUpdateMember{
				UserID: id, TenantID: id, MemberID: id, MemberRole: role,
			}
		},
		"GenerateInvitationLink": func(role authorizer.Role) any {
			return requests.GenerateInvitationLink{
				ForwardedHost: "localhost", UserID: id, TenantID: id,
				MemberEmail: "member@test.com", MemberRole: role,
			}
		},
		"CreateAPIKey": func(role authorizer.Role) any {
			return requests.CreateAPIKey{
				UserID: id, TenantID: id, Role: authorizer.RoleOwner,
				Name: "key", ExpiresAt: 30, OptRole: role,
			}
		},
		"UpdateAPIKey": func(role authorizer.Role) any {
			return requests.UpdateAPIKey{
				UserID: id, TenantID: id, CurrentName: "key", Role: role,
			}
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			accepted, _ := validator.New().Struct(build(authorizer.RoleObserver))
			require.True(t, accepted, "the fixture must be valid apart from the role, or the refusal below proves nothing")

			refused, _ := validator.New().Struct(build(authorizer.RoleService))
			assert.False(t, refused)
		})
	}
}
