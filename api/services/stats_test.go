package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetStats(t *testing.T) {
	storeMock := &mocks.Store{}

	ctx := context.Background()

	cases := []struct {
		description   string
		req           *requests.GetStats
		expectedStats *models.Stats
		expectedError error
		requiredMocks func()
	}{
		{
			description: "fail when store returns error",
			req: &requests.GetStats{
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			expectedStats: nil,
			expectedError: errors.New("store error"),
			requiredMocks: func() {
				storeMock.
					On("GetStats", gomock.Anything, "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("store error")).
					Once()
			},
		},
		{
			description: "success when getting stats without tenantID",
			req: &requests.GetStats{
				TenantID: "",
			},
			expectedStats: &models.Stats{
				RegisteredDevices: 10,
				OnlineDevices:     5,
				ActiveSessions:    15,
				PendingDevices:    2,
				RejectedDevices:   1,
			},
			expectedError: nil,
			requiredMocks: func() {
				storeMock.On("GetStats", gomock.Anything, "").
					Return(
						&models.Stats{
							RegisteredDevices: 10,
							OnlineDevices:     5,
							ActiveSessions:    15,
							PendingDevices:    2,
							RejectedDevices:   1,
						},
						nil,
					).
					Once()
			},
		},
		{
			description: "success when getting stats with tenantID",
			req: &requests.GetStats{
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			expectedStats: &models.Stats{
				RegisteredDevices: 3,
				OnlineDevices:     2,
				ActiveSessions:    5,
				PendingDevices:    1,
				RejectedDevices:   0,
			},
			expectedError: nil,
			requiredMocks: func() {
				storeMock.
					On("GetStats", gomock.Anything, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Stats{
							RegisteredDevices: 3,
							OnlineDevices:     2,
							ActiveSessions:    5,
							PendingDevices:    1,
							RejectedDevices:   0,
						},
						nil,
					).
					Once()
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			stats, err := s.GetStats(ctx, tc.req)
			assert.Equal(t, tc.expectedStats, stats)
			assert.Equal(t, tc.expectedError, err)
		})
	}

	storeMock.AssertExpectations(t)
}
