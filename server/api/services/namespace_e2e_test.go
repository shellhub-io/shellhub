package services

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/require"
)

func TestCreateNamespaceConcurrentCallsRespectOwnershipLimit(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	userID, err := provider.Store().UserCreate(ctx, &models.User{
		UserData: models.UserData{
			Name:     "owner",
			Username: "owner",
			Email:    "owner@example.com",
		},
		Password:      models.UserPassword{Hash: "hashedpassword"},
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: 1,
		CreatedAt:     provider.Clock().Now(),
	})
	require.NoError(t, err)

	svc := NewService(provider.Store(), privateKey, publicKey, storecache.NewNullCache(), WithClock(provider.Clock()))

	const attempts = 16
	errs := make([]error, attempts)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			<-start
			_, errs[i] = svc.CreateNamespace(ctx, &requests.NamespaceCreate{
				UserID: userID,
				Name:   fmt.Sprintf("namespace%d", i),
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

		require.ErrorIs(t, err, ErrNamespaceLimitReached)
	}
	require.Equal(t, 1, succeeded)

	var owned int
	require.NoError(t, provider.DB().NewSelect().Table("namespaces").ColumnExpr("count(*)").Where("owner_id = ?", userID).Scan(ctx, &owned))
	require.Equal(t, 1, owned)
}
