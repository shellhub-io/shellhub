package services

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/require"
)

func TestSetupConcurrentCallsCreateOneAdministrator(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	svc := NewService(provider.Store(), privateKey, publicKey, storecache.NewNullCache(), WithClock(provider.Clock()))

	const attempts = 16
	errs := make([]error, attempts)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			<-start
			_, errs[i] = svc.Setup(ctx, requests.Setup{
				Name:      fmt.Sprintf("admin%d", i),
				Username:  fmt.Sprintf("admin%d", i),
				Email:     fmt.Sprintf("admin%d@example.com", i),
				Password:  "password123",
				Namespace: fmt.Sprintf("namespace%d", i),
			})
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

		require.ErrorIs(t, err, ErrSetupCompleted)
	}
	require.Equal(t, 1, succeeded)

	var admins int
	require.NoError(t, provider.DB().NewSelect().Table("users").ColumnExpr("count(*)").Where("admin").Scan(ctx, &admins))
	require.Equal(t, 1, admins)
}

func TestSetupFailureLeavesNoUser(t *testing.T) {
	ctx := context.Background()
	e := setupEnrollmentE2E(t)

	_, err := e.svc.Setup(ctx, requests.Setup{
		Name:      "admin",
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  "password123",
		Namespace: "e2e",
	})
	require.ErrorIs(t, err, ErrNamespaceDuplicated)

	_, err = e.st.UserResolve(ctx, store.UserUsernameResolver, "admin")
	require.ErrorIs(t, err, store.ErrNoDocuments)
}
