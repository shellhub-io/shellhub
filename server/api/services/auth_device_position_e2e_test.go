package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	geoipmocks "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthDevicePositionOnReconnect(t *testing.T) {
	saoPaulo := geoip.Position{Latitude: -23.55, Longitude: -46.63}
	lisbon := geoip.Position{Latitude: 38.72, Longitude: -9.14}

	cases := []struct {
		description string
		registered  geoip.Position
		reconnectIP string
		lookup      geoip.Position
		lookupErr   error
		expected    models.DevicePosition
	}{
		{
			description: "a device registered without a position gets one when it reconnects",
			registered:  geoip.Position{},
			reconnectIP: "203.0.113.7",
			lookup:      saoPaulo,
			expected:    models.DevicePosition{Latitude: -23.55, Longitude: -46.63},
		},
		{
			description: "a device that moved shows where it connects from now",
			registered:  saoPaulo,
			reconnectIP: "198.51.100.9",
			lookup:      lisbon,
			expected:    models.DevicePosition{Latitude: 38.72, Longitude: -9.14},
		},
		{
			description: "a failed lookup keeps the stored position",
			registered:  saoPaulo,
			reconnectIP: "198.51.100.9",
			lookupErr:   errors.New("lookup failed"),
			expected:    models.DevicePosition{Latitude: -23.55, Longitude: -46.63},
		},
		{
			description: "a lookup that cannot place the address keeps the stored position",
			registered:  saoPaulo,
			reconnectIP: "198.51.100.9",
			lookup:      geoip.Position{},
			expected:    models.DevicePosition{Latitude: -23.55, Longitude: -46.63},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			locator := geoipmocks.NewMockLocator(t)
			locator.EXPECT().GetPosition(mock.Anything).Return(tc.registered, nil).Once()
			locator.EXPECT().GetPosition(mock.Anything).Return(tc.lookup, tc.lookupErr).Maybe()

			e := setupEnrollmentE2EWith(t, WithLocator(locator))
			ctx := context.Background()
			request := e.authRequest("aa:00:00:00:93:01", "")

			registered, err := e.svc.AuthDevice(ctx, request, "203.0.113.7")
			require.NoError(t, err)
			require.NoError(t, e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{TenantID: e.tenantID, UID: registered.UID, Status: string(models.DeviceStatusAccepted)}))

			_, err = e.svc.AuthDevice(ctx, request, tc.reconnectIP)
			require.NoError(t, err)

			device, err := e.st.DeviceResolve(ctx, scope.MustBounded(e.tenantID), store.DeviceUIDResolver, registered.UID)
			require.NoError(t, err)
			require.NotNil(t, device.Position)
			assert.Equal(t, tc.expected, *device.Position)
		})
	}
}
