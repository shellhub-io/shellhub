package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersOnboardingSurveyMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 52)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const existing = "11111111-1111-4111-8111-111111111111"

	seedUser(t, ctx, db, existing, "existing")

	require.NoError(t, provider.ApplyNext(ctx), "053 must apply cleanly")

	var outcome string
	require.NoError(t, db.NewRaw(`SELECT onboarding_survey FROM users WHERE id = ?`, existing).Scan(ctx, &outcome))
	assert.Equal(t, "pending", outcome, "a user from before 053 sees the survey once")

	_, err = db.ExecContext(ctx, `UPDATE users SET onboarding_survey = 'maybe' WHERE id = ?`, existing)
	assert.Error(t, err, "only pending, answered and skipped are stored")
}
