package services

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type removedDevice struct {
	tenantID string
	uid      string
}

func recordDeviceRemovals(t *testing.T) *[]removedDevice {
	t.Helper()

	override(t, &deviceRemovedHooks, nil)

	removed := &[]removedDevice{}
	OnDeviceRemoved(func(_ context.Context, tenantID, uid string) {
		*removed = append(*removed, removedDevice{tenantID: tenantID, uid: uid})
	})

	return removed
}

func TestDeleteDeviceEndsTheDevice(t *testing.T) {
	at := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	clock.Freeze(t, at)

	ctx := context.Background()
	sc := scope.MustBounded("tenant")

	cases := []struct {
		description string
		device      *models.Device
		mocks       func(storeMock *storemock.MockStore, device *models.Device)
	}{
		{
			description: "an owned accepted device is removed as a team device",
			device:      &models.Device{UID: "uid", TenantID: "tenant", Status: models.DeviceStatusAccepted, OwnerID: "member"},
			mocks: func(storeMock *storemock.MockStore, device *models.Device) {
				removed := *device
				removed.Status = models.DeviceStatusRemoved
				removed.RemovedAt = &at

				storeMock.On("DeviceUpdate", ctx, &removed).Return(nil).Once()
				storeMock.On("DeviceSetOwner", ctx, sc, "uid", "").Return(nil).Once()
				storeMock.On("NamespaceIncrementDeviceCount", ctx, sc, models.DeviceStatusRemoved, int64(1)).Return(nil).Once()
				storeMock.On("NamespaceIncrementDeviceCount", ctx, sc, models.DeviceStatusAccepted, int64(-1)).Return(nil).Once()
			},
		},
		{
			description: "a pending device is deleted",
			device:      &models.Device{UID: "uid", TenantID: "tenant", Status: models.DeviceStatusPending},
			mocks: func(storeMock *storemock.MockStore, device *models.Device) {
				storeMock.On("DeviceDelete", ctx, device).Return(nil).Once()
				storeMock.On("NamespaceIncrementDeviceCount", ctx, sc, models.DeviceStatusPending, int64(-1)).Return(nil).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			removals := recordDeviceRemovals(t)

			storeMock := storemock.NewMockStore(t)
			storeMock.On("DeviceResolve", ctx, sc, store.DeviceUIDResolver, "uid").Return(tc.device, nil).Once()
			tc.mocks(storeMock, tc.device)

			cacheMock := cachemock.NewMockCache(t)
			cacheMock.On("Delete", ctx, "auth_device/uid").Return(nil).Once()

			service := NewService(store.Store(storeMock), privateKey, publicKey, cacheMock)
			require.NoError(t, service.DeleteDevice(ctx, "uid", "tenant"))

			assert.Equal(t, []removedDevice{{tenantID: "tenant", uid: "uid"}}, *removals)
		})
	}
}
