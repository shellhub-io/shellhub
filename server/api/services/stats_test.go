package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetStats(t *testing.T) {
	cases := []struct {
		description   string
		sc            scope.Scope
		expectedStats *models.Stats
		expectedError error
	}{
		{
			description:   "reads the counts within a namespace",
			sc:            scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			expectedStats: &models.Stats{RegisteredDevices: 3, OnlineDevices: 2, ActiveSessions: 5, PendingDevices: 1},
		},
		{
			description:   "reads the counts across the instance when the scope spans it",
			sc:            scope.NewUnbounded("test"),
			expectedStats: &models.Stats{RegisteredDevices: 10, OnlineDevices: 5, ActiveSessions: 15, PendingDevices: 2, RejectedDevices: 1},
		},
		{
			description:   "returns the store's error",
			sc:            scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			expectedError: errors.New("store error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := mocks.NewMockStore(t)
			storeMock.On("GetStats", gomock.Anything, tc.sc).Return(tc.expectedStats, tc.expectedError).Once()

			s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

			stats, err := s.GetStats(context.Background(), tc.sc)
			assert.Equal(t, tc.expectedStats, stats)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
