package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	storemocks "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/require"
)

func TestService_DevicesHeartbeat(t *testing.T) {
	storeMock := new(storemocks.Store)
	clockMock := new(clockmock.Clock)

	clock.DefaultBackend = clockMock

	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		payload       []byte
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when cannot set the status",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(0), errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds with duplicated IDs",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001\n0000000000000000000000000000000000000000000000000000000000000000"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(2), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(2), nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)
			require.Equal(tt, tc.expected, s.DevicesHeartbeat()(ctx, tc.payload))
		})
	}
}
