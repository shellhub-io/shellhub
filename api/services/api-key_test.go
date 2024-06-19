package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIKey(t *testing.T) {
	type Expected struct {
		res *responses.CreateAPIKey
		err error
	}

	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		req           *requests.CreateAPIKey
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when namespace does not exists invalid",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
			},
		},
		{
			description: "fails when days to expire is invalid",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: 2,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrBadRequest(errors.New("experid date to APIKey is invalid")),
			},
		},
		{
			description: "fails when opt role is greater than user's role",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "administrator",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: -1,
				OptRole:   "owner",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res: nil,
				err: guard.ErrForbidden,
			},
		},
		{
			description: "fails when attributes are duplicated",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()

				keySum := sha256.Sum256([]byte("cdfd3cb0-c44e-4e54-b931-6d57713ad159"))
				hashedKey := hex.EncodeToString(keySum[:])

				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{ID: hashedKey, Name: "dev"}).
					Return([]string{"id", "name"}, true, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAPIKeyDuplicated([]string{"id", "name"}),
			},
		},
		{
			description: "fails when unable to create the key",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()

				keySum := sha256.Sum256([]byte("cdfd3cb0-c44e-4e54-b931-6d57713ad159"))
				hashedKey := hex.EncodeToString(keySum[:])

				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{ID: hashedKey, Name: "dev"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("APIKeyCreate", ctx, &models.APIKey{
						ID:        hashedKey,
						Name:      "dev",
						CreatedBy: "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "owner",
						ExpiresIn: -1,
					}).
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("error"),
			},
		},
		{
			description: "succeeds",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
				Name:      "dev",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()

				uuidMock := &uuidmock.Uuid{}
				uuid.DefaultBackend = uuidMock
				uuidMock.
					On("Generate").
					Return("cdfd3cb0-c44e-4e54-b931-6d57713ad159").
					Once()

				keySum := sha256.Sum256([]byte("cdfd3cb0-c44e-4e54-b931-6d57713ad159"))
				hashedKey := hex.EncodeToString(keySum[:])

				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{ID: hashedKey, Name: "dev"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("APIKeyCreate", ctx, &models.APIKey{
						ID:        hashedKey,
						Name:      "dev",
						CreatedBy: "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "owner",
						ExpiresIn: -1,
					}).
					Return(hashedKey, nil).
					Once()
				storeMock.
					On("APIKeyGet", ctx, hashedKey).
					Return(&models.APIKey{
						ID:        hashedKey,
						Name:      "dev",
						CreatedBy: "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "owner",
						ExpiresIn: -1,
					}, nil).
					Once()
			},
			expected: Expected{
				res: &responses.CreateAPIKey{
					ID:        "cdfd3cb0-c44e-4e54-b931-6d57713ad159",
					Name:      "dev",
					UserID:    "000000000000000000000000",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Role:      "owner",
					ExpiresIn: -1,
				},
				err: nil,
			},
		},
		{
			description: "succeeds when request key is empty",
			req: &requests.CreateAPIKey{
				UserID:    "000000000000000000000000",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "owner",
				Key:       "",
				Name:      "dev",
				ExpiresAt: -1,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(
						&models.Namespace{
							Name:     "namespace",
							Owner:    "000000000000000000000000",
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:   "000000000000000000000000",
									Role: "owner",
								},
							},
						},
						nil,
					).
					Once()

				uuidMock := &uuidmock.Uuid{}
				uuid.DefaultBackend = uuidMock
				uuidMock.
					On("Generate").
					Return("1e7b0f4b-aca4-48eb-a353-7469f00665ed").
					Once()

				keySum := sha256.Sum256([]byte("1e7b0f4b-aca4-48eb-a353-7469f00665ed"))
				hashedKey := hex.EncodeToString(keySum[:])

				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{ID: hashedKey, Name: "dev"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("APIKeyCreate", ctx, &models.APIKey{
						ID:        hashedKey,
						Name:      "dev",
						CreatedBy: "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "owner",
						ExpiresIn: -1,
					}).
					Return(hashedKey, nil).
					Once()
				storeMock.
					On("APIKeyGet", ctx, hashedKey).
					Return(&models.APIKey{
						ID:        hashedKey,
						Name:      "dev",
						CreatedBy: "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "owner",
						ExpiresIn: -1,
					}, nil).
					Once()
			},
			expected: Expected{
				res: &responses.CreateAPIKey{
					ID:        "1e7b0f4b-aca4-48eb-a353-7469f00665ed",
					Name:      "dev",
					UserID:    "000000000000000000000000",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Role:      "owner",
					ExpiresIn: -1,
				},
				err: nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(&Keys{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, storeMock, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.CreateAPIKey(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListAPIKey(t *testing.T) {
	type Expected struct {
		apiKeys []models.APIKey
		count   int
		err     error
	}

	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		tenantID      string
		req           *requests.ListAPIKey
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails",
			req: &requests.ListAPIKey{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Sorter:    query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("APIKeyList", ctx, "00000000-0000-4000-0000-000000000000", query.Paginator{Page: 1, PerPage: 10}, query.Sorter{By: "expires_in", Order: query.OrderAsc}).
					Return(nil, 0, errors.New("error")).
					Once()
			},
			expected: Expected{
				apiKeys: nil,
				count:   0,
				err:     errors.New("error"),
			},
		},
		{
			description: "succeeds",
			req: &requests.ListAPIKey{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Sorter:    query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("APIKeyList", ctx, "00000000-0000-4000-0000-000000000000", query.Paginator{Page: 1, PerPage: 10}, query.Sorter{By: "expires_in", Order: query.OrderAsc}).
					Return(
						[]models.APIKey{
							{
								CreatedBy: "id",
								Name:      "nameAPIKey",
							},
						},
						1,
						nil,
					).
					Once()
			},
			expected: Expected{
				apiKeys: []models.APIKey{
					{
						CreatedBy: "id",
						Name:      "nameAPIKey",
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(&Keys{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, storeMock, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			apiKeys, count, err := s.ListAPIKeys(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{apiKeys, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateAPIKey(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		req           *requests.UpdateAPIKey
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when namespaces does not exists",
			req: &requests.UpdateAPIKey{
				UserID:      "000000000000000000000000",
				TenantID:    "00000000-0000-4000-0000-000000000000",
				CurrentName: "dev",
				Name:        "newName",
				Role:        "administrator",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error")),
		},
		{
			description: "fails when role is greather than user's role",
			req: &requests.UpdateAPIKey{
				UserID:      "000000000000000000000000",
				TenantID:    "00000000-0000-4000-0000-000000000000",
				CurrentName: "dev",
				Name:        "newName",
				Role:        "owner",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(&models.Namespace{Members: []models.Member{{ID: "000000000000000000000000", Role: "administrator"}}}, nil).
					Once()
			},
			expected: guard.ErrForbidden,
		},
		{
			description: "fails when a conflict is found",
			req: &requests.UpdateAPIKey{
				UserID:      "000000000000000000000000",
				TenantID:    "00000000-0000-4000-0000-000000000000",
				CurrentName: "dev",
				Name:        "newName",
				Role:        "administrator",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(&models.Namespace{Members: []models.Member{{ID: "000000000000000000000000", Role: "owner"}}}, nil).
					Once()
				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{Name: "newName"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: NewErrAPIKeyDuplicated([]string{"name"}),
		},
		{
			description: "fails when api key does not exists",
			req: &requests.UpdateAPIKey{
				UserID:      "000000000000000000000000",
				TenantID:    "00000000-0000-4000-0000-000000000000",
				CurrentName: "dev",
				Name:        "newName",
				Role:        "administrator",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(&models.Namespace{Members: []models.Member{{ID: "000000000000000000000000", Role: "owner"}}}, nil).
					Once()
				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{Name: "newName"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("APIKeyUpdate", ctx, "00000000-0000-4000-0000-000000000000", "dev", &models.APIKeyChanges{Name: "newName", Role: "administrator"}).
					Return(errors.New("error")).
					Once()
			},
			expected: NewErrAPIKeyNotFound("dev", errors.New("error")),
		},
		{
			description: "succeeds",
			req: &requests.UpdateAPIKey{
				UserID:      "000000000000000000000000",
				TenantID:    "00000000-0000-4000-0000-000000000000",
				CurrentName: "dev",
				Name:        "newName",
				Role:        "administrator",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(&models.Namespace{Members: []models.Member{{ID: "000000000000000000000000", Role: "owner"}}}, nil).
					Once()
				storeMock.
					On("APIKeyConflicts", ctx, "00000000-0000-4000-0000-000000000000", &models.APIKeyConflicts{Name: "newName"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("APIKeyUpdate", ctx, "00000000-0000-4000-0000-000000000000", "dev", &models.APIKeyChanges{Name: "newName", Role: "administrator"}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(&Keys{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, storeMock, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			err := s.UpdateAPIKey(ctx, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeleteAPIKey(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		tenantID      string
		req           *requests.DeleteAPIKey
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when api key does not exists",
			req: &requests.DeleteAPIKey{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "dev",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("APIKeyDelete", ctx, "00000000-0000-4000-0000-000000000000", "dev").
					Return(errors.New("error")).
					Once()
			},
			expected: NewErrAPIKeyNotFound("dev", errors.New("error")),
		},
		{
			description: "succeeds",
			req: &requests.DeleteAPIKey{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "dev",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("APIKeyDelete", ctx, "00000000-0000-4000-0000-000000000000", "dev").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(&Keys{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, storeMock, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			err = s.DeleteAPIKey(ctx, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}
