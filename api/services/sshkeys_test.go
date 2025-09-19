package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ssh"
)

const (
	InvalidTenantID        = "invalid_tenant_id"
	InvalidFingerprint     = "invalid_fingerprint"
	invalidTenantIDStr     = "Fails when the tenant is invalid"
	InvalidFingerprintStr  = "Fails when the fingerprint is invalid"
	InvalidFingerTenantStr = "Fails when the fingerprint and tenant is invalid"
)

func TestEvaluateKeyFilter(t *testing.T) {
	storeMock := &storemock.Store{}

	ctx := context.TODO()

	type Expected struct {
		bool
		error
	}

	cases := []struct {
		description   string
		key           *models.PublicKey
		device        models.Device
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to evaluate when filter hostname no match",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: "roo.*",
					},
				},
			},
			device: models.Device{
				Name: "device",
			},
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter hostname",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
					},
				},
			},
			device: models.Device{
				Name: "device",
			},
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "fail to evaluate filter tags when DeviceResolve fails",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}},
					},
				},
			},
			device: models.Device{UID: "uid"},
			requiredMocks: func() {
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{false, NewErrDeviceNotFound("uid", errors.New("error", "", 0))},
		},
		{
			description: "fail to evaluate filter tags when tag does not exist in device",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}},
					},
				},
			},
			device: models.Device{UID: "uid"},
			requiredMocks: func() {
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid").
					Return(&models.Device{UID: "uid", Taggable: models.Taggable{TagIDs: []string{"nonexistent_id"}}}, nil).
					Once()
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter tags",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}},
					},
				},
			},
			device: models.Device{UID: "uid"},
			requiredMocks: func() {
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid").
					Return(&models.Device{UID: "uid", Taggable: models.Taggable{TagIDs: []string{"tag1_id"}}}, nil).
					Once()
			},
			expected: Expected{true, nil},
		},
		{
			description: "success to evaluate when key has no filter",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{},
				},
			},
			device: models.Device{},
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			ok, err := service.EvaluateKeyFilter(ctx, tc.key, tc.device)
			assert.Equal(t, tc.expected, Expected{ok, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListPublicKeys(t *testing.T) {
	storeMock := &storemock.Store{}
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	ctx := context.TODO()

	keys := []models.PublicKey{
		{Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"}},
		{Data: []byte("teste2"), Fingerprint: "fingerprint2", CreatedAt: clock.Now(), TenantID: "tenant2", PublicKeyFields: models.PublicKeyFields{Name: "teste2"}},
	}

	type Expected struct {
		returnedKeys []models.PublicKey
		count        int
		err          error
	}

	cases := []struct {
		description   string
		keys          []models.PublicKey
		req           *requests.ListPublicKeys
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when the query is invalid",
			req: &requests.ListPublicKeys{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{nil, 0, errors.New("error", "", 0)},
		},
		{
			description: "Successful list the keys",
			keys:        keys,
			req: &requests.ListPublicKeys{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(keys, len(keys), nil).
					Once()
			},
			expected: Expected{keys, len(keys), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKeys, count, err := s.ListPublicKeys(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{returnedKeys, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetPublicKeys(t *testing.T) {
	mock := &storemock.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	ctx := context.TODO()

	type Expected struct {
		returnedKey *models.PublicKey
		err         error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, InvalidTenantID).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrNamespaceNotFound(InvalidTenantID, errors.New("error", "", 0))},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := models.Namespace{TenantID: "tenant1"}

				mock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, "tenant1").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "Successful get the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := models.Namespace{TenantID: "tenant1"}
				key := models.PublicKey{
					Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
				}
				mock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant1").Return(&key, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKey, err := s.GetPublicKey(ctx, tc.fingerprint, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{returnedKey, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePublicKeys(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	type Expected struct {
		key *models.PublicKey
		err error
	}

	cases := []struct {
		description   string
		fingerprint   string
		tenantID      string
		keyUpdate     requests.PublicKeyUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail when public key not found",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{nil, NewErrPublicKeyNotFound("fingerprint", store.ErrNoDocuments)},
		},
		{
			description: "fail update the key when tag list retrieval fails",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				existingKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
				}
				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(existingKey, nil).Once()
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", errors.New("error", "", 0))},
		},
		{
			description: "fail to update the key when a tag does not exist",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				existingKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
				}
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag4_id", Name: "tag4", TenantID: "tenant"},
				}
				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(existingKey, nil).Once()
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "fail update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				existingKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{},
					},
				}
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag2_id", Name: "tag2", TenantID: "tenant"},
				}

				expectedKey := *existingKey
				expectedKey.Filter.TagIDs = []string{"tag1_id", "tag2_id"}
				expectedKey.Filter.Tags = nil

				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(existingKey, nil).Once()
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
				storeMock.On("PublicKeyUpdate", ctx, &expectedKey).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				existingKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{},
					},
				}
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag2_id", Name: "tag2", TenantID: "tenant"},
				}

				expectedKey := *existingKey
				expectedKey.Filter.TagIDs = []string{"tag1_id", "tag2_id"}
				expectedKey.Filter.Tags = nil

				updatedKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}},
						},
					},
				}

				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(existingKey, nil).Once()
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
				storeMock.On("PublicKeyUpdate", ctx, &expectedKey).Return(nil).Once()
				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(updatedKey, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				Fingerprint: "fingerprint",
				TenantID:    "tenant",
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}},
					},
				},
			}, nil},
		},
		{
			description: "successful update the key when filter is hostname",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				existingKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{},
					},
				}

				expectedKey := *existingKey
				expectedKey.Filter.Hostname = ".*"
				expectedKey.Filter.TagIDs = []string{}
				expectedKey.Filter.Tags = nil

				updatedKey := &models.PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(existingKey, nil).Once()
				storeMock.On("PublicKeyUpdate", ctx, &expectedKey).Return(nil).Once()
				storeMock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(updatedKey, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				Fingerprint: "fingerprint",
				TenantID:    "tenant",
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
					},
				},
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedKey, err := s.UpdatePublicKey(ctx, tc.fingerprint, tc.tenantID, tc.keyUpdate)
			assert.Equal(t, tc.expected, Expected{returnedKey, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeletePublicKeys(t *testing.T) {
	storeMock := new(storemock.Store)

	ctx := context.TODO()

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	type Expected struct {
		err error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, InvalidTenantID).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{NewErrNamespaceNotFound(InvalidTenantID, errors.New("error", "", 0))},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}

				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, namespace.TenantID).Return(namespace, nil).Once()
				storeMock.On("PublicKeyGet", ctx, InvalidFingerprint, namespace.TenantID).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{NewErrPublicKeyNotFound(InvalidFingerprint, errors.New("error", "", 0))},
		},
		{
			description: "fail to delete the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}
				publicKey := &models.PublicKey{
					Data:            []byte("teste"),
					Fingerprint:     "fingerprint",
					CreatedAt:       clock.Now(),
					TenantID:        "tenant1",
					PublicKeyFields: models.PublicKeyFields{Name: "teste"},
				}

				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, namespace.TenantID).Return(namespace, nil).Once()
				storeMock.On("PublicKeyGet", ctx, "fingerprint", namespace.TenantID).
					Return(publicKey, nil).Once()
				storeMock.On("PublicKeyDelete", ctx, publicKey).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{errors.New("error", "", 0)},
		},
		{
			description: "successful to delete the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}
				publicKey := &models.PublicKey{
					Data:            []byte("teste"),
					Fingerprint:     "fingerprint",
					CreatedAt:       clock.Now(),
					TenantID:        "tenant1",
					PublicKeyFields: models.PublicKeyFields{Name: "teste"},
				}

				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, namespace.TenantID).Return(namespace, nil).Once()
				storeMock.On("PublicKeyGet", ctx, "fingerprint", namespace.TenantID).
					Return(publicKey, nil).Once()
				storeMock.On("PublicKeyDelete", ctx, publicKey).Return(nil).Once()
			},
			expected: Expected{nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := s.DeletePublicKey(ctx, tc.fingerprint, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestCreatePublicKeys(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	clockMock.On("Now").Return(now)

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	pubKey, _ := ssh.NewPublicKey(publicKey)

	type Expected struct {
		res *responses.PublicKeyCreate
		err error
	}

	cases := []struct {
		description   string
		tenantID      string
		req           requests.PublicKeyCreate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to create the key when tag list retrieval fails",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1"},
				},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", errors.New("error", "", 0))},
		},
		{
			description: "fail to create the key when a tag does not exist",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2", "tag4"},
				},
			},
			requiredMocks: func() {
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag4_id", Name: "tag4", TenantID: "tenant"},
				}
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "fail when data in public key is not valid",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        nil,
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
			},
			expected: Expected{nil, NewErrPublicKeyDataInvalid(requests.PublicKeyCreate{
				Data:        nil,
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			}.Data, nil)},
		},
		{
			description: "fail when cannot get the public key",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				storeMock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrPublicKeyNotFound(requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			}.Fingerprint, errors.New("error", "", 0))},
		},
		{
			description: "fail when public key is duplicated",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				storeMock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(&keyWithHostnameModel, nil).Once()
			},
			expected: Expected{nil, NewErrPublicKeyDuplicated([]string{ssh.FingerprintLegacyMD5(pubKey)}, nil)},
		},
		{
			description: "fail to create a public key when filter is hostname",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Taggable: models.Taggable{TagIDs: []string{}, Tags: nil},
						},
					},
				}

				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				storeMock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "success to create a public key when filter is hostname",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Taggable: models.Taggable{TagIDs: []string{}, Tags: nil},
						},
					},
				}

				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				storeMock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(nil).Once()
			},
			expected: Expected{&responses.PublicKeyCreate{
				Data: ssh.MarshalAuthorizedKey(pubKey),
				Filter: responses.PublicKeyFilter{
					Hostname: ".*",
					Tags:     nil,
				},
				Name:        "",
				Username:    "",
				TenantID:    "tenant",
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
			}, nil},
		},
		{
			description: "fail to create a public key when filter is tags",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag2_id", Name: "tag2", TenantID: "tenant"},
				}

				keyWithTags := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				}

				keyWithTagsModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}, Tags: nil},
						},
					},
				}

				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
				storeMock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "success to create a public key when filter is tags",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				tags := []models.Tag{
					{ID: "tag1_id", Name: "tag1", TenantID: "tenant"},
					{ID: "tag2_id", Name: "tag2", TenantID: "tenant"},
				}

				keyWithTags := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				}

				keyWithTagsModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Taggable: models.Taggable{TagIDs: []string{"tag1_id", "tag2_id"}, Tags: nil},
						},
					},
				}

				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.On("TagList", ctx, mock.AnythingOfType("store.QueryOption")).Return(tags, len(tags), nil).Once()
				storeMock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(nil).Once()
			},
			expected: Expected{&responses.PublicKeyCreate{
				Data: ssh.MarshalAuthorizedKey(pubKey),
				Filter: responses.PublicKeyFilter{
					Hostname: "",
					Tags:     []string{"tag1", "tag2"},
				},
				Name:        "",
				Username:    "",
				TenantID:    "tenant",
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			res, err := s.CreatePublicKey(ctx, tc.req, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}
