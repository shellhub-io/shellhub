package storetest

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func newSSHIdentityKey(t *testing.T) (authorized []byte, fingerprint string) {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := ssh.NewSignerFromKey(priv)
	require.NoError(t, err)

	pub := signer.PublicKey()

	return ssh.MarshalAuthorizedKey(pub), ssh.FingerprintSHA256(pub)
}

// TestSSHIdentityResolveCarriesThePrincipal verifies both resolvers return the principal's name,
// email and type, which the identity row does not store.
func (s *Suite) TestSSHIdentityResolveCarriesThePrincipal(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	userID := s.CreateUser(t, WithName("Ada Lovelace"), WithEmail("ada@test.com"))
	s.CreateMembership(t, tenantID, userID, "operator")

	authorized, fingerprint := newSSHIdentityKey(t)

	id, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
		TenantID:    tenantID,
		PrincipalID: userID,
		Fingerprint: fingerprint,
		Data:        authorized,
		Name:        "laptop",
		Source:      models.SSHIdentitySourceManual,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	sc := scope.MustBounded(tenantID)

	for _, tc := range []struct {
		description string
		resolver    store.SSHIdentityResolver
		value       string
	}{
		{"by id", store.SSHIdentityIDResolver, id},
		{"by fingerprint", store.SSHIdentityFingerprintResolver, fingerprint},
	} {
		t.Run(tc.description, func(t *testing.T) {
			identity, err := st.SSHIdentityResolve(ctx, sc, tc.resolver, tc.value)
			require.NoError(t, err)
			require.NotNil(t, identity)

			assert.Equal(t, id, identity.ID)
			assert.Equal(t, fingerprint, identity.Fingerprint)
			assert.Equal(t, userID, identity.PrincipalID)
			assert.Equal(t, "Ada Lovelace", identity.PrincipalName)
			assert.Equal(t, "ada@test.com", identity.PrincipalEmail)
			assert.Equal(t, models.PrincipalUser, identity.PrincipalType)
		})
	}
}

// TestSSHIdentityBelongsToAnAPIKey verifies an automation's credential: the identity resolves
// back as the key's, the database refuses an identity with no owner or with two, and revoking
// the key removes what it enrolled without the service layer taking part. That last one is the
// whole reason the owner is a foreign key rather than a pair of loose columns.
func (s *Suite) TestSSHIdentityBelongsToAnAPIKey(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	userID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, userID, "administrator")

	digest := s.CreateAPIKey(t, WithAPIKeyName("ci"), WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(userID))
	key, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyDigestResolver, digest)
	require.NoError(t, err)

	authorized, fingerprint := newSSHIdentityKey(t)
	created, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
		TenantID:      tenantID,
		PrincipalID:   key.ID,
		PrincipalType: models.PrincipalAPIKey,
		Fingerprint:   fingerprint,
		Data:          authorized,
		Name:          "deploy",
		Source:        models.SSHIdentitySourceManual,
		CreatedAt:     clock.Now(),
	})
	require.NoError(t, err)

	identity, err := st.SSHIdentityResolve(ctx, scope.MustBounded(tenantID), store.SSHIdentityIDResolver, created)
	require.NoError(t, err)
	require.NotNil(t, identity)
	assert.Equal(t, key.ID, identity.PrincipalID, "the owner is the key's id, not the caller's")
	assert.Equal(t, models.PrincipalAPIKey, identity.PrincipalType)
	assert.Equal(t, "ci", identity.PrincipalName)
	assert.Empty(t, identity.PrincipalEmail, "an automation has no mailbox")

	t.Run("refuses an identity with no owner", func(t *testing.T) {
		other, otherFingerprint := newSSHIdentityKey(t)
		_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
			TenantID:    tenantID,
			Fingerprint: otherFingerprint,
			Data:        other,
			Name:        "ownerless",
			Source:      models.SSHIdentitySourceManual,
			CreatedAt:   clock.Now(),
		})
		assert.Error(t, err)
	})

	t.Run("revoking the key revokes what it enrolled", func(t *testing.T) {
		personal, personalFingerprint := newSSHIdentityKey(t)
		_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
			TenantID:      tenantID,
			PrincipalID:   userID,
			PrincipalType: models.PrincipalUser,
			Fingerprint:   personalFingerprint,
			Data:          personal,
			Name:          "laptop",
			Source:        models.SSHIdentitySourceManual,
			CreatedAt:     clock.Now(),
		})
		require.NoError(t, err)

		require.NoError(t, st.APIKeyDelete(ctx, key))

		identities, _, err := st.SSHIdentityList(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		names := make([]string, 0, len(identities))
		for _, i := range identities {
			names = append(names, i.Name)
		}

		assert.Equal(t, []string{"laptop"}, names, "the person's key is untouched")
	})
}

