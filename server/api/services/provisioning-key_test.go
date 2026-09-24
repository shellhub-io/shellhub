package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateProvisioningKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clockMock := clockmock.NewMockClock(t)
	prevClock := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	defer func() { clock.DefaultBackend = prevClock }()
	clockMock.On("Now").Return(now).Maybe()

	prevUUID := uuid.DefaultBackend
	defer func() { uuid.DefaultBackend = prevUUID }()

	const tenant = "00000000-0000-4000-0000-000000000000"

	namespace := &models.Namespace{Name: "namespace", Owner: "000000000000000000000000", TenantID: tenant}
	generated := "1e7b0f4b-aca4-48eb-a353-7469f00665ed"
	plain := generated
	keySum := sha256.Sum256([]byte(plain))
	hashedKey := hex.EncodeToString(keySum[:])

	future := now.AddDate(0, 0, 30)
	days30 := 30

	matchCreate := func(want *models.ProvisioningKey) any {
		return mock.MatchedBy(func(got *models.ProvisioningKey) bool {
			if got.KeyEncrypted == "" || got.KeyHint != provisioningKeyHint(generated) {
				return false
			}

			if got.Mode != models.ProvisioningKeyModeAutomatic || len(got.AllowedIdentities) != 0 {
				return false
			}

			c := *got
			c.KeyEncrypted = ""
			c.KeyHint = ""
			c.Mode = ""
			c.AllowedIdentities = nil

			return reflect.DeepEqual(&c, want)
		})
	}

	cases := []struct {
		description   string
		req           *requests.CreateProvisioningKey
		requiredMocks func(ctx context.Context)
		expectedKey   string
		expectedErr   error
	}{
		{
			description: "fails when namespace does not exist",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(nil, errors.New("error")).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, errors.New("error")),
		},
		{
			description: "creates a key that never expires when expires_in is omitted",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(&models.ProvisioningKey{ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "fails when webhook mode has no http(s) url",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci", Mode: "webhook", WebhookSecret: "s"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{"webhook_url": "must be an http or https URL"}),
		},
		{
			description: "fails when webhook mode has no secret",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci", Mode: "webhook", WebhookURL: "https://hook.example/enroll"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{"webhook_secret": "is required for webhook mode"}),
		},
		{
			description: "fails when allowlist mode has no identity",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci", Mode: "allowlist"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{"allowed_identities": "at least one identity is required for allowlist mode"}),
		},
		{
			description: "fails when the name is duplicated",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{"name"}, true, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyDuplicated([]string{"name"}),
		},
		{
			description: "fails when the created key cannot be resolved back",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(nil, errors.New("resolve error")).Once()
			},
			expectedErr: errors.New("resolve error"),
		},
		{
			description: "derives a reusable key from a usage limit above one",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", UsageLimit: 10, Tags: []string{"prod"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true, UsageLimit: 10,
					Tags: []string{"prod"}, CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(&models.ProvisioningKey{
						ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true, UsageLimit: 10,
						Tags: []string{"prod"}, CreatedBy: "000000000000000000000000",
					}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "creates a single-use key when the usage limit is one",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", UsageLimit: 1},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: false, UsageLimit: 1,
					CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(&models.ProvisioningKey{
						ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: false, UsageLimit: 1,
						CreatedBy: "000000000000000000000000",
					}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "converts expires_in days to an absolute expiry",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", ExpiresIn: &days30},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					ExpiresAt: &future, CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(&models.ProvisioningKey{ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true}, nil).Once()
			},
			expectedKey: plain,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.CreateProvisioningKey(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				require.NotNil(t, res)
				require.Equal(t, tc.expectedKey, res.Key)
			}
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListProvisioningKeys(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	const tenant = "00000000-0000-4000-0000-000000000000"

	req := &requests.ListProvisioningKey{
		TenantID:  tenant,
		Paginator: query.Paginator{Page: 1, PerPage: 10},
		Sorter:    query.Sorter{By: "created_at", Order: query.OrderDesc},
	}

	queryOptionsMock.On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "key_digest"}).Return(nil).Once()
	queryOptionsMock.On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).Return(nil).Once()
	storeMock.On("ProvisioningKeyList", mock.Anything, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
		Return([]models.ProvisioningKey{{Name: "ci", TenantID: tenant}}, 1, nil).Once()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	keys, count, err := s.ListProvisioningKeys(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, []models.ProvisioningKey{{Name: "ci", TenantID: tenant}}, keys)

	storeMock.AssertExpectations(t)
}

func TestUpdateProvisioningKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clockMock := clockmock.NewMockClock(t)
	prevClock := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	defer func() { clock.DefaultBackend = prevClock }()
	clockMock.On("Now").Return(now).Maybe()

	const tenant = "00000000-0000-4000-0000-000000000000"
	namespace := &models.Namespace{Name: "namespace", TenantID: tenant}
	truePtr := true
	falsePtr := false
	limitTwo := 2
	limitUnlimited := 0
	ephemeralTimeout5 := 5
	modeAutomatic := "automatic"
	days60 := 60
	days0 := 0
	days36501 := 36501

	cases := []struct {
		description   string
		req           *requests.UpdateProvisioningKey
		requiredMocks func(ctx context.Context)
		expectedErr   error
	}{
		{
			description: "fails when namespace does not exist",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(nil, errors.New("error")).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, errors.New("error")),
		},
		{
			description: "fails when the key does not exist",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: NewErrProvisioningKeyNotFound("ci", store.ErrNoDocuments),
		},
		{
			description: "fails when the new name is duplicated",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Name: "runners"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{Name: "runners"}).
					Return([]string{"name"}, true, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyDuplicated([]string{"name"}),
		},
		{
			description: "revokes the key",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Revoked: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Reusable: true}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Reusable: true, Revoked: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "changes the legacy key's enrollment mode",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Mode: &modeAutomatic},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeAutomatic}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "allows disabling the legacy key to turn off keyless enrollment",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Disabled: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual, Disabled: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects changing a fixed field (name/limit/tags) on the legacy key",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Name: "renamed"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyForbidden(),
		},
		{
			description: "fails when lowering the usage limit below the used count",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", UsageLimit: &limitTwo},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 5, UsedTimes: 3}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{
				"usage_limit": "cannot be lower than the number of times the key was already used",
			}),
		},
		{
			description: "allows setting the usage limit to unlimited regardless of used count",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", UsageLimit: &limitUnlimited},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 5, UsedTimes: 3}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 0, UsedTimes: 3, Reusable: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "renames and retags the key",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Name: "runners", Tags: []string{"prod"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{Name: "runners"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "runners", TenantID: tenant, Tags: []string{"prod"}}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "turns on ephemeral with a timeout",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &truePtr, EphemeralTimeout: &ephemeralTimeout5},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 5}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "defaults the timeout to the max when ephemeral is turned on without one",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 10}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "clears the timeout when ephemeral is turned off",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &falsePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 10}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: false, EphemeralTimeout: 0}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "sets a new expiry from expires_in days",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", ExpiresIn: requests.OptionalInt{Present: true, Value: &days60}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				expiry := now.AddDate(0, 0, 60)
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, ExpiresAt: &expiry}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "clears the expiry when expires_in is null",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", ExpiresIn: requests.OptionalInt{Present: true, Value: nil}},
			requiredMocks: func(ctx context.Context) {
				existing := now.AddDate(0, 0, 30)
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, ExpiresAt: &existing}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, ExpiresAt: nil}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "leaves the expiry unchanged when expires_in is omitted",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Revoked: &truePtr},
			requiredMocks: func(ctx context.Context) {
				existing := now.AddDate(0, 0, 30)
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, ExpiresAt: &existing, Reusable: true}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, ExpiresAt: &existing, Reusable: true, Revoked: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects expires_in below 1",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", ExpiresIn: requests.OptionalInt{Present: true, Value: &days0}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{
				"expires_in": "must be between 1 and 36500",
			}),
		},
		{
			description: "rejects expires_in above 36500",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", ExpiresIn: requests.OptionalInt{Present: true, Value: &days36501}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyInvalidField(map[string]string{
				"expires_in": "must be between 1 and 36500",
			}),
		},
		{
			description: "rejects changing ephemeral on the legacy key",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Ephemeral: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyForbidden(),
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			err := s.UpdateProvisioningKey(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestAppendProvisioningKeyEvent(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	req := requests.DeviceAuth{
		TenantID: tenant,
		Identity: &requests.DeviceIdentity{MAC: "00:1a:2b:3c:4d:5e"},
		Info:     &requests.DeviceInfo{ID: "debian", PrettyName: "Debian GNU/Linux 12", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
		RealIP:   "203.0.113.7",
	}
	key := &models.ProvisioningKey{ID: "digest", Name: "ci", TenantID: tenant, Ephemeral: true}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("records the enrollment with the denormalized device facts", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("ProvisioningKeyEventCreate", mock.Anything, mock.MatchedBy(func(e *models.ProvisioningKeyEvent) bool {
			return e.ProvisioningKeyID == "digest" && e.TenantID == tenant && e.DeviceUID == "uid-1" &&
				e.Hostname == "web-01" && e.Identity == "00:1a:2b:3c:4d:5e" && e.SourceIP == "203.0.113.7" &&
				e.Ephemeral && e.ReRegistration && e.Info != nil && e.Info.Arch == "amd64"
		})).Return(nil).Once()

		s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())
		s.appendProvisioningKeyEvent(context.Background(), key, req, "uid-1", "web-01", true)

		storeMock.AssertExpectations(t)
	})

	t.Run("is best-effort: a store error never propagates", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("ProvisioningKeyEventCreate", mock.Anything, mock.Anything).Return(errors.New("boom")).Once()

		s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())
		require.NotPanics(t, func() {
			s.appendProvisioningKeyEvent(context.Background(), key, req, "uid-1", "web-01", false)
		})

		storeMock.AssertExpectations(t)
	})
}

