package services

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// newTestPublicKey returns a fresh ed25519 OpenSSH authorized key and its SHA256
// fingerprint so tests exercise the same parse/fingerprint path as production.
func newTestPublicKey(t *testing.T) (authorized string, fingerprint string) {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := ssh.NewSignerFromKey(priv)
	require.NoError(t, err)

	pub := signer.PublicKey()

	return string(ssh.MarshalAuthorizedKey(pub)), ssh.FingerprintSHA256(pub)
}

func TestResolveSSHIdentity(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		fingerprint = "SHA256:abc"
	)

	cases := []struct {
		description   string
		requireMocks  func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedFound bool
		expectedUser  string
		expectedErr   bool
	}{
		{
			description: "returns not found when the fingerprint is not enrolled",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedFound: false,
			expectedErr:   false,
		},
		{
			description: "resolves the identity and stamps last-used on a hit",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{ID: "id1", PrincipalID: "user1", TenantID: tenantID, Fingerprint: fingerprint}, nil).Once()
				storeMock.On("SSHIdentityTouchLastUsed", ctx, tenantID, fingerprint).
					Return(nil).Once()
			},
			expectedFound: true,
			expectedUser:  "user1",
			expectedErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			identity, found, err := service.ResolveSSHIdentity(ctx, tenantID, fingerprint)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedFound, found)
				if tc.expectedFound {
					require.Equal(t, tc.expectedUser, identity.PrincipalID)
				}
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestEnrollSSHIdentity(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		fingerprint = "SHA256:abc"
		userID      = "user1"
	)

	cases := []struct {
		description  string
		requireMocks func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedErr  error
	}{
		{
			description: "creates the binding when the fingerprint is free",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
					return identity.PrincipalID == userID && identity.Fingerprint == fingerprint && identity.TenantID == tenantID
				})).Return("id1", nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects when the same account already holds the key",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{ID: "id1", PrincipalID: userID, Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: NewErrSSHIdentityDuplicated(fingerprint, nil),
		},
		{
			description: "rejects when the fingerprint is bound to another identity",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{ID: "id2", PrincipalID: "other", Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: NewErrSSHIdentityDuplicated(fingerprint, nil),
		},
		{
			description: "maps a store uniqueness violation to a duplicated error",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("SSHIdentityCreate", ctx, mock.Anything).
					Return("", store.ErrDuplicate).Once()
			},
			expectedErr: NewErrSSHIdentityDuplicated(fingerprint, store.ErrDuplicate),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			_, err := service.enrollSSHIdentity(ctx, &models.SSHIdentity{
				TenantID:    tenantID,
				PrincipalID: userID,
				Fingerprint: fingerprint,
				Data:        []byte("ssh-ed25519 AAAA"),
				Source:      models.SSHIdentitySourceApproval,
			})
			require.Equal(t, tc.expectedErr, err)

			storeMock.AssertExpectations(t)
		})
	}
}

