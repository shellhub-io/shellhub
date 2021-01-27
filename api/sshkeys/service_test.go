package sshkeys

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListPublicKeys(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()
	keys := []models.PublicKey{
		{Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"}},
		{Data: []byte("teste2"), Fingerprint: "fingerprint2", CreatedAt: time.Now(), TenantID: "tenant2", PublicKeyFields: models.PublicKeyFields{Name: "teste2"}},
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	mock.On("ListPublicKeys", ctx, query).Return(keys, len(keys), nil).Once()

	returnedKeys, count, err := s.ListPublicKeys(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, keys, returnedKeys)
	assert.Equal(t, count, len(keys))

	mock.AssertExpectations(t)
}

func TestGetPublicKeys(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	mock.On("GetPublicKey", ctx, key.Fingerprint, key.TenantID).Return(key, nil).Once()

	returnedKey, err := s.GetPublicKey(ctx, key.Fingerprint, key.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, key, returnedKey)

	mock.AssertExpectations(t)
}

func TestUpdatePublicKeys(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}
	keyUpdate := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste2"},
	}

	mock.On("UpdatePublicKey", ctx, key.Fingerprint, key.TenantID, keyUpdate).Return(newKey, nil).Once()

	returnedKey, err := s.UpdatePublicKey(ctx, key.Fingerprint, key.TenantID, keyUpdate)
	assert.NoError(t, err)
	assert.Equal(t, newKey, returnedKey)

	mock.AssertExpectations(t)
}

func TestDeletePublicKeys(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: time.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	mock.On("DeletePublicKey", ctx, key.Fingerprint, key.TenantID).Return(nil).Once()

	err := s.DeletePublicKey(ctx, key.Fingerprint, key.TenantID)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestCreatePublicKey(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()
	sshRsa := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDhZgLqiZYDKCWhyi2gUXIRwIPyMSyXZ6yrwsm3PYfIvFB60kVlNgqDpPVhWoH6eRfaQ1y/xbg4nClZmHEDTvLbTQ1ZoQzzjZ7zvM6aQ4nADmKcCYswEuU94axouVjsHNyMLfOkPXuGec0fChwQ2JDh/B9LCiSDxyhCOgHvETXGXsyBMKjn498iPjJ6snzk35dy5wPZRz41g3dLaygF+wYAT791u/JchHQL7OP7RoNgby+RM16SYZs1tgQVkfU//o+AyTarWYLVDpFU6HPPenE4xEXhbgqd7x3wSNPBsMvY8Zjcu3kdHtboJidyMtKeD8ghV/T24kME58TW15T8Eg8R"
	key := &models.PublicKey{
		Data: []byte(sshRsa), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	mock.On("CreatePublicKey", ctx, key).Return(nil).Once()

	err := s.CreatePublicKey(ctx, key)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
