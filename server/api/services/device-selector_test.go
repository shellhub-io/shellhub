package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type deviceSelectorSourceFunc struct {
	permission authorizer.Permission
	list       func(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error)
}

func (f deviceSelectorSourceFunc) EditPermission() authorizer.Permission {
	return f.permission
}

func (f deviceSelectorSourceFunc) DeviceSelectors(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error) {
	return f.list(ctx, sc)
}

func firewallSelectors(selectors ...DeviceSelector) deviceSelectorSourceFunc {
	return deviceSelectorSourceFunc{
		permission: authorizer.FirewallEdit,
		list: func(context.Context, scope.Scope) ([]DeviceSelector, error) {
			return selectors, nil
		},
	}
}

func asRole(role authorizer.Role) context.Context {
	return identity.NewContext(context.Background(), identity.Identity{ID: "member", TenantID: "tenant1", Role: role})
}

const (
	selectorTagID     = "tag_00000000-0000-4000-0000-000000000000"
	selectorDeviceUID = "device_00000000-0000-4000-0000-000000000000"
)

func TestDeviceTagLinkFollowsTheRulesSelectingByTheTag(t *testing.T) {
	tag := &models.Tag{ID: selectorTagID, Name: "production"}
	byTag := models.Taggable{TagIDs: []string{selectorTagID}}

	cases := []struct {
		description string
		role        authorizer.Role
		policies    []models.AccessPolicy
		publicKeys  []models.PublicKey
		extra       []DeviceSelector
		linked      bool
	}{
		{
			description: "an operator moves a tag no rule selects by",
			role:        authorizer.RoleOperator,
			linked:      true,
		},
		{
			description: "an operator does not move a tag an access policy selects by",
			role:        authorizer.RoleOperator,
			policies:    []models.AccessPolicy{{Name: "production", Filter: models.AccessPolicyFilter{Taggable: byTag}}},
		},
		{
			description: "an operator does not move a tag a public key selects by",
			role:        authorizer.RoleOperator,
			publicKeys:  []models.PublicKey{{PublicKeyFields: models.PublicKeyFields{Name: "deploy", Filter: models.PublicKeyFilter{Taggable: byTag}}}},
		},
		{
			description: "an operator does not move a tag a registered rule selects by",
			role:        authorizer.RoleOperator,
			extra:       []DeviceSelector{{Name: `firewall rule "rule1"`, TagIDs: []string{selectorTagID}}},
		},
		{
			description: "an administrator moves a tag every rule it may edit selects by",
			role:        authorizer.RoleAdministrator,
			policies:    []models.AccessPolicy{{Name: "production", Filter: models.AccessPolicyFilter{Taggable: byTag}}},
			publicKeys:  []models.PublicKey{{PublicKeyFields: models.PublicKeyFields{Name: "deploy", Filter: models.PublicKeyFilter{Taggable: byTag}}}},
			extra:       []DeviceSelector{{Name: `firewall rule "rule1"`, TagIDs: []string{selectorTagID}}},
			linked:      true,
		},
	}

	for _, tc := range cases {
		for _, attach := range []bool{true, false} {
			name := "detach: " + tc.description
			if attach {
				name = "attach: " + tc.description
			}

			t.Run(name, func(t *testing.T) {
				ctx := asRole(tc.role)

				storeMock := storemock.NewMockStore(t)
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").Return(&models.Namespace{}, nil).Once()
				storeMock.On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").Return(tag, nil).Once()
				storeMock.On("AccessPolicyList", ctx, mock.Anything).Return(tc.policies, int64(len(tc.policies)), nil).Maybe()
				storeMock.On("PublicKeyList", ctx, mock.Anything).Return(tc.publicKeys, int64(len(tc.publicKeys)), nil).Maybe()

				if tc.linked && attach {
					storeMock.On("TagPushToTarget", ctx, selectorTagID, store.TagTargetDevice, selectorDeviceUID).Return(nil).Once()
				}

				if tc.linked && !attach {
					storeMock.On("TagPullFromTarget", ctx, selectorTagID, store.TagTargetDevice, []string{selectorDeviceUID}).Return(nil).Once()
				}

				service := NewService(storeMock, privateKey, publicKey, nil, WithDeviceSelectorSource(firewallSelectors(tc.extra...)))

				var err error
				if attach {
					err = service.PushTagTo(ctx, store.TagTargetDevice, &requests.PushTag{TenantID: "tenant1", Name: "production", TargetID: selectorDeviceUID})
				} else {
					err = service.PullTagFrom(ctx, store.TagTargetDevice, &requests.PullTag{TenantID: "tenant1", Name: "production", TargetID: selectorDeviceUID})
				}

				if tc.linked {
					require.NoError(t, err)

					return
				}

				require.ErrorIs(t, err, ErrDeviceSelectorOutOfReach)
				storeMock.AssertNotCalled(t, "TagPushToTarget", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				storeMock.AssertNotCalled(t, "TagPullFromTarget", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		}
	}
}

func TestDeviceRenameFollowsTheRulesSelectingByName(t *testing.T) {
	webKey := models.PublicKey{PublicKeyFields: models.PublicKeyFields{Name: "web", Filter: models.PublicKeyFilter{Hostname: "web-.*"}}}

	cases := []struct {
		description string
		role        authorizer.Role
		from, to    string
		publicKeys  []models.PublicKey
		extra       []DeviceSelector
		renamed     bool
	}{
		{
			description: "an operator renames a device no name pattern tells apart from its new name",
			role:        authorizer.RoleOperator,
			from:        "db-1",
			to:          "db-2",
			publicKeys:  []models.PublicKey{webKey},
			renamed:     true,
		},
		{
			description: "an operator does not rename a device out of a public key's name pattern",
			role:        authorizer.RoleOperator,
			from:        "web-1",
			to:          "db-1",
			publicKeys:  []models.PublicKey{webKey},
		},
		{
			description: "an operator does not rename a device into a registered rule's name pattern",
			role:        authorizer.RoleOperator,
			from:        "db-1",
			to:          "prod-db-1",
			extra:       []DeviceSelector{{Name: `firewall rule "rule1"`, Hostname: "prod-.*"}},
		},
		{
			description: "an administrator renames a device across every pattern it may edit",
			role:        authorizer.RoleAdministrator,
			from:        "web-1",
			to:          "prod-db-1",
			publicKeys:  []models.PublicKey{webKey},
			extra:       []DeviceSelector{{Name: `firewall rule "rule1"`, Hostname: "prod-.*"}},
			renamed:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := asRole(tc.role)
			device := &models.Device{UID: selectorDeviceUID, Name: tc.from, TenantID: "tenant1"}

			storeMock := storemock.NewMockStore(t)
			storeMock.On("DeviceResolve", ctx, mock.Anything, store.DeviceUIDResolver, selectorDeviceUID).Return(device, nil).Once()
			storeMock.On("AccessPolicyList", ctx, mock.Anything).Return([]models.AccessPolicy{}, int64(0), nil).Maybe()
			storeMock.On("PublicKeyList", ctx, mock.Anything).Return(tc.publicKeys, int64(len(tc.publicKeys)), nil).Maybe()

			if tc.renamed {
				storeMock.On("DeviceConflicts", ctx, mock.Anything, mock.Anything).Return(nil, false, nil).Once()
				storeMock.On("DeviceUpdate", ctx, mock.Anything, mock.Anything).Return(nil).Once()
			}

			service := NewService(storeMock, privateKey, publicKey, nil, WithDeviceSelectorSource(firewallSelectors(tc.extra...)))

			err := service.RenameDevice(ctx, models.UID(selectorDeviceUID), tc.to, "tenant1")

			if tc.renamed {
				require.NoError(t, err)
				assert.Equal(t, tc.to, device.Name)

				return
			}

			require.ErrorIs(t, err, ErrDeviceSelectorOutOfReach)
			storeMock.AssertNotCalled(t, "DeviceUpdate", mock.Anything, mock.Anything)
		})
	}
}

func TestDeviceSelectorSourceNamesAPermissionThatSetsItApart(t *testing.T) {
	cases := []struct {
		description string
		permission  authorizer.Permission
		accepted    bool
	}{
		{"a permission only administrators hold", authorizer.FirewallEdit, true},
		{"the permission that moves tags", authorizer.TagCreate, false},
		{"the permission that renames devices", authorizer.DeviceRename, false},
		{"a permission every role that moves tags holds", authorizer.DeviceStatusUpdate, false},
		{"a permission no role holds", authorizer.Permission(-1), false},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			source := deviceSelectorSourceFunc{permission: tc.permission}
			build := func() { NewService(nil, privateKey, publicKey, nil, WithDeviceSelectorSource(source)) }

			if tc.accepted {
				assert.NotPanics(t, build)
			} else {
				assert.Panics(t, build)
			}
		})
	}
}
