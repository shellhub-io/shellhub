package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func instanceKeyDigest(plain string) string {
	sum := sha256.Sum256([]byte(plain))

	return hex.EncodeToString(sum[:])
}

func TestCreateInstanceAPIKey(t *testing.T) {
	type Expected struct {
		res *responses.CreateInstanceAPIKey
		err error
	}

	storeMock := storemock.NewMockStore(t)

	cases := []struct {
		description   string
		req           *requests.CreateInstanceAPIKey
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the requesting user does not exist",
			req: &requests.CreateInstanceAPIKey{
				Username:  "ghost",
				Name:      "billing-export",
				ExpiresAt: 30,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserUsernameResolver, "ghost").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrUserNotFound("ghost", errors.New("error")),
			},
		},
		{
			description: "fails when the requesting user is not an instance administrator",
			req: &requests.CreateInstanceAPIKey{
				Username:  "regular",
				Name:      "billing-export",
				ExpiresAt: 30,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserUsernameResolver, "regular").
					Return(&models.User{ID: "user-1", Admin: false}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAuthForbidden(),
			},
		},
		{
			description: "fails when the expiration is outside the permitted set",
			req: &requests.CreateInstanceAPIKey{
				Username:  "admin",
				Name:      "billing-export",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserUsernameResolver, "admin").
					Return(&models.User{ID: "user-1", Admin: true}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrBadRequest(errors.New("expires_at must be one of 30, 60, 90 or 365 days")),
			},
		},
		{
			description: "fails when the name is already taken",
			req: &requests.CreateInstanceAPIKey{
				Username:  "admin",
				Name:      "billing-export",
				ExpiresAt: 30,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserUsernameResolver, "admin").
					Return(&models.User{ID: "user-1", Admin: true}, nil).
					Once()

				uuidMock := new(uuidmock.MockUUID)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return("cdfd3cb0-c44e-4e54-b931-6d57713ad159").Once()

				clockMock.On("Now").Return(now)
				storeMock.
					On("InstanceAPIKeyCreate", ctx, &models.InstanceAPIKey{
						ID:        instanceKeyDigest(models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159"),
						Name:      "billing-export",
						CreatedBy: "user-1",
						ExpiresAt: now.AddDate(0, 0, 30),
					}).
					Return("", store.ErrDuplicate).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrInstanceAPIKeyDuplicated([]string{"name"}),
			},
		},
		{
			description: "succeeds and returns the prefixed plaintext key once",
			req: &requests.CreateInstanceAPIKey{
				Username:  "admin",
				Name:      "license-sync",
				ExpiresAt: 365,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserUsernameResolver, "admin").
					Return(&models.User{ID: "user-1", Admin: true}, nil).
					Once()

				uuidMock := new(uuidmock.MockUUID)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return("cdfd3cb0-c44e-4e54-b931-6d57713ad159").Once()

				clockMock.On("Now").Return(now)
				storeMock.
					On("InstanceAPIKeyCreate", ctx, &models.InstanceAPIKey{
						ID:        instanceKeyDigest(models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159"),
						Name:      "license-sync",
						CreatedBy: "user-1",
						ExpiresAt: now.AddDate(1, 0, 0),
					}).
					Return(instanceKeyDigest(models.InstanceAPIKeyPrefix+"cdfd3cb0-c44e-4e54-b931-6d57713ad159"), nil).
					Once()
			},
			expected: Expected{
				res: &responses.CreateInstanceAPIKey{
					ID:        models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
					Name:      "license-sync",
					CreatedBy: "user-1",
					ExpiresAt: now.AddDate(1, 0, 0),
				},
				err: nil,
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.CreateInstanceAPIKey(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{res, err})
		})
	}

	uuid.DefaultBackend = realUUIDBackend
	storeMock.AssertExpectations(t)
}

