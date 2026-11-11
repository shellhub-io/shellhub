package services

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/require"
)

func TestSetDeviceCustomFieldConcurrentCallsRespectTheLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	e := setupEnrollmentE2E(t)

	uid, err := e.st.DeviceCreate(ctx, scope.MustBounded(e.tenantID), &models.Device{
		UID:       "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
		TenantID:  e.tenantID,
		Name:      "device",
		Status:    models.DeviceStatusAccepted,
		Identity:  &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
		PublicKey: "key",
	})
	require.NoError(t, err)

	set := func(key string) error {
		return e.svc.SetDeviceCustomField(ctx, &requests.DeviceSetCustomField{
			TenantID:    e.tenantID,
			DeviceParam: requests.DeviceParam{UID: uid},
			Key:         key,
			Value:       "v",
		})
	}

	for i := range maxCustomFieldsPerDevice - 1 {
		require.NoError(t, set(fmt.Sprintf("seed%d", i)))
	}

	const attempts = 16
	errs := make([]error, attempts)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			<-start
			errs[i] = set(fmt.Sprintf("race%d", i))
		})
	}

	close(start)
	wg.Wait()

	succeeded := 0
	for _, err := range errs {
		if err == nil {
			succeeded++

			continue
		}

		require.ErrorIs(t, err, ErrDeviceCustomFieldLimitReached)
	}
	require.Equal(t, 1, succeeded)

	device, err := e.st.DeviceResolve(ctx, scope.MustBounded(e.tenantID), store.DeviceUIDResolver, uid)
	require.NoError(t, err)
	require.Len(t, device.CustomFields, maxCustomFieldsPerDevice)
}
