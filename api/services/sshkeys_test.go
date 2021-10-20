package services

import (
	"context"
	"errors"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const (
	InvalidTenantID        = "invalid_tenant_id"
	InvalidFingerprint     = "invalid_fingerprint"
	invalidTenantIDStr     = "Fails when the tenant is invalid"
	InvalidFingerprintStr  = "Fails when the fingerprint is invalid"
	InvalidFingerTenantStr = "Fails when the fingerprint and tenant is invalid"
)

func TestListPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	keys := []models.PublicKey{
		{Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"}},
		{Data: []byte("teste2"), Fingerprint: "fingerprint2", CreatedAt: clock.Now(), TenantID: "tenant2", PublicKeyFields: models.PublicKeyFields{Name: "teste2"}},
	}

	validQuery := paginator.Query{Page: 1, PerPage: 10}
	invalidQuery := paginator.Query{Page: -1, PerPage: 10}

	Err := errors.New("error")

	type Expected struct {
		returnedKeys []models.PublicKey
		count        int
		err          error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		keys          []models.PublicKey
		query         paginator.Query
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when the querry is invalid",
			ctx:         ctx,
			keys:        keys,
			query:       invalidQuery,
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, invalidQuery).Return(nil, 0, Err).Once()
			},
			expected: Expected{nil, 0, Err},
		},
		{
			description: "Successful list the keys",
			ctx:         ctx,
			keys:        keys,
			query:       validQuery,
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, validQuery).Return(keys, len(keys), nil).Once()
			},
			expected: Expected{keys, len(keys), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKeys, count, err := s.ListPublicKeys(ctx, tc.query)
			assert.Equal(t, tc.expected, Expected{returnedKeys, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	Err := errors.New("error")

	type Expected struct {
		returnedKey *models.PublicKey
		err         error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		key           *models.PublicKey
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: key.Fingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, key.Fingerprint, InvalidTenantID).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{

			description: InvalidFingerprintStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: InvalidFingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, key.TenantID).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{

			description: InvalidFingerTenantStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: InvalidFingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, InvalidTenantID).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Successful get the key",
			ctx:         ctx,
			key:         &key,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, key.Fingerprint, key.TenantID).Return(&key, nil).Once()
			},
			expected: Expected{&key, nil},
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
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}
	keyUpdate := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste2"},
	}

	Err := errors.New("error")

	type Expected struct {
		returnedKey *models.PublicKey
		err         error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		key           *models.PublicKey
		fingerprint   string
		tenantID      string
		keyUpdate     *models.PublicKeyUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: key.Fingerprint,
			tenantID:    InvalidTenantID,
			keyUpdate:   keyUpdate,
			requiredMocks: func() {
				mock.On("PublicKeyUpdate", ctx, key.Fingerprint, InvalidTenantID, keyUpdate).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{

			description: InvalidFingerprintStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: InvalidFingerprint,
			tenantID:    key.TenantID,
			keyUpdate:   keyUpdate,
			requiredMocks: func() {
				mock.On("PublicKeyUpdate", ctx, InvalidFingerprint, key.TenantID, keyUpdate).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{

			description: InvalidFingerTenantStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: InvalidFingerprint,
			tenantID:    InvalidTenantID,
			keyUpdate:   keyUpdate,
			requiredMocks: func() {
				mock.On("PublicKeyUpdate", ctx, InvalidFingerprint, InvalidTenantID, keyUpdate).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Successful update the key",
			ctx:         ctx,
			key:         newKey,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			keyUpdate:   keyUpdate,
			requiredMocks: func() {
				mock.On("PublicKeyUpdate", ctx, key.Fingerprint, key.TenantID, keyUpdate).Return(newKey, nil).Once()
			},
			expected: Expected{newKey, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKey, err := s.UpdatePublicKey(ctx, tc.fingerprint, tc.tenantID, keyUpdate)
			assert.Equal(t, tc.expected, Expected{returnedKey, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeletePublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	Err := errors.New("error")

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
			fingerprint: key.Fingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("PublicKeyDelete", ctx, key.Fingerprint, InvalidTenantID).Return(Err).Once()
			},
			expected: Expected{Err},
		},
		{

			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("PublicKeyDelete", ctx, InvalidFingerprint, key.TenantID).Return(Err).Once()
			},
			expected: Expected{Err},
		},
		{

			description: InvalidFingerTenantStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("PublicKeyDelete", ctx, InvalidFingerprint, InvalidTenantID).Return(Err).Once()
			},
			expected: Expected{Err},
		},
		{
			description: "Successful delete the key",
			ctx:         ctx,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("PublicKeyDelete", ctx, key.Fingerprint, key.TenantID).Return(nil).Once()
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

	mock.AssertExpectations(t)
}
