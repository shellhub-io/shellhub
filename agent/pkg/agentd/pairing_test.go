package agentd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAwaitPairing(t *testing.T) {
	start := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

	pending := func(expiresIn int) *models.DevicePairingStatus {
		return &models.DevicePairingStatus{Status: models.DeviceStatusPending, ExpiresIn: expiresIn}
	}

	accepted := &models.DevicePairingStatus{Status: models.DeviceStatusAccepted, TenantID: "tenant1"}

	cases := []struct {
		description string
		expiresIn   int
		polls       []*models.DevicePairingStatus
		expected    string
		expectedErr error
	}{
		{
			description: "returns the tenant once the device is accepted",
			expiresIn:   600,
			polls:       []*models.DevicePairingStatus{pending(600), accepted},
			expected:    "tenant1",
		},
		{
			description: "keeps waiting past the first deadline while the server renews the code",
			expiresIn:   600,
			polls:       []*models.DevicePairingStatus{pending(600), pending(600), pending(600), accepted},
			expected:    "tenant1",
		},
		{
			description: "keeps waiting through a failed poll",
			expiresIn:   600,
			polls:       []*models.DevicePairingStatus{nil, accepted},
			expected:    "tenant1",
		},
		{
			description: "fails once the server stops renewing the code",
			expiresIn:   600,
			polls:       []*models.DevicePairingStatus{pending(600), pending(300), pending(0), pending(0), pending(0)},
			expectedErr: ErrPairingExpired,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			clk := clock.NewFixed(start)

			polls := 0
			status := func() (*models.DevicePairingStatus, error) {
				clk.Advance(5 * time.Minute)

				if polls >= len(tc.polls) {
					return pending(0), nil
				}

				poll := tc.polls[polls]
				polls++

				if poll == nil {
					return nil, errors.New("server unreachable")
				}

				return poll, nil
			}

			tenant, err := AwaitPairing(context.Background(), clk, time.Millisecond, tc.expiresIn, status)
			require.ErrorIs(tt, err, tc.expectedErr)
			assert.Equal(tt, tc.expected, tenant)
		})
	}
}
