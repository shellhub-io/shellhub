package scope_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/stretchr/testify/require"
)

func TestNewBounded(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		expected    scope.Scope
		expectedErr error
	}{
		{
			description: "fails when the tenant ID is empty",
			tenantID:    "",
			expected:    scope.Scope{},
			expectedErr: scope.ErrEmptyTenantID,
		},
		{
			description: "succeeds with a tenant ID",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			expected:    scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			sc, err := scope.NewBounded(tc.tenantID)
			require.ErrorIs(tt, err, tc.expectedErr)
			require.Equal(tt, tc.expected, sc)
		})
	}
}

func TestScopeAccessors(t *testing.T) {
	cases := []struct {
		description        string
		scope              scope.Scope
		expectedBounded    bool
		expectedValid      bool
		expectedTenantID   string
		expectedReason     string
		expectedStringRepr string
	}{
		{
			description:        "the zero value is neither bounded nor unbounded",
			scope:              scope.Scope{},
			expectedBounded:    false,
			expectedValid:      false,
			expectedTenantID:   "",
			expectedReason:     "",
			expectedStringRepr: "scope(invalid)",
		},
		{
			description:        "a bounded scope carries its tenant ID and no reason",
			scope:              scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			expectedBounded:    true,
			expectedValid:      true,
			expectedTenantID:   "00000000-0000-4000-0000-000000000000",
			expectedReason:     "",
			expectedStringRepr: "scope(bounded:00000000-0000-4000-0000-000000000000)",
		},
		{
			description:        "an unbounded scope carries its reason and no tenant ID",
			scope:              scope.NewUnbounded("devices are identified by their public key alone"),
			expectedBounded:    false,
			expectedValid:      true,
			expectedTenantID:   "",
			expectedReason:     "devices are identified by their public key alone",
			expectedStringRepr: "scope(unbounded:devices are identified by their public key alone)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			require.Equal(tt, tc.expectedBounded, tc.scope.IsBounded())
			require.Equal(tt, tc.expectedValid, tc.scope.IsValid())
			require.Equal(tt, tc.expectedTenantID, tc.scope.TenantID())
			require.Equal(tt, tc.expectedReason, tc.scope.Reason())
			require.Equal(tt, tc.expectedStringRepr, tc.scope.String())
		})
	}
}

func TestMustBoundedPanicsOnEmptyTenantID(t *testing.T) {
	require.Panics(t, func() { scope.MustBounded("") })
}