func TestListProvisioningKeyEvents(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	const tenant = "00000000-0000-4000-0000-000000000000"

	events := []models.ProvisioningKeyEvent{{ID: "e1", ProvisioningKeyID: "hash", TenantID: tenant, Hostname: "web-01"}}

	cases := []struct {
		description    string
		req            *requests.ListProvisioningKeyEvents
		requiredMocks  func(ctx context.Context)
		expectedEvents []models.ProvisioningKeyEvent
		expectedCount  int
		expectedErr    error
	}{
		{
			description: "fails when the key does not exist",
			req:         &requests.ListProvisioningKeyEvents{TenantID: tenant, ID: "hash", Paginator: query.Paginator{Page: 1, PerPage: 10}, Sorter: query.Sorter{By: "created_at", Order: query.OrderDesc}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, "hash").
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedEvents: nil,
			expectedCount:  0,
			expectedErr:    NewErrProvisioningKeyNotFound("hash", store.ErrNoDocuments),
		},
		{
			description: "lists the key's enrollment history, defaulting the sort to created_at with an id tiebreak",
			req:         &requests.ListProvisioningKeyEvents{TenantID: tenant, ID: "hash", Paginator: query.Paginator{Page: 1, PerPage: 10}, Sorter: query.Sorter{By: "created_at", Order: query.OrderDesc}},
			requiredMocks: func(ctx context.Context) {
				key := &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, "hash").
					Return(key, nil).Once()
				queryOptionsMock.On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "id"}).Return(nil).Once()
				queryOptionsMock.On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).Return(nil).Once()
				storeMock.On("ProvisioningKeyEventList", ctx, scope.MustBounded(tenant), "hash", mock.AnythingOfType("[]store.QueryOption")).
					Return(events, 1, nil).Once()
			},
			expectedEvents: events,
			expectedCount:  1,
			expectedErr:    nil,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			result, count, err := s.ListProvisioningKeyEvents(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedCount, count)
			require.Equal(t, tc.expectedEvents, result)
		})
	}

	storeMock.AssertExpectations(t)
}
