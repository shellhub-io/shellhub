package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserList exercises UserList against the store under test.
func (s *Suite) TestUserList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when users are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithUsername("user1"), WithEmail("user1@test.com"))
		s.CreateUser(t, WithUsername("user2"), WithEmail("user2@test.com"))
		s.CreateUser(t, WithUsername("user3"), WithEmail("user3@test.com"))

		users, count, err := st.UserList(ctx,
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
		assert.Len(t, users, 3)
	})
}

// TestUserResolve exercises UserResolve against the store under test.
func (s *Suite) TestUserResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user not found by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		nonExistentID := "00000000-0000-0000-0000-000000000000"

		user, err := st.UserResolve(ctx, store.UserIDResolver, nonExistentID)
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, user)
	})

	t.Run("succeeds resolving user by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("testuser"), WithUserStatus(models.UserStatusConfirmed))

		user, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, models.UserStatusConfirmed, user.Status)
	})

	t.Run("fails when user not found by email", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		user, err := st.UserResolve(ctx, store.UserEmailResolver, "nonexistent@test.com")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, user)
	})

	t.Run("succeeds resolving user by email", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithEmail("jane.smith@test.com"), WithUserStatus(models.UserStatusConfirmed))

		user, err := st.UserResolve(ctx, store.UserEmailResolver, "jane.smith@test.com")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, models.UserStatusConfirmed, user.Status)
	})

	t.Run("fails when user not found by username", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		user, err := st.UserResolve(ctx, store.UserUsernameResolver, "nonexistent_user")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, user)
	})

	t.Run("succeeds resolving user by username", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("bob_johnson"), WithUserStatus(models.UserStatusConfirmed))

		user, err := st.UserResolve(ctx, store.UserUsernameResolver, "bob_johnson")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, models.UserStatusConfirmed, user.Status)
	})

	t.Run("fails when user not found by external ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t)

		user, err := st.UserResolve(ctx, store.UserExternalIDResolver, "idp-subject")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, user)
	})

	t.Run("succeeds resolving user by external ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithExternalID("other-subject"))
		userID := s.CreateUser(t, WithExternalID("idp-subject"))

		user, err := st.UserResolve(ctx, store.UserExternalIDResolver, "idp-subject")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "idp-subject", user.ExternalID)
	})
}

// TestUserCreatePasswordRoundTrip locks that a password survives storage and still verifies, so a hashing
// change cannot silently lock every account out.
func (s *Suite) TestUserCreatePasswordRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("password can be verified after round-trip through database", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		password, err := models.HashUserPassword("testpassword")
		require.NoError(t, err)

		user := &models.User{
			UserData: models.UserData{
				Name:     "roundtrip user",
				Username: "roundtrip_user",
				Email:    "roundtrip@test.com",
			},
			Password: password,
		}

		userID, err := st.UserCreate(ctx, user)
		require.NoError(t, err)
		require.NotEmpty(t, userID)

		resolved, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		require.NotNil(t, resolved)

		assert.True(t, resolved.Password.Compare("testpassword"), "correct password must match after round-trip")
		assert.False(t, resolved.Password.Compare("wrongpassword"), "wrong password must not match")
	})
}

// TestUserCreateDuplicate verifies that UserCreate enforces unique constraints on
// email and username at the real-database level, returning store.ErrDuplicate with
// the correct field name reported by store.DuplicatedField.
func (s *Suite) TestUserCreateDuplicate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("duplicate email is rejected with ErrDuplicate and field=email", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithUsername("baseline_user"), WithEmail("shared@test.com"))

		_, err := st.UserCreate(ctx, &models.User{
			UserData: models.UserData{
				Name:     "second user",
				Username: "other_user",
				Email:    "shared@test.com",
			},
			Password: models.UserPassword{Hash: "somehash"},
		})
		require.ErrorIs(t, err, store.ErrDuplicate)

		field, ok := store.DuplicatedField(err)
		assert.True(t, ok)
		assert.Equal(t, "email", field)
	})

	t.Run("duplicate username is rejected with ErrDuplicate and field=username", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithUsername("shared_name"), WithEmail("first@test.com"))

		_, err := st.UserCreate(ctx, &models.User{
			UserData: models.UserData{
				Name:     "second user",
				Username: "shared_name",
				Email:    "second@test.com",
			},
			Password: models.UserPassword{Hash: "somehash"},
		})
		require.ErrorIs(t, err, store.ErrDuplicate)

		field, ok := store.DuplicatedField(err)
		assert.True(t, ok)
		assert.Equal(t, "username", field)
	})
}

