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

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateInstallKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clockMock := clockmock.NewMockClock(t)
	prevClock := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	defer func() { clock.DefaultBackend = prevClock }()
	clockMock.On("Now").Return(now).Maybe()

	// Subtests below swap the global uuid backend for a mock; restore it so the
	// mock does not leak into other (parallel) tests that call uuid.Generate.
	prevUUID := uuid.DefaultBackend
	defer func() { uuid.DefaultBackend = prevUUID }()

	const tenant = "00000000-0000-4000-0000-000000000000"

	namespace := &models.Namespace{Name: "namespace", Owner: "000000000000000000000000", TenantID: tenant}
	generated := "1e7b0f4b-aca4-48eb-a353-7469f00665ed"
	plain := generated
	keySum := sha256.Sum256([]byte(plain))
	hashedKey := hex.EncodeToString(keySum[:])

	future := now.AddDate(0, 0, 30)
	past := now.Add(-time.Hour)

	// InstallKeyCreate receives a struct whose KeyEncrypted has a random nonce, so match on the
	// deterministic fields and assert the ciphertext and hint were populated.
	matchCreate := func(want *models.InstallKey) any {
		return mock.MatchedBy(func(got *models.InstallKey) bool {
			if got.KeyEncrypted == "" || got.KeyHint != installKeyHint(generated) {
				return false
			}

			// These cases create default-mode keys: assert the mode defaults to automatic with an empty
			// allowlist, then clear them so the remaining fields can be compared against `want`.
			if got.Mode != models.InstallKeyModeAutomatic || len(got.AllowedMACs) != 0 {
				return false
			}

			c := *got
			c.KeyEncrypted = ""
			c.KeyHint = ""
			c.Mode = ""
			c.AllowedMACs = nil

			return reflect.DeepEqual(&c, want)
		})
	}

	cases := []struct {
		description   string
		req           *requests.CreateInstallKey
		requiredMocks func(ctx context.Context)
		expectedKey   string
		expectedErr   error
	}{
		{
			description: "fails when namespace does not exist",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(nil, errors.New("error")).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, errors.New("error")),
		},
		{
			description: "fails when the expiration is in the past",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci", ExpiresAt: &past},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrBadRequest(errors.New("expires_at must be a future date")),
		},
		{
			description: "fails when webhook mode has no http(s) url",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci", Mode: "webhook", WebhookSecret: "s"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrInstallKeyInvalidField(map[string]string{"webhook_url": "must be an http or https URL"}),
		},
		{
			description: "fails when webhook mode has no secret",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci", Mode: "webhook", WebhookURL: "https://hook.example/enroll"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrInstallKeyInvalidField(map[string]string{"webhook_secret": "is required for webhook mode"}),
		},
		{
			description: "fails when allowlist mode has no MAC",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci", Mode: "allowlist"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
			},
			expectedErr: NewErrInstallKeyInvalidField(map[string]string{"allowed_macs": "at least one MAC is required for allowlist mode"}),
		},
		{
			description: "fails when the name is duplicated",
			req:         &requests.CreateInstallKey{TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{"name"}, true, nil).Once()
			},
			expectedErr: NewErrInstallKeyDuplicated([]string{"name"}),
		},
		{
			description: "fails when the created key cannot be resolved back",
			req:         &requests.CreateInstallKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("InstallKeyCreate", ctx, matchCreate(&models.InstallKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, hashedKey).
					Return(nil, errors.New("resolve error")).Once()
			},
			expectedErr: errors.New("resolve error"),
		},
		{
			description: "derives a reusable key from a usage limit above one",
			req:         &requests.CreateInstallKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", UsageLimit: 10, Tags: []string{"prod"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("InstallKeyCreate", ctx, matchCreate(&models.InstallKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true, UsageLimit: 10,
					Tags: []string{"prod"}, CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, hashedKey).
					Return(&models.InstallKey{
						ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true, UsageLimit: 10,
						Tags: []string{"prod"}, CreatedBy: "000000000000000000000000",
					}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "creates a single-use key when the usage limit is one",
			req:         &requests.CreateInstallKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", UsageLimit: 1},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("InstallKeyCreate", ctx, matchCreate(&models.InstallKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: false, UsageLimit: 1,
					CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, hashedKey).
					Return(&models.InstallKey{
						ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: false, UsageLimit: 1,
						CreatedBy: "000000000000000000000000",
					}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "stores the provided expiration date",
			req:         &requests.CreateInstallKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", ExpiresAt: &future},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuidMock := uuidmock.NewMockUUID(t)
				uuid.DefaultBackend = uuidMock
				uuidMock.On("Generate").Return(generated).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("InstallKeyCreate", ctx, matchCreate(&models.InstallKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					ExpiresAt: &future, CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, hashedKey).
					Return(&models.InstallKey{ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true}, nil).Once()
			},
			expectedKey: plain,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.CreateInstallKey(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
			if tc.expectedErr == nil {
				require.NotNil(t, res)
				require.Equal(t, tc.expectedKey, res.Key)
			}
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListInstallKeys(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)

	const tenant = "00000000-0000-4000-0000-000000000000"

	req := &requests.ListInstallKey{
		TenantID:  tenant,
		Paginator: query.Paginator{Page: 1, PerPage: 10},
		Sorter:    query.Sorter{By: "created_at", Order: query.OrderDesc},
	}

	queryOptionsMock.On("InNamespace", tenant).Return(nil).Once()
	queryOptionsMock.On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "key_digest"}).Return(nil).Once()
	queryOptionsMock.On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).Return(nil).Once()
	storeMock.On("InstallKeyList", mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
		Return([]models.InstallKey{{Name: "ci", TenantID: tenant}}, 1, nil).Once()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)

	keys, count, err := s.ListInstallKeys(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, []models.InstallKey{{Name: "ci", TenantID: tenant}}, keys)

	storeMock.AssertExpectations(t)
}

func TestUpdateInstallKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)
	queryOptionsMock.On("InNamespace", mock.Anything).Return(nil)

	const tenant = "00000000-0000-4000-0000-000000000000"
	namespace := &models.Namespace{Name: "namespace", TenantID: tenant}
	truePtr := true
	falsePtr := false
	limitTwo := 2
	limitUnlimited := 0
	ephemeralTimeout5 := 5
	modeAutomatic := "automatic"

	cases := []struct {
		description   string
		req           *requests.UpdateInstallKey
		requiredMocks func(ctx context.Context)
		expectedErr   error
	}{
		{
			description: "fails when namespace does not exist",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(nil, errors.New("error")).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, errors.New("error")),
		},
		{
			description: "fails when the key does not exist",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: NewErrInstallKeyNotFound("ci", store.ErrNoDocuments),
		},
		{
			description: "fails when the new name is duplicated",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Name: "runners"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{Name: "runners"}).
					Return([]string{"name"}, true, nil).Once()
			},
			expectedErr: NewErrInstallKeyDuplicated([]string{"name"}),
		},
		{
			description: "revokes the key",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Revoked: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Reusable: true}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Reusable: true, Revoked: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "changes the legacy key's enrollment mode",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "legacy", Mode: &modeAutomatic},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "legacy", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeManual}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeAutomatic}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "allows disabling the legacy key to turn off keyless enrollment",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "legacy", Disabled: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "legacy", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeManual}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeManual, Disabled: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects changing a fixed field (name/limit/tags) on the legacy key",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "legacy", Name: "renamed"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "legacy", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrInstallKeyForbidden(),
		},
		{
			description: "fails when lowering the usage limit below the used count",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", UsageLimit: &limitTwo},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 5, UsedTimes: 3}, nil).Once()
			},
			expectedErr: NewErrInstallKeyInvalidField(map[string]string{
				"usage_limit": "cannot be lower than the number of times the key was already used",
			}),
		},
		{
			description: "allows setting the usage limit to unlimited regardless of used count",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", UsageLimit: &limitUnlimited},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 5, UsedTimes: 3}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, UsageLimit: 0, UsedTimes: 3, Reusable: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "renames and retags the key",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Name: "runners", Tags: []string{"prod"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("InstallKeyConflicts", ctx, tenant, &models.InstallKeyConflicts{Name: "runners"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "runners", TenantID: tenant, Tags: []string{"prod"}}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "turns on ephemeral with a timeout",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &truePtr, EphemeralTimeout: &ephemeralTimeout5},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 5}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "defaults the timeout to the max when ephemeral is turned on without one",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				// EphemeralTimeout must land at the max (10), never 0 — a 0 timeout would delete devices the
				// moment they disconnect.
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 10}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "clears the timeout when ephemeral is turned off",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "ci", Ephemeral: &falsePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "ci", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: true, EphemeralTimeout: 10}, nil).Once()
				storeMock.On("InstallKeyUpdate", ctx, &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant, Ephemeral: false, EphemeralTimeout: 0}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "rejects changing ephemeral on the legacy key",
			req:         &requests.UpdateInstallKey{TenantID: tenant, CurrentName: "legacy", Ephemeral: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyNameResolver, "legacy", mock.AnythingOfType("[]store.QueryOption")).
					Return(&models.InstallKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true, Mode: models.InstallKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrInstallKeyForbidden(),
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			err := s.UpdateInstallKey(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestAppendInstallKeyEvent(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	req := requests.DeviceAuth{
		TenantID: tenant,
		Identity: &requests.DeviceIdentity{MAC: "00:1a:2b:3c:4d:5e"},
		Info:     &requests.DeviceInfo{ID: "debian", PrettyName: "Debian GNU/Linux 12", Version: "v0.18.0", Arch: "amd64", Platform: "docker"},
		RealIP:   "203.0.113.7",
	}
	key := &models.InstallKey{ID: "digest", Name: "ci", TenantID: tenant, Ephemeral: true}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("records the enrollment with the denormalized device facts", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("InstallKeyEventCreate", mock.Anything, mock.MatchedBy(func(e *models.InstallKeyEvent) bool {
			return e.InstallKeyID == "digest" && e.TenantID == tenant && e.DeviceUID == "uid-1" &&
				e.Hostname == "web-01" && e.MAC == "00:1a:2b:3c:4d:5e" && e.SourceIP == "203.0.113.7" &&
				e.Ephemeral && e.ReRegistration && e.Info != nil && e.Info.Arch == "amd64"
		})).Return(nil).Once()

		s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)
		s.appendInstallKeyEvent(context.Background(), key, req, "uid-1", "web-01", true)

		storeMock.AssertExpectations(t)
	})

	t.Run("is best-effort: a store error never propagates", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("InstallKeyEventCreate", mock.Anything, mock.Anything).Return(errors.New("boom")).Once()

		s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)
		require.NotPanics(t, func() {
			s.appendInstallKeyEvent(context.Background(), key, req, "uid-1", "web-01", false)
		})

		storeMock.AssertExpectations(t)
	})
}

