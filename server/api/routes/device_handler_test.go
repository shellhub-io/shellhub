package routes

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDeviceHandler(t *testing.T) {
	cases := []struct {
		description    string
		sc             scope.Scope
		uid            string
		requiredMocks  func(*mocks.MockService)
		expectedDevice *models.Device
		expectedErr    error
	}{
		{
			description: "passes the namespace scope it was given through to the service",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			uid:         "uid",
			requiredMocks: func(service *mocks.MockService) {
				service.
					On("GetDevice", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("uid")).
					Return(&models.Device{UID: "uid"}, nil).
					Once()
			},
			expectedDevice: &models.Device{UID: "uid"},
		},
		{
			description: "reports the service's failure unchanged",
			sc:          scope.NewUnbounded("the admin console reads every namespace"),
			uid:         "missing",
			requiredMocks: func(service *mocks.MockService) {
				service.
					On("GetDevice", gomock.Anything, scope.NewUnbounded("the admin console reads every namespace"), models.UID("missing")).
					Return(nil, svc.ErrDeviceNotFound).
					Once()
			},
			expectedErr: svc.ErrDeviceNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			tc.requiredMocks(service)

			handler := NewHandler(service)

			device, err := handler.GetDevice(t.Context(), tc.sc, gateway.Actor{ID: "user-id"},
				&requests.DeviceGet{DeviceParam: requests.DeviceParam{UID: tc.uid}})

			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedDevice, device)
		})
	}
}