// TestUserUpdate checks that UserUpdate writes each field changes sets and nothing else, that an
// empty recovery email, external ID or last login clears the column, that a username or email
// another user holds is named in the duplicate error, and that an unknown user yields
// ErrNoDocuments.
func (s *Suite) TestUserUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		name := "New Value"
		err := st.UserUpdate(ctx, "00000000-0000-0000-0000-000000000000", &models.UserChanges{Name: &name})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("writes every field it sets", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("john_doe"), WithEmail("john.doe@test.com"))

		name, username, email, recoveryEmail := "John", "johnny", "john@new.test.com", "john@recovery.test.com"
		hash, externalID := "newhash", "saml-subject"
		status := models.UserStatusNotConfirmed
		admin, awaitingApproval, maxNamespaces := true, true, 7
		methods := []models.UserAuthMethod{models.UserAuthMethodLocal, models.UserAuthMethodSAML}
		lastLogin := s.Clock().Now().UTC().Add(time.Hour).Truncate(time.Second)

		require.NoError(t, st.UserUpdate(ctx, userID, &models.UserChanges{
			Name:             &name,
			Username:         &username,
			Email:            &email,
			RecoveryEmail:    &recoveryEmail,
			PasswordHash:     &hash,
			Status:           &status,
			Admin:            &admin,
			AwaitingApproval: &awaitingApproval,
			MaxNamespaces:    &maxNamespaces,
			AuthMethods:      &methods,
			ExternalID:       &externalID,
			LastLogin:        &lastLogin,
		}))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, name, updated.Name)
		assert.Equal(t, username, updated.Username)
		assert.Equal(t, email, updated.Email)
		assert.Equal(t, recoveryEmail, updated.RecoveryEmail)
		assert.Equal(t, hash, updated.Password.Hash)
		assert.Equal(t, status, updated.Status)
		assert.True(t, updated.Admin)
		assert.True(t, updated.AwaitingApproval)
		assert.Equal(t, maxNamespaces, updated.MaxNamespaces)
		assert.Equal(t, methods, updated.Preferences.AuthMethods)
		assert.Equal(t, externalID, updated.ExternalID)
		assert.WithinDuration(t, lastLogin, updated.LastLogin, time.Second)
	})

	t.Run("leaves every field it does not set", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("john_doe"), WithEmail("john.doe@test.com"), WithExternalID("subject"))

		before, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)

		name := "John"
		require.NoError(t, st.UserUpdate(ctx, userID, &models.UserChanges{Name: &name}))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)

		expected := *before
		expected.Name = name
		assert.Equal(t, &expected, updated)
	})

	t.Run("clears the recovery email, external ID and last login when empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithExternalID("subject"))

		recoveryEmail := "recovery@test.com"
		require.NoError(t, st.UserUpdate(ctx, userID, &models.UserChanges{RecoveryEmail: &recoveryEmail}))

		empty, zero := "", time.Time{}
		require.NoError(t, st.UserUpdate(ctx, userID, &models.UserChanges{RecoveryEmail: &empty, ExternalID: &empty, LastLogin: &zero}))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Empty(t, updated.RecoveryEmail)
		assert.Empty(t, updated.ExternalID)
		assert.True(t, updated.LastLogin.IsZero())
	})

	t.Run("names the email when another user holds it", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithUsername("jane_doe"), WithEmail("jane.doe@test.com"))
		userID := s.CreateUser(t, WithUsername("john_doe"), WithEmail("john.doe@test.com"))

		email := "jane.doe@test.com"
		err := st.UserUpdate(ctx, userID, &models.UserChanges{Email: &email})
		require.ErrorIs(t, err, store.ErrDuplicate)

		field, ok := store.DuplicatedField(err)
		assert.True(t, ok)
		assert.Equal(t, "email", field)
	})

	t.Run("names the username when another user holds it", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateUser(t, WithUsername("jane_doe"), WithEmail("jane.doe@test.com"))
		userID := s.CreateUser(t, WithUsername("john_doe"), WithEmail("john.doe@test.com"))

		username := "jane_doe"
		err := st.UserUpdate(ctx, userID, &models.UserChanges{Username: &username})
		require.ErrorIs(t, err, store.ErrDuplicate)

		field, ok := store.DuplicatedField(err)
		assert.True(t, ok)
		assert.Equal(t, "username", field)
	})
}

