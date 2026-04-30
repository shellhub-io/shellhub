package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// pkceChallenge computes S256 PKCE challenge for a given verifier (test helper).
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestOAuthRegisterClient(t *testing.T) {
	type Expected struct {
		err error
	}

	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		tenantID      string
		name          string
		redirectURIs  []string
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description:  "fails when store returns error",
			tenantID:     "tenant-1",
			name:         "my-client",
			redirectURIs: []string{"http://localhost/callback"},
			requiredMocks: func(ctx context.Context) {
				uuidMock := &uuidmock.Uuid{}
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return("plain-secret-uuid").Once()
				uuidMock.On("Generate").Return("client-id-uuid").Once()
				uuidMock.On("Generate").Return("id-uuid").Once()

				storeMock.
					On("OAuthClientCreate", ctx, mock.AnythingOfType("*models.OAuthClient")).
					Return("", errors.New("store error")).
					Once()
			},
			expected: Expected{err: errors.New("store error")},
		},
		{
			description:  "succeeds and returns plain secret",
			tenantID:     "tenant-1",
			name:         "my-client",
			redirectURIs: []string{"http://localhost/callback"},
			requiredMocks: func(ctx context.Context) {
				uuidMock := &uuidmock.Uuid{}
				uuid.DefaultBackend = uuidMock
				// uuid.Generate called 3x: plainSecret, client.ID, client.ClientID
				uuidMock.On("Generate").Return("plain-secret-uuid").Once()
				uuidMock.On("Generate").Return("client-id-uuid").Once()
				uuidMock.On("Generate").Return("id-uuid").Once()

				// The persisted client must hold the SHA-256 hex hash of the
				// plain secret, never the plain secret itself — otherwise the
				// secret would be recoverable from the database.
				expectedHash := sha256.Sum256([]byte("plain-secret-uuid"))
				expectedHex := hex.EncodeToString(expectedHash[:])

				storeMock.
					On("OAuthClientCreate", ctx, mock.MatchedBy(func(c *models.OAuthClient) bool {
						return c != nil &&
							c.ClientSecret == expectedHex &&
							c.ClientSecret != "plain-secret-uuid" &&
							c.TenantID == "tenant-1" &&
							c.Name == "my-client"
					})).
					Return("id-1", nil).
					Once()
			},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)
			client, plainSecret, err := svc.OAuthRegisterClient(ctx, tc.tenantID, tc.name, tc.redirectURIs)

			if tc.expected.err != nil {
				require.Error(t, err)
				require.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				require.NotEmpty(t, plainSecret)
				require.Equal(t, plainSecret, client.ClientSecret)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestOAuthGetClient(t *testing.T) {
	type Expected struct {
		clientID string
		err      bool
	}

	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		clientID      string
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when client not found",
			clientID:    "unknown",
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("OAuthClientResolve", ctx, mock.Anything, "unknown").
					Return(nil, errors.New("not found")).
					Once()
			},
			expected: Expected{err: true},
		},
		{
			description: "succeeds",
			clientID:    "client-id-1",
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("OAuthClientResolve", ctx, mock.Anything, "client-id-1").
					Return(&models.OAuthClient{ClientID: "client-id-1"}, nil).
					Once()
			},
			expected: Expected{clientID: "client-id-1", err: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)
			client, err := svc.OAuthGetClient(ctx, tc.clientID)

			if tc.expected.err {
				require.Error(t, err)
				require.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected.clientID, client.ClientID)
			}

			storeMock.AssertExpectations(t)
		})
	}
}

func TestOAuthDeleteClient(t *testing.T) {
	storeMock := new(storemock.Store)

	t.Run("fails when client not found", func(t *testing.T) {
		ctx := context.Background()
		storeMock.
			On("OAuthClientResolve", ctx, mock.Anything, "id-1").
			Return(nil, errors.New("not found")).
			Once()

		svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)
		err := svc.OAuthDeleteClient(ctx, "id-1", "tenant-1")
		require.Error(t, err)
		storeMock.AssertExpectations(t)
	})

	t.Run("succeeds when client tenant matches", func(t *testing.T) {
		ctx := context.Background()
		c := &models.OAuthClient{ID: "id-1", ClientID: "client-id-1", TenantID: "tenant-1"}
		storeMock.
			On("OAuthClientResolve", ctx, mock.Anything, "id-1").
			Return(c, nil).
			Once()
		storeMock.
			On("OAuthClientDelete", ctx, c).
			Return(nil).
			Once()

		svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)
		err := svc.OAuthDeleteClient(ctx, "id-1", "tenant-1")
		require.NoError(t, err)
		storeMock.AssertExpectations(t)
	})

	t.Run("returns not found when client belongs to a different tenant", func(t *testing.T) {
		ctx := context.Background()
		c := &models.OAuthClient{ID: "id-1", ClientID: "client-id-1", TenantID: "tenant-other"}
		storeMock.
			On("OAuthClientResolve", ctx, mock.Anything, "id-1").
			Return(c, nil).
			Once()

		svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)
		err := svc.OAuthDeleteClient(ctx, "id-1", "tenant-1")
		require.Error(t, err)
		storeMock.AssertExpectations(t)
	})
}

