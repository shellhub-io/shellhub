package services

import (
	"context"
	"errors"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAddPublicKeyTag(t *testing.T) {
	mock := &mocks.Store{}
	services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()
	err := errors.New("generic errors")

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tag           string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fail when namespace was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(nil, err)
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "fail when public key was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(nil, err).Once()
			},
			expected: ErrPublicKeyNotFound,
		},
		{
			description: "fail when the tag limit on public key has reached",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}

				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
			},
			expected: ErrMaxTagReached,
		},
		{
			description: "fail when the tag does not exist in a device",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag1", "tag2"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: tags,
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Twice()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("DeviceGetTags", ctx, "tenant").Return(tags, len(tags), nil).Once()
			},
			expected: ErrTagNameNotFound,
		},
		{
			description: "fail when cannot add tag to public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag", "tag3", "tag6"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Twice()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("DeviceGetTags", ctx, "tenant").Return(tags, len(tags), nil).Once()
				mock.On("PublicKeyAddTag", ctx, "tenant", "fingerprint", "tag").Return(err).Once()
			},
			expected: err,
		},
		{
			description: "success to add a to public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag", "tag3", "tag6"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Twice()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("DeviceGetTags", ctx, "tenant").Return(tags, len(tags), nil).Once()
				mock.On("PublicKeyAddTag", ctx, "tenant", "fingerprint", "tag").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := services.AddPublicKeyTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
