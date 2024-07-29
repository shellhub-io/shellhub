package services

import (
	"context"
	"errors"
	"testing"
	"time"

	storemocks "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestService_DevicesHeartbeat(t *testing.T) {
	storeMock := new(storemocks.Store)

	cases := []struct {
		description   string
		payload       []byte
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when cannot set the status",
			payload:     []byte("00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000000=1721912837\n00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000001=1721912837"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOnline", ctx, []models.ConnectedDevice{
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000001",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
					}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds but one payload does not have ':'",
			payload:     []byte("00000000-0000-4000-0000-0000000000000000000000000000000000000000000000000000000000000000000000000000=1721912837\n00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000001=1721912837"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOnline", ctx, []models.ConnectedDevice{
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000001",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
					}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds but one payload does not have '='",
			payload:     []byte("00000000-0000-4000-0000-000000000000:00000000000000000000000000000000000000000000000000000000000000001721912837\n00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000001=1721912837"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOnline", ctx, []models.ConnectedDevice{
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000001",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
					}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds",
			payload:     []byte("00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000000=1721912837\n00000000-0000-4000-0000-000000000000:0000000000000000000000000000000000000000000000000000000000000001=1721912837"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOnline", ctx, []models.ConnectedDevice{
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
						{
							UID:      "0000000000000000000000000000000000000000000000000000000000000001",
							TenantID: "00000000-0000-4000-0000-000000000000",
							LastSeen: time.Unix(1721912837, 0),
						},
					}).
					Return(nil).
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

func TestService_CleanupSessions(t *testing.T) {
	storeMock := new(storemocks.Store)
	clockMock := new(clockmocks.Clock)

	clock.DefaultBackend = clockMock
	now := time.Now()
	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		retention     int
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails",
			retention:   30,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("SessionDeleteRecordFrameByDate", ctx, now.UTC().AddDate(0, 0, 30*(-1))).
					Return(int64(0), int64(0), errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			retention:   30,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("SessionDeleteRecordFrameByDate", ctx, now.UTC().AddDate(0, 0, 30*(-1))).
					Return(int64(30), int64(0), nil).
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
			require.Equal(tt, tc.expected, s.CleanupSessions(tc.retention)(ctx))
		})
	}
}