// TestUserReplacePasswordHash checks that the hash is replaced only while it is still the one the
// caller read, and that neither a moved-on hash nor an unknown user is an error.
func (s *Suite) TestUserReplacePasswordHash(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("replaces the hash it was given", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)

		require.NoError(t, st.UserReplacePasswordHash(ctx, userID, "hashedpassword", "$upgraded"))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, "$upgraded", updated.Password.Hash)
		assert.Equal(t, 0, updated.TokenVersion)
	})

	t.Run("keeps a hash that changed since it was read", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)

		reset := "$reset"
		require.NoError(t, st.UserUpdate(ctx, userID, &models.UserChanges{PasswordHash: &reset}))

		require.NoError(t, st.UserReplacePasswordHash(ctx, userID, "hashedpassword", "$upgraded"))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, reset, updated.Password.Hash)
	})

	t.Run("ignores an unknown user", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		assert.NoError(t, st.UserReplacePasswordHash(ctx, "00000000-0000-0000-0000-000000000000", "old", "new"))
	})
}

// TestUserRevokeTokens checks that each revocation advances the user's token version by one, and
// that revoking an unknown user's tokens yields ErrNoDocuments.
func (s *Suite) TestUserRevokeTokens(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.UserRevokeTokens(ctx, "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("advances the token version on each revocation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)

		user, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		require.Equal(t, 0, user.TokenVersion)

		require.NoError(t, st.UserRevokeTokens(ctx, userID))
		require.NoError(t, st.UserRevokeTokens(ctx, userID))

		user, err = st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, 2, user.TokenVersion)
	})
}

// TestUserUpdateRevokingTokens checks that UserUpdateRevokingTokens writes the changes and advances their token
// version together, and that an unknown user yields ErrNoDocuments.
func (s *Suite) TestUserUpdateRevokingTokens(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		hash := "newhash"
		err := st.UserUpdateRevokingTokens(ctx, "00000000-0000-0000-0000-000000000000", &models.UserChanges{PasswordHash: &hash})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("saves the user and advances the token version", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)

		hash := "newhash"
		require.NoError(t, st.UserUpdateRevokingTokens(ctx, userID, &models.UserChanges{PasswordHash: &hash}))

		updated, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, "newhash", updated.Password.Hash)
		assert.Equal(t, 1, updated.TokenVersion)
	})
}

// TestUserDelete exercises UserDelete against the store under test.
func (s *Suite) TestUserDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		nonExistentID := "00000000-0000-0000-0000-000000000000"

		err := st.UserDelete(ctx, &models.User{ID: nonExistentID})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when user is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("john_doe"))

		err := st.UserDelete(ctx, &models.User{ID: userID})
		require.NoError(t, err)
	})
}

// TestUserGetInfo tests getting user namespace information
func (s *Suite) TestUserGetInfo(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("returns empty lists for user with no namespaces", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("lonelyuser"))

		userInfo, err := st.UserGetInfo(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.Empty(t, userInfo.OwnedNamespaces)
		assert.Empty(t, userInfo.AssociatedNamespaces)
	})

	t.Run("returns owned namespaces for user who owns namespaces", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t, WithUsername("owner"))

		tenant1 := s.CreateNamespace(t, WithNamespaceName("ns1"), WithOwner(userID))
		tenant2 := s.CreateNamespace(t, WithNamespaceName("ns2"), WithOwner(userID))

		s.CreateMembership(t, tenant1, userID, "owner")
		s.CreateMembership(t, tenant2, userID, "owner")

		userInfo, err := st.UserGetInfo(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.Len(t, userInfo.OwnedNamespaces, 2)
		assert.Empty(t, userInfo.AssociatedNamespaces)
	})

	t.Run("returns associated namespaces for member users", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		memberID := s.CreateUser(t, WithUsername("member"))

		tenantID := s.CreateNamespace(t, WithNamespaceName("shared-ns"))

		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), &models.Member{
			ID:   memberID,
			Role: "observer",
		})
		require.NoError(t, err)

		userInfo, err := st.UserGetInfo(ctx, memberID)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.NotNil(t, userInfo.OwnedNamespaces)
		assert.NotNil(t, userInfo.AssociatedNamespaces)
	})

	t.Run("separates owned and associated namespaces correctly", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		user1ID := s.CreateUser(t, WithUsername("user1"))

		s.CreateNamespace(t, WithNamespaceName("ns1"))
		ns2ID := s.CreateNamespace(t, WithNamespaceName("ns2"))

		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(ns2ID), &models.Member{
			ID:   user1ID,
			Role: "observer",
		})
		require.NoError(t, err)

		userInfo, err := st.UserGetInfo(ctx, user1ID)
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.NotNil(t, userInfo.OwnedNamespaces)
		assert.NotNil(t, userInfo.AssociatedNamespaces)
	})

	t.Run("returns info for non-existent user without error", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userInfo, err := st.UserGetInfo(ctx, "00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		require.NotNil(t, userInfo)

		assert.Empty(t, userInfo.OwnedNamespaces)
		assert.Empty(t, userInfo.AssociatedNamespaces)
	})
}