func TestOAuthCreateAuthCode(t *testing.T) {
	t.Run("persists the code in the cache with the right key, payload and TTL", func(t *testing.T) {
		ctx := context.Background()
		storeMock := new(storemock.Store)
		cm := new(cachemock.Cache)

		var captured string
		cm.
			On("Set",
				ctx,
				mock.MatchedBy(func(key string) bool {
					ok := len(key) == len(oauthCodeCachePrefix)+64 &&
						key[:len(oauthCodeCachePrefix)] == oauthCodeCachePrefix
					if ok {
						captured = key[len(oauthCodeCachePrefix):]
					}

					return ok
				}),
				mock.MatchedBy(func(v *models.OAuthAuthCode) bool {
					return v != nil &&
						v.Code == captured &&
						v.ClientID == "client-id" &&
						v.UserID == "user-id" &&
						v.TenantID == "tenant-id" &&
						v.RedirectURI == "http://localhost/cb" &&
						v.CodeChallenge == "challenge"
				}),
				oauthCodeTTL,
			).
			Return(nil).
			Once()

		svc := NewService(storeMock, privateKey, publicKey, cm, clientMock)
		code, err := svc.OAuthCreateAuthCode(ctx, "client-id", "user-id", "tenant-id", "http://localhost/cb", "challenge")
		require.NoError(t, err)
		require.Equal(t, captured, code, "returned code must match the one stored in cache")
		require.Len(t, code, 64)
		cm.AssertExpectations(t)
	})

	t.Run("generates unique codes across calls", func(t *testing.T) {
		ctx := context.Background()
		storeMock := new(storemock.Store)
		svc := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

		code1, err := svc.OAuthCreateAuthCode(ctx, "client-id", "user-id", "tenant-id", "http://localhost/cb", "challenge")
		require.NoError(t, err)

		code2, err := svc.OAuthCreateAuthCode(ctx, "client-id", "user-id", "tenant-id", "http://localhost/cb", "challenge")
		require.NoError(t, err)
		require.NotEqual(t, code1, code2)
	})
}

func TestVerifyPKCE(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkceChallenge(verifier)

	require.True(t, verifyPKCE(verifier, challenge))
	require.False(t, verifyPKCE(verifier, "wrong-challenge"))
	require.False(t, verifyPKCE("wrong-verifier", challenge))

	// base64url is case-sensitive (RFC 4648); a case-only difference must NOT
	// pass — matching behavior in PKCE per RFC 7636 §4.6.
	require.False(t, verifyPKCE(verifier, swapCase(challenge)))
}

// swapCase returns s with ASCII letter case flipped.
func swapCase(s string) string {
	out := make([]byte, len(s))
	for i, c := range []byte(s) {
		switch {
		case c >= 'a' && c <= 'z':
			out[i] = c - 32
		case c >= 'A' && c <= 'Z':
			out[i] = c + 32
		default:
			out[i] = c
		}
	}

	return string(out)
}

func TestOAuthExchangeCode_Validation(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkceChallenge(verifier)

	cases := []struct {
		description   string
		stored        models.OAuthAuthCode // value returned by cache.GetDelete
		cacheErr      error                // error returned by cache.GetDelete
		clientID      string               // client_id sent at /oauth/token
		redirectURI   string               // redirect_uri sent at /oauth/token
		codeVerifier  string               // code_verifier sent at /oauth/token
		expectedError error
	}{
		{
			description:   "fails when cache returns error (key absent or already consumed)",
			cacheErr:      errors.New("not found"),
			clientID:      "c",
			redirectURI:   "https://app/cb",
			codeVerifier:  verifier,
			expectedError: ErrOAuthCodeInvalid,
		},
		{
			description:   "fails when stored code is empty (cache miss)",
			stored:        models.OAuthAuthCode{Code: ""},
			clientID:      "c",
			redirectURI:   "https://app/cb",
			codeVerifier:  verifier,
			expectedError: ErrOAuthCodeInvalid,
		},
		{
			description:   "fails when client_id does not match",
			stored:        models.OAuthAuthCode{Code: "abc", ClientID: "real-client", RedirectURI: "https://app/cb", CodeChallenge: challenge},
			clientID:      "different-client",
			redirectURI:   "https://app/cb",
			codeVerifier:  verifier,
			expectedError: ErrOAuthClientIDMismatch,
		},
		{
			description:   "fails when redirect_uri does not match",
			stored:        models.OAuthAuthCode{Code: "abc", ClientID: "c", RedirectURI: "https://app/cb", CodeChallenge: challenge},
			clientID:      "c",
			redirectURI:   "https://attacker/cb",
			codeVerifier:  verifier,
			expectedError: ErrOAuthRedirectURIMismatch,
		},
		{
			description:   "fails when PKCE verifier does not match the challenge",
			stored:        models.OAuthAuthCode{Code: "abc", ClientID: "c", RedirectURI: "https://app/cb", CodeChallenge: challenge},
			clientID:      "c",
			redirectURI:   "https://app/cb",
			codeVerifier:  "tampered-verifier",
			expectedError: ErrOAuthPKCEInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			cm := new(cachemock.Cache)
			storeMock := new(storemock.Store)

			cm.On("GetDelete", ctx, oauthCodeCachePrefix+"abc", mock.AnythingOfType("*models.OAuthAuthCode")).
				Run(func(args mock.Arguments) {
					if tc.cacheErr != nil {
						return
					}
					out := args.Get(2).(*models.OAuthAuthCode)
					*out = tc.stored
				}).
				Return(tc.cacheErr).
				Once()

			svc := NewService(storeMock, privateKey, publicKey, cm, clientMock)
			_, err := svc.OAuthExchangeCode(ctx, "abc", tc.clientID, tc.redirectURI, tc.codeVerifier)
			require.Error(t, err)
			require.ErrorIs(t, err, tc.expectedError)
			cm.AssertExpectations(t)
		})
	}
}