func TestListInstallKeyEvents(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)
	queryOptionsMock.On("InNamespace", mock.Anything).Return(nil)

	const tenant = "00000000-0000-4000-0000-000000000000"

	events := []models.InstallKeyEvent{{ID: "e1", InstallKeyID: "hash", TenantID: tenant, Hostname: "web-01"}}

	cases := []struct {
		description    string
		req            *requests.ListInstallKeyEvents
		requiredMocks  func(ctx context.Context)
		expectedEvents []models.InstallKeyEvent
		expectedCount  int
		expectedErr    error
	}{
		{
			description: "fails when the key does not exist",
			req:         &requests.ListInstallKeyEvents{TenantID: tenant, ID: "hash", Paginator: query.Paginator{Page: 1, PerPage: 10}, Sorter: query.Sorter{By: "created_at", Order: query.OrderDesc}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, "hash", mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedEvents: nil,
			expectedCount:  0,
			expectedErr:    NewErrInstallKeyNotFound("hash", store.ErrNoDocuments),
		},
		{
			description: "lists the key's enrollment history, defaulting the sort to created_at with an id tiebreak",
			req:         &requests.ListInstallKeyEvents{TenantID: tenant, ID: "hash", Paginator: query.Paginator{Page: 1, PerPage: 10}, Sorter: query.Sorter{By: "created_at", Order: query.OrderDesc}},
			requiredMocks: func(ctx context.Context) {
				key := &models.InstallKey{ID: "hash", Name: "ci", TenantID: tenant}
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, "hash", mock.AnythingOfType("[]store.QueryOption")).
					Return(key, nil).Once()
				queryOptionsMock.On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "id"}).Return(nil).Once()
				queryOptionsMock.On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).Return(nil).Once()
				storeMock.On("InstallKeyEventList", ctx, tenant, "hash", mock.AnythingOfType("[]store.QueryOption")).
					Return(events, 1, nil).Once()
			},
			expectedEvents: events,
			expectedCount:  1,
			expectedErr:    nil,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			result, count, err := s.ListInstallKeyEvents(ctx, tc.req)
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedCount, count)
			require.Equal(t, tc.expectedEvents, result)
		})
	}

	storeMock.AssertExpectations(t)
}
