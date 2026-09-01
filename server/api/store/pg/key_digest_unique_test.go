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

	st := provider.Store()
	f := &keyDigestFixture{provider: provider, st: st, users: map[string]string{}}

	mk := func(name string) string {
		owner, err := st.UserCreate(ctx, &models.User{
			Origin: models.UserOriginLocal, Status: models.UserStatusConfirmed, MaxNamespaces: -1,
			UserData: models.UserData{Name: name, Email: name + "@example.com", Username: name},
			Password: models.UserPassword{Hash: "hash"},
		})
		require.NoError(t, err)

		tenant, err := st.NamespaceCreate(ctx, &models.Namespace{
			Name: name, Owner: owner, MaxDevices: -1,
			Members:  []models.Member{{ID: owner, Role: authorizer.RoleOwner}},
			Settings: &models.NamespaceSettings{},
		})
		require.NoError(t, err)
		f.users[tenant] = owner

		return tenant
	}

	f.victim = mk("victimns")
	f.attacker = mk("attackerns")

	return f
}

func (f *keyDigestFixture) apiKey(tenant, name string) *models.APIKey {
	return &models.APIKey{
		ID: collidingDigest, Name: name, TenantID: tenant,
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

// TestInstallKeyDigestIsGloballyUnique locks the same invariant for install keys, which
// installKeyTenant resolves by digest alone.
func TestInstallKeyDigestIsGloballyUnique(t *testing.T) {
	ctx := context.Background()
	f := setupKeyDigest(t)

	mk := func(tenant, name string) *models.InstallKey {
		return &models.InstallKey{
			ID: collidingDigest, Name: name, TenantID: tenant,
			Mode: models.InstallKeyModeAutomatic, Type: models.InstallKeyTypeUser,
			Reusable: true, Tags: []string{}, CreatedBy: f.users[tenant],
		}
	}

	_, err := f.st.InstallKeyCreate(ctx, mk(f.victim, "prodkey"))
	require.NoError(t, err)

	_, err = f.st.InstallKeyCreate(ctx, mk(f.attacker, "mallorykey"))
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
	apiKey, err := f.st.APIKeyResolve(ctx, sc, store.APIKeyIDResolver, collidingDigest)
	assert.Nil(t, apiKey)
	require.ErrorIs(t, err, store.ErrAmbiguous,
		"a digest matching two namespaces must not authenticate into either of them")
}

// TestKeyDigestUniqueMigrationRevokesCollisions locks migration 023 against data that predates it:
// both sides of a collision go, a key with an unshared digest stays, and the index is left behind.
func TestKeyDigestUniqueMigrationRevokesCollisions(t *testing.T) {
	ctx := context.Background()
	f := setupKeyDigest(t)

	for _, stmt := range migrationStatements(t, "023_key_digest_globally_unique.tx.down.sql") {
		_, err := f.provider.DB().ExecContext(ctx, stmt)
		require.NoError(t, err)
	}

	_, err := f.st.APIKeyCreate(ctx, f.apiKey(f.victim, "prodkey"))
	require.NoError(t, err)
	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.attacker, "mallorykey"))
	require.NoError(t, err)

	lone := f.apiKey(f.victim, "lonekey")
	lone.ID = strings.Repeat("a", 64)
	_, err = f.st.APIKeyCreate(ctx, lone)
	require.NoError(t, err)

	for _, stmt := range migrationStatements(t, "023_key_digest_globally_unique.tx.up.sql") {
		_, err := f.provider.DB().ExecContext(ctx, stmt)
		require.NoError(t, err, "the migration must survive pre-existing collisions")
	}

	var names []string
	require.NoError(t, f.provider.DB().NewRaw("SELECT name FROM api_keys ORDER BY name").Scan(ctx, &names))
	assert.Equal(t, []string{"lonekey"}, names,
		"both sides of the collision are revoked; a key with an unshared digest is untouched")

	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.victim, "reissued"))
	require.NoError(t, err)
	_, err = f.st.APIKeyCreate(ctx, f.apiKey(f.attacker, "recollide"))
	require.ErrorIs(t, err, store.ErrDuplicate, "the migration leaves the index in place")
}
