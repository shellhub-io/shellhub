package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestNamespaceCreateMembership(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		member      *models.Member
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenantID:    "nonexistent",
			member: &models.Member{
				ID:   "6509de884238881ac1b2b289",
				Role: authorizer.RoleObserver,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: store.ErrNoDocuments},
		},
		{
			description: "fails when member has already been added",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:   "6509e169ae6144b2f56bf288",
				Role: authorizer.RoleObserver,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: mongo.ErrNamespaceDuplicatedMember},
		},
		{
			description: "succeeds when tenant is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:   "6509de884238881ac1b2b289",
				Role: authorizer.RoleObserver,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceCreateMembership(ctx, tc.tenantID, tc.member); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID, "members.id": tc.member.ID}).Err())
		})
	}
}

func TestNamespaceUpdateMembership(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		member      *models.Member
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:   "000000000000000000000000",
				Role: authorizer.RoleObserver,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: mongo.ErrUserNotFound},
		},
		{
			description: "succeeds when tenant and user is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:      "6509e169ae6144b2f56bf288",
				Role:    authorizer.RoleAdministrator,
				AddedAt: time.Now(),
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceUpdateMembership(ctx, tc.tenantID, tc.member); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			namespace := new(models.Namespace)
			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID, "members.id": tc.member.ID}).Decode(namespace))
			require.Equal(t, 2, len(namespace.Members))
			require.Equal(t, tc.member.ID, namespace.Members[1].ID)
			require.Equal(t, tc.member.Role, namespace.Members[1].Role)
		})
	}
}

func TestNamespaceDeleteMembership(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		member      *models.Member
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenantID:    "nonexistent",
			member: &models.Member{
				ID: "6509de884238881ac1b2b289",
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: store.ErrNoDocuments},
		},
		{
			description: "fails when member is not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID: "nonexistent",
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: mongo.ErrUserNotFound},
		},
		{
			description: "succeeds when tenant and user is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID: "6509e169ae6144b2f56bf288",
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceDeleteMembership(ctx, tc.tenantID, tc.member); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			namespace := new(models.Namespace)
			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID}).Decode(namespace))
			require.Equal(t, 1, len(namespace.Members))
		})
	}
}