func TestReenrollSSHIdentity(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID    = "00000000-0000-4000-0000-000000000000"
		fingerprint = "SHA256:abc"
		userID      = "user1"
	)

	cases := []struct {
		description  string
		requireMocks func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedID   string
		expectedErr  error
	}{
		{
			description: "returns the existing binding when the same account already holds the key",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{ID: "existing", PrincipalID: userID, Fingerprint: fingerprint}, nil).Once()
			},
			expectedID:  "existing",
			expectedErr: nil,
		},
		{
			description: "creates the binding when the fingerprint is free",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("SSHIdentityCreate", ctx, mock.Anything).Return("created", nil).Once()
			},
			expectedID:  "created",
			expectedErr: nil,
		},
		{
			description: "rejects when the fingerprint is bound to another identity",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
					Return(&models.SSHIdentity{ID: "id2", PrincipalID: "other", Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: NewErrSSHIdentityDuplicated(fingerprint, nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			identity, err := service.reenrollSSHIdentity(ctx, &models.SSHIdentity{
				TenantID:    tenantID,
				PrincipalID: userID,
				Fingerprint: fingerprint,
				Data:        []byte("ssh-ed25519 AAAA"),
				Source:      models.SSHIdentitySourceApproval,
			})
			require.Equal(t, tc.expectedErr, err)

			if tc.expectedErr == nil {
				require.NotNil(t, identity)
				require.Equal(t, tc.expectedID, identity.ID)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestCreateSSHIdentity(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "user1"
	)

	authorized, fingerprint := newTestPublicKey(t)

	t.Run("rejects an unparseable public key", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		service := NewService(storeMock, privateKey, publicKey, nil)

		_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{TenantID: tenantID, UserID: userID, Data: "not-a-key"})
		require.ErrorContains(t, err, "ssh identity public key invalid")

		storeMock.AssertExpectations(t)
	})

	t.Run("enrolls a valid pasted public key with the SHA256 fingerprint", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(nil, store.ErrNoDocuments).Once()
		storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			return identity.Fingerprint == fingerprint && identity.PrincipalID == userID
		})).Return("id1", nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil)

		identity, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized})
		require.NoError(t, err)
		require.Equal(t, fingerprint, identity.Fingerprint)

		storeMock.AssertExpectations(t)
	})

	t.Run("reports a duplicate when the caller already enrolled the key", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(&models.SSHIdentity{ID: "id1", PrincipalID: userID, Fingerprint: fingerprint}, nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil)

		_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized})
		require.Equal(t, NewErrSSHIdentityDuplicated(fingerprint, nil), err)

		storeMock.AssertExpectations(t)
	})
}

// The source is decided per enrollment path and written once, at creation, so
// each path is pinned to what it claims.
func TestSSHIdentitySourceIsRecordedPerPath(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "user1"
	)

	authorized, fingerprint := newTestPublicKey(t)

	// enroll runs the given path and reports the source that reached the store.
	enroll := func(t *testing.T, run func(service *APIService) error) models.SSHIdentitySource {
		t.Helper()

		var got models.SSHIdentitySource

		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(nil, store.ErrNoDocuments).Once()
		storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			got = identity.Source

			return true
		})).Return("id1", nil).Once()

		require.NoError(t, run(NewService(storeMock, privateKey, publicKey, nil)))
		storeMock.AssertExpectations(t)

		return got
	}

	t.Run("a key accepted at login is recorded as an approval", func(t *testing.T) {
		source := enroll(t, func(service *APIService) error {
			_, err := service.enrollSSHIdentity(ctx, &models.SSHIdentity{
				TenantID:    tenantID,
				PrincipalID: userID,
				Fingerprint: fingerprint,
				Data:        []byte(authorized),
				Name:        "laptop",
				Source:      models.SSHIdentitySourceApproval,
			})

			return err
		})
		require.Equal(t, models.SSHIdentitySourceApproval, source)
	})

	t.Run("the web terminal's own key is recorded as a browser key", func(t *testing.T) {
		source := enroll(t, func(service *APIService) error {
			_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{
				TenantID: tenantID, UserID: userID, Name: "chrome", Data: authorized,
				Source: models.SSHIdentitySourceBrowser,
			})

			return err
		})
		require.Equal(t, models.SSHIdentitySourceBrowser, source)
	})

	// A TTL only matters if it survives the trip to the row: the gateway reads
	// expires_at and nothing else to decide a key is dead.
	t.Run("a TTL in days becomes an absolute expiry", func(t *testing.T) {
		var got *time.Time

		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(nil, store.ErrNoDocuments).Once()
		storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			got = identity.ExpiresAt

			return true
		})).Return("id1", nil).Once()

		clockMock.On("Now").Return(now)
		service := NewService(storeMock, privateKey, publicKey, nil)

		days := 30
		_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{
			TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized, ExpiresIn: &days,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, now.AddDate(0, 0, 30), *got)

		storeMock.AssertExpectations(t)
	})

	t.Run("no TTL leaves the key durable", func(t *testing.T) {
		var got *time.Time

		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityFingerprintResolver, fingerprint).
			Return(nil, store.ErrNoDocuments).Once()
		storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			got = identity.ExpiresAt

			return true
		})).Return("id1", nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil)

		_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{
			TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized,
		})
		require.NoError(t, err)
		require.Nil(t, got)

		storeMock.AssertExpectations(t)
	})

	t.Run("a request claiming no source falls back to a pasted key", func(t *testing.T) {
		source := enroll(t, func(service *APIService) error {
			_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{
				TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized,
			})

			return err
		})
		require.Equal(t, models.SSHIdentitySourceManual, source)
	})
}

