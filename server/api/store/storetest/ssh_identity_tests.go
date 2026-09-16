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
			assert.Equal(t, models.UserTypeHuman, identity.PrincipalType)
		})
	}
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