func TestAuthInstanceAPIKey(t *testing.T) {
	type Expected struct {
		res *models.InstanceAPIKey
		err error
	}

	storeMock := storemock.NewMockStore(t)

	plain := models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159"
	digest := instanceKeyDigest(plain)

	cases := []struct {
		description   string
		key           string
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the key does not exist",
			key:         plain,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("InstanceAPIKeyResolve", ctx, store.InstanceAPIKeyIDResolver, digest).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrInstanceAPIKeyNotFound("", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when the key has expired",
			key:         plain,
			requiredMocks: func(ctx context.Context) {
				clockMock.On("Now").Return(now)
				storeMock.
					On("InstanceAPIKeyResolve", ctx, store.InstanceAPIKeyIDResolver, digest).
					Return(&models.InstanceAPIKey{
						ID:        digest,
						Name:      "expired",
						CreatedBy: "user-1",
						ExpiresAt: now.Add(-time.Hour),
					}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrInstanceAPIKeyInvalid("expired"),
			},
		},
		{
			description: "fails when the creator is no longer an instance administrator",
			key:         plain,
			requiredMocks: func(ctx context.Context) {
				clockMock.On("Now").Return(now)
				storeMock.
					On("InstanceAPIKeyResolve", ctx, store.InstanceAPIKeyIDResolver, digest).
					Return(&models.InstanceAPIKey{
						ID:        digest,
						Name:      "demoted",
						CreatedBy: "user-1",
						ExpiresAt: now.Add(time.Hour),
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "user-1").
					Return(&models.User{ID: "user-1", Admin: false}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrInstanceAPIKeyInvalid("demoted"),
			},
		},
		{
			description: "succeeds when the key is valid and its creator is still an administrator",
			key:         plain,
			requiredMocks: func(ctx context.Context) {
				clockMock.On("Now").Return(now)
				storeMock.
					On("InstanceAPIKeyResolve", ctx, store.InstanceAPIKeyIDResolver, digest).
					Return(&models.InstanceAPIKey{
						ID:        digest,
						Name:      "license-sync",
						CreatedBy: "user-1",
						ExpiresAt: now.Add(time.Hour),
					}, nil).
					Once()
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "user-1").
					Return(&models.User{ID: "user-1", Admin: true}, nil).
					Once()
			},
			expected: Expected{
				res: &models.InstanceAPIKey{
					ID:        digest,
					Name:      "license-sync",
					CreatedBy: "user-1",
					ExpiresAt: now.Add(time.Hour),
				},
				err: nil,
			},
		},
	}

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.AuthInstanceAPIKey(ctx, tc.key)
			require.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListInstanceAPIKeys(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

	ctx := context.Background()
	expected := []models.InstanceAPIKey{{ID: "digest", Name: "license-sync", CreatedBy: "user-1"}}

	queryOptionsMock.
		On("Sort", &query.Sorter{By: "created_at", Order: query.OrderAsc, Tiebreak: "key_digest"}).
		Return(nil).
		Once()
	queryOptionsMock.
		On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
		Return(nil).
		Once()
	storeMock.
		On("InstanceAPIKeyList", ctx, mock.AnythingOfType("[]store.QueryOption")).
		Return(expected, 1, nil).
		Once()

	apiKeys, count, err := s.ListInstanceAPIKeys(ctx, &requests.ListInstanceAPIKey{
		Paginator: query.Paginator{Page: 1, PerPage: 10},
		Sorter:    query.Sorter{By: "created_at", Order: query.OrderAsc},
	})

	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, expected, apiKeys)

	storeMock.AssertExpectations(t)
}

func TestDeleteInstanceAPIKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

	ctx := context.Background()

	t.Run("fails when the key does not exist", func(t *testing.T) {
		storeMock.
			On("InstanceAPIKeyDelete", ctx, "ghost").
			Return(store.ErrNoDocuments).
			Once()

		err := s.DeleteInstanceAPIKey(ctx, &requests.DeleteInstanceAPIKey{Name: "ghost"})
		require.Equal(t, NewErrInstanceAPIKeyNotFound("ghost", store.ErrNoDocuments), err)
	})

	t.Run("succeeds when the key exists", func(t *testing.T) {
		storeMock.
			On("InstanceAPIKeyDelete", ctx, "retired").
			Return(nil).
			Once()

		require.NoError(t, s.DeleteInstanceAPIKey(ctx, &requests.DeleteInstanceAPIKey{Name: "retired"}))
	})

	storeMock.AssertExpectations(t)
}
