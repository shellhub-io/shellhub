package pgprovider_test

import (
	"context"
	"os"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/pg/dbtest"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.Run(m))
}

func countUsers(t *testing.T, ctx context.Context, provider *pgprovider.Provider) int {
	t.Helper()

	var count int
	require.NoError(t, provider.DB().QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&count))

	return count
}

func TestProvidersDoNotShareRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	first, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, first.Close(t)) })

	second, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, second.Close(t)) })

	_, err = first.Store().UserCreate(ctx, &models.User{
		Origin:   models.UserOriginLocal,
		Status:   models.UserStatusConfirmed,
		UserData: models.UserData{Name: "isolated", Email: "isolated@example.com", Username: "isolated"},
		Password: models.UserPassword{Hash: "hash"},
	})
	require.NoError(t, err)

	require.Equal(t, 1, countUsers(t, ctx, first))
	require.Equal(t, 0, countUsers(t, ctx, second))
}

func TestProvidersAtAVersionStartThere(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for range 2 {
		provider, err := pgprovider.NewProviderAt(ctx, 3)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, provider.Close(t)) })

		var latest string
		require.NoError(t, provider.DB().QueryRowContext(ctx, "SELECT max(name) FROM bun_migrations").Scan(&latest))
		require.Equal(t, "003", latest)

		require.NoError(t, provider.ApplyNext(ctx))
	}
}
