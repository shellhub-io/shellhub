package mongo_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestEnrichMembersData(t *testing.T) {
	type Expected struct {
		emails []string
		err    error
	}

	cases := []struct {
		description string
		tenant      string
		ctx         func() context.Context
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when context does not have db in values",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.Background()
			},
			fixtures: []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				emails: []string{},
				err:    errors.New("db not found in context"),
			},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "db", db) //nolint:revive
			},
			fixtures: []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				emails: []string{"john.doe@test.com", "maria.garcia@test.com"},
				err:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := tc.ctx()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			ns := &models.Namespace{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Members:  []models.Member{{ID: "507f1f77bcf86cd799439011"}, {ID: "6509e169ae6144b2f56bf288"}},
			}

			err := s.Options().EnrichMembersData()(ctx, ns)
			require.Equal(tt, tc.expected.err, err)

			if err == nil {
				for _, m := range ns.Members {
					require.Equal(tt, true, slices.Contains(tc.expected.emails, m.Email))
				}
			}
		})
	}
}