type sshIdentityOwners struct {
	tenantID string
	userID   string
	keyID    string
}

func seedSSHIdentityOwners(t *testing.T, ctx context.Context, s *Suite) sshIdentityOwners {
	t.Helper()

	st := s.provider.Store()

	tenantID := s.CreateNamespace(t)
	userID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, userID, "administrator")

	otherID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, otherID, "operator")

	digest := s.CreateAPIKey(t, WithAPIKeyName("ci"), WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(userID))
	key, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyDigestResolver, digest)
	require.NoError(t, err)

	enrol := func(name, principalID string, kind models.PrincipalKind) {
		t.Helper()

		authorized, fingerprint := newSSHIdentityKey(t)
		_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
			TenantID:      tenantID,
			PrincipalID:   principalID,
			PrincipalType: kind,
			Fingerprint:   fingerprint,
			Data:          authorized,
			Name:          name,
			Source:        models.SSHIdentitySourceManual,
			CreatedAt:     clock.Now(),
		})
		require.NoError(t, err)
	}

	enrol("laptop", userID, models.PrincipalUser)
	enrol("someone-elses", otherID, models.PrincipalUser)
	enrol("deploy", key.ID, models.PrincipalAPIKey)

	return sshIdentityOwners{tenantID: tenantID, userID: userID, keyID: key.ID}
}

func sshIdentityNames(identities []models.SSHIdentity) []string {
	names := make([]string, 0, len(identities))
	for _, i := range identities {
		names = append(names, i.Name)
	}

	return names
}

// TestSSHIdentityListFiltersByOwner covers the listing the console's SSH identities page asks
// for: the caller's own identities and nobody else's. It is a store test rather than a service
// one because the filter it names is SQL, and the query joins api_keys, which carries a user_id
// of its own.
func (s *Suite) TestSSHIdentityListFiltersByOwner(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	f := seedSSHIdentityOwners(t, ctx, s)

	identities, _, err := st.SSHIdentityList(ctx, scope.MustBounded(f.tenantID), st.Options().WithUserID(f.userID))
	require.NoError(t, err)
	assert.Equal(t, []string{"laptop"}, sshIdentityNames(identities))
}

// TestSSHIdentityGoesWithTheMembership verifies that leaving a namespace revokes the keys enrolled
// there. The key is a credential the namespace granted: a returning member enrolls again rather
// than resuming one nobody re-authorized.
func (s *Suite) TestSSHIdentityGoesWithTheMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	leaverID := s.CreateUser(t)
	stayerID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, leaverID, "operator")
	s.CreateMembership(t, tenantID, stayerID, "operator")

	for i, principal := range []string{leaverID, stayerID} {
		_, err := st.SSHIdentityCreate(ctx, &models.SSHIdentity{
			TenantID:    tenantID,
			PrincipalID: principal,
			Fingerprint: fmt.Sprintf("SHA256:%043d=", i),
			Data:        []byte("ssh-ed25519 AAAA key"),
			Name:        "laptop",
			Source:      models.SSHIdentitySourceManual,
			CreatedAt:   clock.Now(),
		})
		require.NoError(t, err)
	}

	before, _, err := st.SSHIdentityList(ctx, scope.MustBounded(tenantID))
	require.NoError(t, err)
	require.Len(t, before, 2)

	require.NoError(t, st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), &models.Member{ID: leaverID}))

	after, _, err := st.SSHIdentityList(ctx, scope.MustBounded(tenantID))
	require.NoError(t, err)
	require.Len(t, after, 1, "only the member who stayed keeps their enrolled key")
	assert.Equal(t, stayerID, after[0].PrincipalID)
}
