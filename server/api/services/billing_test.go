package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockBillingProvider stands in for BillingProvider. It lives here rather than in the generated
// mocks package because this package's own tests use it: services/mocks imports services, so an
// internal test importing it would form a cycle.
type mockBillingProvider struct {
	mock.Mock
}

func (m *mockBillingProvider) Evaluate(ctx context.Context, tenant string) (*models.BillingEvaluation, error) {
	args := m.Called(ctx, tenant)

	evaluation, _ := args.Get(0).(*models.BillingEvaluation)

	return evaluation, args.Error(1)
}

func (m *mockBillingProvider) Report(ctx context.Context, tenant string, action BillingAction) error {
	args := m.Called(ctx, tenant, action)

	return args.Error(0)
}

func TestValidateBillingForDeviceAcceptance(t *testing.T) {
	ctx := context.Background()

	namespace := &models.Namespace{ //nolint:exhaustruct
		TenantID: "00000000-0000-4000-0000-000000000000",
		Billing:  models.NewBilling("cus_H9J5n2eZvKYlo2C7X1QX2Qg"),
	}

	cases := []struct {
		description string
		evaluation  *models.BillingEvaluation
		expected    error
	}{
		{
			description: "succeeds when the namespace can accept the device",
			evaluation:  &models.BillingEvaluation{CanAccept: true, CanConnect: true, Blocked: ""},
			expected:    nil,
		},
		{
			description: "reports a device limit when the namespace used its whole allowance",
			evaluation: &models.BillingEvaluation{
				CanAccept:  false,
				CanConnect: true,
				Blocked:    models.BillingBlockedQuota,
			},
			expected: ErrDeviceLimit,
		},
		{
			description: "reports a billing block when the subscription denies the device",
			evaluation: &models.BillingEvaluation{
				CanAccept:  false,
				CanConnect: false,
				Blocked:    models.BillingBlockedSubscription,
			},
			expected: ErrDeviceBillingBlocked,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			billing := new(mockBillingProvider)
			billing.On("Evaluate", ctx, namespace.TenantID).Return(tc.evaluation, nil).Once()

			s := &service{billing: billing} //nolint:exhaustruct

			require.ErrorIs(t, s.validateBillingForDeviceAcceptance(ctx, namespace), tc.expected)
			billing.AssertExpectations(t)
		})
	}
}

// TestErrDeviceBillingBlocked asserts that ErrDeviceBillingBlocked is a distinct sentinel from
// ErrDeviceLimit. They share a layer and a code, so only the message separates them, and a caller
// that tells a namespace to free a device slot must not match a billing block.
func TestErrDeviceBillingBlocked(t *testing.T) {
	assert.False(t, errors.Is(ErrDeviceBillingBlocked, ErrDeviceLimit))
	assert.False(t, errors.Is(ErrDeviceLimit, ErrDeviceBillingBlocked))
}