func TestDeleteSSHIdentity(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "user1"
		idOwn    = "id-own"
		idOther  = "id-other"
	)

	cases := []struct {
		description  string
		req          *requests.SSHIdentityDelete
		requireMocks func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions)
		expectedErr  error
	}{
		{
			description: "deletes the caller's own identity",
			req:         &requests.SSHIdentityDelete{SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: idOwn}, TenantID: tenantID, UserID: userID, Manage: false},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityIDResolver, idOwn).
					Return(&models.SSHIdentity{ID: idOwn, PrincipalID: userID, TenantID: tenantID}, nil).Once()
				storeMock.On("SSHIdentityDelete", ctx, mock.Anything).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "forbids deleting another member's identity without manage",
			req:         &requests.SSHIdentityDelete{SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: idOther}, TenantID: tenantID, UserID: userID, Manage: false},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityIDResolver, idOther).
					Return(&models.SSHIdentity{ID: idOther, PrincipalID: "someone-else", TenantID: tenantID}, nil).Once()
			},
			expectedErr: NewErrForbidden(ErrForbidden, nil),
		},
		{
			description: "allows deleting another member's identity with manage",
			req:         &requests.SSHIdentityDelete{SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: idOther}, TenantID: tenantID, UserID: userID, Manage: true},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				storeMock.On("SSHIdentityResolve", ctx, mock.Anything, store.SSHIdentityIDResolver, idOther).
					Return(&models.SSHIdentity{ID: idOther, PrincipalID: "someone-else", TenantID: tenantID}, nil).Once()
				storeMock.On("SSHIdentityDelete", ctx, mock.Anything).Return(nil).Once()
			},
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			tc.requireMocks(storeMock, queryOptionsMock)

			service := NewService(storeMock, privateKey, publicKey, nil)

			err := service.DeleteSSHIdentity(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)

			storeMock.AssertExpectations(t)
		})
	}
}

func TestListSSHIdentities(t *testing.T) {
	ctx := context.TODO()

	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		userID   = "user1"
	)

	t.Run("scopes to the caller by default", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		queryOptionsMock.On("WithUserID", userID).Return(nil).Once()
		storeMock.On("SSHIdentityList", ctx, mock.Anything, mock.Anything).
			Return([]models.SSHIdentity{{ID: "id1", PrincipalID: userID}}, 1, nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil)

		list, err := service.ListSSHIdentities(ctx, &requests.SSHIdentityList{TenantID: tenantID, UserID: userID, All: false})
		require.NoError(t, err)
		require.Len(t, list, 1)

		storeMock.AssertExpectations(t)
	})

	t.Run("lists every member when all is set", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		storeMock.On("SSHIdentityList", ctx, mock.Anything).
			Return([]models.SSHIdentity{{ID: "id1", PrincipalID: userID}, {ID: "id2", PrincipalID: "user2"}}, 2, nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil)

		list, err := service.ListSSHIdentities(ctx, &requests.SSHIdentityList{TenantID: tenantID, UserID: userID, All: true})
		require.NoError(t, err)
		require.Len(t, list, 2)

		// WithUserID must not be applied for the namespace-wide view.
		queryOptionsMock.AssertNotCalled(t, "WithUserID", mock.Anything)

		storeMock.AssertExpectations(t)
	})
}
