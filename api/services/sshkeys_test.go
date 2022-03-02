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
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	type Expected struct {
		bool
		error
	}

	keyTagsNoExist := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	deviceTagsNoExist := models.Device{
		Tags: []string{"tag4"},
	}

	keyTags := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	deviceTags := models.Device{
		Tags: []string{"tag1"},
	}

	keyHostname := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}
	deviceHostname := models.Device{
		Name: "device",
	}

	keyHostnameNoMatch := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: "roo.*",
			},
		},
	}
	deviceHostnameNoMatch := models.Device{
		Name: "device",
	}

	keyNoFilter := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{},
		},
	}
	deviceNoFilter := models.Device{}

	cases := []struct {
		description   string
		key           *models.PublicKey
		device        models.Device
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to evaluate when filter hostname no match",
			key:         keyHostnameNoMatch,
			device:      deviceHostnameNoMatch,
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter hostname",
			key:         keyHostname,
			device:      deviceHostname,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "fail to evaluate filter tags when tag does not exist in device",
			key:         keyTagsNoExist,
			device:      deviceTagsNoExist,
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter tags",
			key:         keyTags,
			device:      deviceTags,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "success to evaluate when key has no filter",
			key:         keyNoFilter,
			device:      deviceNoFilter,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			ok, err := s.EvaluateKeyFilter(ctx, tc.key, tc.device)
			assert.Equal(t, tc.expected, Expected{ok, err})
		})
	}

	mock.AssertExpectations(t)
}

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
			description: "Fails when the query is invalid",
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

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	keyWithTags := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{Tags: []string{"tag3", "tag4"}},
		},
	}
	keyUpdateWithTags := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	keyWithHostname := &models.PublicKey{
		Fingerprint: "fingerprint",
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{Hostname: ".*"},
		},
	}
	keyUpdateWithHostname := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{Hostname: ".*"},
		},
	}
	keyInvalidUpdateNoHostnameTags := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{},
	}
	keyInvalidUpdateTwoFilters := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
				Tags:     []string{"tag1", "tag2"},
			},
		},
	}
	keyInvalidUpdateTagsEmpty := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{},
			},
		},
	}
	keyInvalidUpdateHostnameEmpty := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: "",
			},
		},
	}

	type Expected struct {
		key *models.PublicKey
		err error
	}

	cases := []struct {
		description   string
		fingerprint   string
		tenantID      string
		keyUpdate     *models.PublicKeyUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description:   "fail when public does not contains hostname neither tags",
			fingerprint:   "fingerprint",
			tenantID:      "tenant",
			keyUpdate:     keyInvalidUpdateNoHostnameTags,
			requiredMocks: func() {},
			expected:      Expected{key: nil, err: ErrPublicKeyInvalid},
		},
		{
			description: "fails to update a public key when filter has hostname and tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyInvalidUpdateTwoFilters,
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrPublicKeyInvalid},
		},
		{
			description: "fail update the key when filter hostname is empty",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyInvalidUpdateHostnameEmpty,
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrPublicKeyInvalid},
		},
		{
			description: "successful update the key when filter is hostname",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", keyUpdateWithHostname).Return(keyWithHostname, nil).Once()
			},
			expected: Expected{keyWithHostname, nil},
		},
		{
			description: "fail to update the key when a tag does not exist in a device",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag4"}, 2, nil).Once()
			},
			expected: Expected{nil, ErrTagNameNotFound},
		},
		{
			description: "fail update the key when filter tags is empty",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyInvalidUpdateTagsEmpty,
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrPublicKeyInvalid},
		},
		{
			description: "Successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", keyUpdateWithTags).Return(keyWithTags, nil).Once()
			},
			expected: Expected{keyWithTags, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedKey, err := s.UpdatePublicKey(ctx, tc.fingerprint, tc.tenantID, tc.keyUpdate)
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

func TestCreatePublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now)

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	err := errors.New("")

	ctx := context.TODO()

	pubKey, _ := ssh.NewPublicKey(publicKey)
	data := ssh.MarshalAuthorizedKey(pubKey)
	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	keyInvalidNoFilter := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Username: "",
			Filter:   models.PublicKeyFilter{},
		},
	}
	keyInvalidBothFilter := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Username: "",
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
				Tags:     []string{"tag1", "tag2"},
			},
		},
	}
	keyInvalidData := &models.PublicKey{
		Data:        nil,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}
	keyInvalidEmptyTags := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{},
			},
		},
	}
	keyWithTags := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	keyInvalidHostnameEmpty := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: "",
			},
		},
	}
	keyWithHostname := &models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}

	cases := []struct {
		description   string
		tenantID      string
		key           *models.PublicKey
		requiredMocks func()
		expected      error
	}{
		{
			description: "fail when public key has no filter",
			tenantID:    "tenant",
			key:         keyInvalidNoFilter,
			requiredMocks: func() {
			},
			expected: ErrPublicKeyInvalid,
		},
		{
			description: "fail when public key has hostname and tags filter",
			tenantID:    "tenant",
			key:         keyInvalidBothFilter,
			requiredMocks: func() {
			},
			expected: ErrPublicKeyInvalid,
		},
		{
			description: "fail when data in public key is not valid",
			tenantID:    "tenant",
			key:         keyInvalidData,
			requiredMocks: func() {
			},
			expected: ErrInvalidFormat,
		},
		{
			description: "fail when can not get the public key",
			tenantID:    "tenant",
			key:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, err).Once()
			},
			expected: err,
		},
		{
			description: "fail when public key is duplicated",
			tenantID:    "tenant",
			key:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(keyWithHostname, nil).Once()
			},
			expected: ErrDuplicateFingerprint,
		},
		{
			description: "fail when can not create the public key",
			tenantID:    "tenant",
			key:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, keyWithHostname).Return(err).Once()
			},
			expected: err,
		},
		{
			description: "fail to create a public key when filter is hostname is empty",
			tenantID:    "tenant",
			key:         keyInvalidHostnameEmpty,
			requiredMocks: func() {
			},
			expected: ErrPublicKeyInvalid,
		},
		{
			description: "success create a public key when filter is hostname",
			tenantID:    "tenant",
			key:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, keyWithHostname).Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "fail to create a public key when filter tag is empty",
			tenantID:    "tenant",
			key:         keyInvalidEmptyTags,
			requiredMocks: func() {
			},
			expected: ErrPublicKeyInvalid,
		},
		{
			description: "success create a public key when filter is tags",
			tenantID:    "tenant",
			key:         keyWithTags,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("TagsGet", ctx, keyWithTags.TenantID).Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyCreate", ctx, keyWithTags).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := s.CreatePublicKey(ctx, tc.key, tc.tenantID)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
