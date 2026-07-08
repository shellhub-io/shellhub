package services

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
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
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedFound: false,
			expectedErr:   false,
		},
		{
			description: "resolves the identity and stamps last-used on a hit",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
					Return(&models.SSHIdentity{ID: "id1", UserID: "user1", TenantID: tenantID, Fingerprint: fingerprint}, nil).Once()
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

			service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

			identity, found, err := service.ResolveSSHIdentity(ctx, tenantID, fingerprint)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedFound, found)
				if tc.expectedFound {
					require.Equal(t, tc.expectedUser, identity.UserID)
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
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
					return identity.UserID == userID && identity.Fingerprint == fingerprint && identity.TenantID == tenantID
				})).Return("id1", nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "is idempotent when the same account already holds the key",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
					Return(&models.SSHIdentity{ID: "id1", UserID: userID, Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects when the fingerprint is bound to another identity",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
					Return(&models.SSHIdentity{ID: "id2", UserID: "other", Fingerprint: fingerprint}, nil).Once()
			},
			expectedErr: NewErrSSHIdentityDuplicated(fingerprint, nil),
		},
		{
			description: "maps a store uniqueness violation to a duplicated error",
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
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

			service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

			err := service.EnrollSSHIdentity(ctx, userID, tenantID, fingerprint, []byte("ssh-ed25519 AAAA"), "")
			require.Equal(t, tc.expectedErr, err)

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
		service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

		_, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{TenantID: tenantID, UserID: userID, Data: "not-a-key"})
		require.ErrorContains(t, err, "ssh identity public key invalid")

		storeMock.AssertExpectations(t)
	})

	t.Run("enrolls a valid pasted public key with the SHA256 fingerprint", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
		storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityFingerprintResolver, fingerprint, mock.Anything).
			Return(nil, store.ErrNoDocuments).Once()
		storeMock.On("SSHIdentityCreate", ctx, mock.MatchedBy(func(identity *models.SSHIdentity) bool {
			return identity.Fingerprint == fingerprint && identity.UserID == userID
		})).Return("id1", nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

		identity, err := service.CreateSSHIdentity(ctx, &requests.SSHIdentityCreate{TenantID: tenantID, UserID: userID, Name: "laptop", Data: authorized})
		require.NoError(t, err)
		require.Equal(t, fingerprint, identity.Fingerprint)

		storeMock.AssertExpectations(t)
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
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityIDResolver, idOwn, mock.Anything).
					Return(&models.SSHIdentity{ID: idOwn, UserID: userID, TenantID: tenantID}, nil).Once()
				storeMock.On("SSHIdentityDelete", ctx, mock.Anything).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "forbids deleting another member's identity without manage",
			req:         &requests.SSHIdentityDelete{SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: idOther}, TenantID: tenantID, UserID: userID, Manage: false},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityIDResolver, idOther, mock.Anything).
					Return(&models.SSHIdentity{ID: idOther, UserID: "someone-else", TenantID: tenantID}, nil).Once()
			},
			expectedErr: NewErrForbidden(ErrForbidden, nil),
		},
		{
			description: "allows deleting another member's identity with manage",
			req:         &requests.SSHIdentityDelete{SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: idOther}, TenantID: tenantID, UserID: userID, Manage: true},
			requireMocks: func(storeMock *storemock.MockStore, queryOptionsMock *storemock.MockQueryOptions) {
				queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
				storeMock.On("SSHIdentityResolve", ctx, store.SSHIdentityIDResolver, idOther, mock.Anything).
					Return(&models.SSHIdentity{ID: idOther, UserID: "someone-else", TenantID: tenantID}, nil).Once()
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

			service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

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
		queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
		queryOptionsMock.On("WithUserID", userID).Return(nil).Once()
		storeMock.On("SSHIdentityList", ctx, mock.Anything, mock.Anything).
			Return([]models.SSHIdentity{{ID: "id1", UserID: userID}}, 1, nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

		list, err := service.ListSSHIdentities(ctx, &requests.SSHIdentityList{TenantID: tenantID, UserID: userID, All: false})
		require.NoError(t, err)
		require.Len(t, list, 1)

		storeMock.AssertExpectations(t)
	})

	t.Run("lists every member when all is set", func(t *testing.T) {
		storeMock := new(storemock.MockStore)
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		queryOptionsMock.On("InNamespace", tenantID).Return(nil).Once()
		storeMock.On("SSHIdentityList", ctx, mock.Anything).
			Return([]models.SSHIdentity{{ID: "id1", UserID: userID}, {ID: "id2", UserID: "user2"}}, 2, nil).Once()

		service := NewService(storeMock, privateKey, publicKey, nil, clientMock)

		list, err := service.ListSSHIdentities(ctx, &requests.SSHIdentityList{TenantID: tenantID, UserID: userID, All: true})
		require.NoError(t, err)
		require.Len(t, list, 2)

		// WithUserID must not be applied for the namespace-wide view.
		queryOptionsMock.AssertNotCalled(t, "WithUserID", mock.Anything)

		storeMock.AssertExpectations(t)
	})
}
