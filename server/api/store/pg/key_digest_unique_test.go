package pg_test

import (
	"context"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const collidingDigest = "89d07480e9d0f39ee0f1f2f00d5d6a0c8bb3f6f4b2f5b6f2e0a1c3d4e5f60718"

type keyDigestFixture struct {
	provider *pgprovider.Provider
	st       store.Store
	victim   string
	attacker string
	users    map[string]string
}

func setupKeyDigest(t *testing.T) *keyDigestFixture {
	t.Helper()
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	return seedKeyDigest(t, ctx, provider)
}

func seedKeyDigest(t *testing.T, ctx context.Context, provider *pgprovider.Provider) *keyDigestFixture {
	t.Helper()

	st := provider.Store()
	db := provider.DB()
	f := &keyDigestFixture{provider: provider, st: st, users: map[string]string{}}

	const (
		victimOwner    = "11111111-1111-4111-8111-111111111111"
		victimTenant   = "22222222-2222-4222-8222-222222222222"
		attackerOwner  = "33333333-3333-4333-8333-333333333333"
		attackerTenant = "44444444-4444-4444-8444-444444444444"
	)

	mk := func(name, owner, tenant string) {
		t.Helper()

		seedUser(t, ctx, db, owner, name)
		seedNamespace(t, ctx, db, tenant, name, owner, fixtureTime)

		f.users[tenant] = owner
	}

	mk("victimns", victimOwner, victimTenant)
	mk("attackerns", attackerOwner, attackerTenant)

	f.victim = victimTenant
	f.attacker = attackerTenant

	return f
}

func (f *keyDigestFixture) apiKey(tenant, name string) *models.APIKey {
	return &models.APIKey{
		Digest: collidingDigest, Name: name, TenantID: tenant,
		Role: authorizer.RoleOwner, CreatedBy: f.users[tenant],
		CreatedAt: clock.Now(), UpdatedAt: clock.Now(), ExpiresIn: -1,
	}
}

// TestAPIKeyDigestIsGloballyUnique locks that a digest already held by one namespace cannot be
// stored in a second, which is what lets the authenticator resolve one by digest alone.
func TestAPIKeyDigestIsGloballyUnique(t *testing.T) {
	ctx := context.Background()
	f := setupKeyDigest(t)

	_, err := f.st.APIKeyCreate(ctx, f.apiKey(f.victim, "prodkey"))
	require.NoError(t, err)

	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.attacker, "mallorykey"))
	require.ErrorIs(t, err, store.ErrDuplicate,
		"a digest already held by another namespace must be refused, not stored alongside it")
}

// TestProvisioningKeyDigestIsGloballyUnique locks the same invariant for provisioning keys, which
// provisioningKeyTenant resolves by digest alone.
func TestProvisioningKeyDigestIsGloballyUnique(t *testing.T) {
	ctx := context.Background()
	f := setupKeyDigest(t)

	mk := func(tenant, name string) *models.ProvisioningKey {
		return &models.ProvisioningKey{
			ID: collidingDigest, Name: name, TenantID: tenant,
			Mode: models.ProvisioningKeyModeAutomatic, Type: models.ProvisioningKeyTypeUser,
			Reusable: true, Tags: []string{}, CreatedBy: f.users[tenant],
		}
	}

	_, err := f.st.ProvisioningKeyCreate(ctx, mk(f.victim, "prodkey"))
	require.NoError(t, err)

	_, err = f.st.ProvisioningKeyCreate(ctx, mk(f.attacker, "mallorykey"))
	require.ErrorIs(t, err, store.ErrDuplicate)
}

// TestAPIKeyResolveRefusesAmbiguousDigest locks that the resolver refuses a digest matching more
// than one row instead of picking one, with the index dropped to reach a state it otherwise forbids.
func TestAPIKeyResolveRefusesAmbiguousDigest(t *testing.T) {
	ctx := context.Background()
	f := setupKeyDigest(t)

	_, err := f.provider.DB().ExecContext(ctx, "DROP INDEX api_keys_key_digest_unique")
	require.NoError(t, err)

	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.victim, "prodkey"))
	require.NoError(t, err)
	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.attacker, "mallorykey"))
	require.NoError(t, err)

	sc := scope.NewUnbounded("test: the authenticator resolves a digest with no namespace to bound by")
	apiKey, err := f.st.APIKeyResolve(ctx, sc, store.APIKeyDigestResolver, collidingDigest)
	assert.Nil(t, apiKey)
	require.ErrorIs(t, err, store.ErrAmbiguous,
		"a digest matching two namespaces must not authenticate into either of them")
}

// TestKeyDigestUniqueMigrationRevokesCollisions locks migration 023 against data that predates it:
// both sides of a collision go, a key with an unshared digest stays, and the index is left behind.
func TestKeyDigestUniqueMigrationRevokesCollisions(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 22)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	f := seedKeyDigest(t, ctx, provider)

	seed := func(tenant, name, digest string) {
		t.Helper()

		_, err := provider.DB().ExecContext(ctx, `
			INSERT INTO api_keys (key_digest, namespace_id, name, role, user_id, created_at, updated_at, expires_in)
			VALUES (?, ?, ?, 'owner', ?, now(), now(), -1)
		`, digest, tenant, name, f.users[tenant])
		require.NoError(t, err)
	}

	seed(f.victim, "prodkey", collidingDigest)
	seed(f.attacker, "mallorykey", collidingDigest)
	seed(f.victim, "lonekey", strings.Repeat("a", 64))

	require.NoError(t, provider.ApplyNext(ctx), "the migration must survive pre-existing collisions")

	var names []string
	require.NoError(t, f.provider.DB().NewRaw("SELECT name FROM api_keys ORDER BY name").Scan(ctx, &names))
	assert.Equal(t, []string{"lonekey"}, names,
		"both sides of the collision are revoked; a key with an unshared digest is untouched")

	seed(f.victim, "reissued", collidingDigest)

	_, err = provider.DB().ExecContext(ctx, `
		INSERT INTO api_keys (key_digest, namespace_id, name, role, user_id, created_at, updated_at, expires_in)
		VALUES (?, ?, 'recollide', 'owner', ?, now(), now(), -1)
	`, collidingDigest, f.attacker, f.users[f.attacker])
	require.ErrorContains(t, err, "api_keys_key_digest_unique", "the migration leaves the index in place")
}
