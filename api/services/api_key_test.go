package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateAPIKey(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	type Expected struct {
		APIKeys string
		err     error
	}

	tests := []struct {
		description            string
		APIKeyRequest          requests.CreateAPIKey
		apiKey, tenant, userID string
		requiredMocks          func()
		expected               Expected
	}{
		{
			description: "fails when tenant is invalid",
			userID:      "id",
			tenant:      "",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "", false).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				APIKeys: "",
				err:     NewErrNamespaceNotFound("", errors.New("error", "", 0)),
			},
		},
		{
			description: "expired time invalid",
			userID:      "id",
			tenant:      "00000000-0000-4000-0000-000000000000",
			APIKeyRequest: requests.CreateAPIKey{
				Name:      "nameAPIKey",
				ExpiresAt: 14,
				TenantParam: requests.TenantParam{
					Tenant: "00000000-0000-4000-0000-000000000000",
				},
			},
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "id",
					TenantID: "xxxxxx",
					Members: []models.Member{
						{
							ID:   "memberID",
							Role: "owner",
						},
					},
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).Return(namespace, nil).Once()

				clockMock.On("Now").Return(now).Twice()
			},
			expected: Expected{
				APIKeys: "",
				err:     errors.New("experid date to APIKey is invalid", "", 0),
			},
		},
		{
			description: "fails when try to create a APIKey with different namespace",
			tenant:      "00000000-0000-4000-0000-000000000000",
			APIKeyRequest: requests.CreateAPIKey{
				Name:      "nameAPIKey",
				ExpiresAt: -1,
				TenantParam: requests.TenantParam{
					Tenant: "00000000-0000-5000-0000-000000000000",
				},
			},
			apiKey: "key",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "id",
					TenantID: "xxxxxx",
					Members: []models.Member{
						{
							ID:   "memberID",
							Role: "owner",
						},
					},
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-5000-0000-000000000000", false).Return(namespace, nil).Once()

				clockMock.On("Now").Return(now).Twice()
			},
			expected: Expected{
				APIKeys: "",
				err:     NewErrAuthUnathorized(errors.New("APIKey creation not allowed to different namespace", "", 0)),
			},
		},
		{
			description: "success when try to get a APIKey",
			userID:      "id",
			tenant:      "00000000-0000-4000-0000-000000000000",
			APIKeyRequest: requests.CreateAPIKey{
				Name:      "nameAPIKey",
				ExpiresAt: -1,
				TenantParam: requests.TenantParam{
					Tenant: "00000000-0000-4000-0000-000000000000",
				},
			},
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "id",
					TenantID: "xxxxxx",
					Members: []models.Member{
						{
							ID:   "memberID",
							Role: "owner",
						},
					},
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).Return(namespace, nil).Once()
				mock.On("APIKeyGetByName", ctx, "00000000-0000-4000-0000-000000000000", "nameAPIKey").Return(nil, nil).Once()
				mock.On("APIKeyCreate", ctx, gomock.Anything).Return(nil).Once()

				clockMock.On("Now").Return(now).Twice()
			},
			expected: Expected{
				APIKeys: "apikey",
				err:     nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(mock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			authRes, err := service.CreateAPIKey(ctx, tc.userID, tc.tenant, tc.apiKey, &tc.APIKeyRequest)
			if tc.expected.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expected.err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, authRes)
			}

			mock.AssertExpectations(t)
		})
	}
}

func TestListAPIKey(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	type Expected struct {
		APIKeys []models.APIKey
		Count   int
		err     error
	}
	tests := []struct {
		description   string
		requestParams *requests.APIKeyList
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when user is not found",
			requestParams: &requests.APIKeyList{
				TenantParam: requests.TenantParam{Tenant: ""},
				Paginator:   query.Paginator{Page: 1, PerPage: 10},
				Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			requiredMocks: func() {
				req := &requests.APIKeyList{
					TenantParam: requests.TenantParam{Tenant: ""},
					Paginator:   query.Paginator{Page: 1, PerPage: 10},
					Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
				}
				mock.On("APIKeyList", ctx, "", req.Paginator, req.Sorter).Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				APIKeys: nil,
				Count:   0,
				err:     NewErrAPIKeyNotFound("", errors.New("error", "", 0)),
			},
		},
		{
			description: "success when user is found",
			requestParams: &requests.APIKeyList{
				TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
				Paginator:   query.Paginator{Page: 1, PerPage: 10},
				Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			requiredMocks: func() {
				APIKey := []models.APIKey{
					{
						UserID: "id",
						Name:   "nameAPIKey",
					},
				}

				req := &requests.APIKeyList{
					TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
					Paginator:   query.Paginator{Page: 1, PerPage: 10},
					Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
				}

				mock.On("APIKeyList", ctx, "00000000-0000-4000-0000-000000000000", req.Paginator, req.Sorter).Return(APIKey, 1, nil).Once()
			},
			expected: Expected{
				APIKeys: []models.APIKey{
					{
						UserID: "id",
						Name:   "nameAPIKey",
					},
				},
				Count: 1,
				err:   nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(mock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			APIKeys, count, err := service.ListAPIKeys(ctx, tc.requestParams)
			assert.Equal(t, tc.expected, Expected{APIKeys, count, err})
			mock.AssertExpectations(t)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	tests := []struct {
		description   string
		tenant        string
		id            string
		requiredMocks func()
		expectedErr   error
	}{
		{
			description: "fails when try delete a apikey",
			id:          "",
			requiredMocks: func() {
				mock.On("APIKeyDelete", ctx, "", "").Return(errors.New("APIKey not found", "", 0)).Once()
			},
			expectedErr: NewErrAPIKeyNotFound("", errors.New("APIKey not found", "", 0)),
		},
		{
			description: "success when try delete a apikey",
			id:          "id",
			requiredMocks: func() {
				mock.On("APIKeyDelete", ctx, "id", "").Return(nil).Once()
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(mock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			err = service.DeleteAPIKey(ctx, tc.id, tc.tenant)
			assert.Equal(t, tc.expectedErr, err)

			mock.AssertExpectations(t)
		})
	}
}

func TestEditAPIKey(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	type Expected struct {
		APIKey *models.APIKey
		err    error
	}
	tests := []struct {
		description   string
		requestParams *requests.APIKeyChanges
		tenant        string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success when try rename a APIKey",
			requestParams: &requests.APIKeyChanges{
				ID:   "id",
				Name: "newName",
			},
			requiredMocks: func() {
				req := &requests.APIKeyChanges{
					ID:   "id",
					Name: "newName",
				}
				mock.On("APIKeyEdit", ctx, req).Return(nil).Once()
				mock.On("APIKeyGetByName", ctx, "tenant", "newName").Return(nil, nil).Once()
				mock.On("APIKeyGetByUID", ctx, "id").Return(&models.APIKey{}, nil).Once()
			},
			expected: Expected{
				APIKey: &models.APIKey{},
				err:    nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(mock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			req, err := service.EditAPIKey(ctx, "tenant", tc.requestParams)
			assert.Equal(t, tc.expected.err, err)
			assert.True(t, reflect.DeepEqual(tc.expected.APIKey, req))

			mock.AssertExpectations(t)
		})
	}
}