func TestOAuthExchangeCode_Success(t *testing.T) {
	// Other tests in this file install global uuid mocks via uuid.DefaultBackend.
	// EncodeUserClaims internally calls uuid.Generate, so install a real
	// generator for the duration of this test.
	uuidMock := &uuidmock.Uuid{}
	uuidMock.On("Generate").Return("token-jti-1").Maybe()
	uuid.DefaultBackend = uuidMock

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkceChallenge(verifier)

	ctx := context.Background()
	cm := new(cachemock.Cache)
	storeMock := new(storemock.Store)

	tenant := "00000000-0000-4000-0000-000000000000"
	userID := "111111111111111111111111"

	authCode := models.OAuthAuthCode{
		Code:          "the-code",
		ClientID:      "client-id",
		UserID:        userID,
		TenantID:      tenant,
		RedirectURI:   "https://app/cb",
		CodeChallenge: challenge,
	}

	// GetDelete is the only cache op needed to read the auth code — it
	// atomically returns the value AND removes the key, so a second concurrent
	// request would get ErrGetNotFound and be rejected.
	cm.On("GetDelete", ctx, oauthCodeCachePrefix+"the-code", mock.AnythingOfType("*models.OAuthAuthCode")).
		Run(func(args mock.Arguments) {
			out := args.Get(2).(*models.OAuthAuthCode)
			*out = authCode
		}).
		Return(nil).
		Once()

	user := &models.User{
		ID:        userID,
		Status:    models.UserStatusConfirmed,
		LastLogin: now,
		UserData:  models.UserData{Username: "alice", Email: "a@b.c", Name: "Alice"},
		Preferences: models.UserPreferences{
			PreferredNamespace: tenant, // matches → no UserUpdate call
		},
	}
	storeMock.On("UserResolve", ctx, store.UserIDResolver, userID).Return(user, nil).Once()

	namespace := &models.Namespace{
		Name:     "ns",
		TenantID: tenant,
		Members: []models.Member{
			{ID: userID, Role: authorizer.RoleObserver},
		},
	}
	storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(namespace, nil).Once()

	// CreateUserToken caches the issued JWT — accept the Set without
	// asserting; it returns nil so the flow continues.
	cm.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := NewService(storeMock, privateKey, publicKey, cm, clientMock)
	resp, err := svc.OAuthExchangeCode(ctx, "the-code", "client-id", "https://app/cb", verifier)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Token, "issued token must be non-empty")
	require.Equal(t, userID, resp.ID, "token must be issued for the auth code's user")
	require.Equal(t, tenant, resp.Tenant, "token must be scoped to the auth code's tenant")

	cm.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

// Per RFC 6749 §4.1.2, authorization codes are single-use. The atomic
// GetDelete consumption means even if token issuance fails afterwards, the
// code is already gone — the client must restart the flow with a new code.
// This test pins the consume-before-validate ordering so a future refactor
// can't accidentally re-introduce a window where the code stays alive after
// being read.
func TestOAuthExchangeCode_ConsumesCodeBeforeTokenIssuance(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkceChallenge(verifier)

	ctx := context.Background()
	cm := new(cachemock.Cache)
	storeMock := new(storemock.Store)

	authCode := models.OAuthAuthCode{
		Code:          "the-code",
		ClientID:      "client-id",
		UserID:        "deleted-user",
		TenantID:      "tenant",
		RedirectURI:   "https://app/cb",
		CodeChallenge: challenge,
	}

	cm.On("GetDelete", ctx, oauthCodeCachePrefix+"the-code", mock.AnythingOfType("*models.OAuthAuthCode")).
		Run(func(args mock.Arguments) {
			out := args.Get(2).(*models.OAuthAuthCode)
			*out = authCode
		}).
		Return(nil).
		Once()

	// Token issuance fails — but GetDelete already consumed the code.
	storeMock.On("UserResolve", ctx, store.UserIDResolver, "deleted-user").
		Return(nil, store.ErrNoDocuments).Once()

	svc := NewService(storeMock, privateKey, publicKey, cm, clientMock)
	_, err := svc.OAuthExchangeCode(ctx, "the-code", "client-id", "https://app/cb", verifier)
	require.Error(t, err)

	// GetDelete was called exactly once (as expected). No separate Delete is
	// expected because consumption is atomic with the read.
	cm.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}
