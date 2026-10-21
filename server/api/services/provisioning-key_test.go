package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateProvisioningKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	queryOptionsMock.On("ForUpdate").Return(nil)
	forUpdate := mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 1 })
	storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Maybe()

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewFixed(now)

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

			if !slices.Equal(got.TagIDs, want.TagIDs) {
				return false
			}

			c.TagIDs = want.TagIDs

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
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, store.ErrNoDocuments),
		},
		{
			description: "returns the store's error when the namespace cannot be read",
			req:         &requests.CreateProvisioningKey{TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(nil, errors.New("error")).Once()
			},
			expectedErr: errors.New("error"),
		},
		{
			description: "creates a key that never expires when expires_in is omitted",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuid.Fix(t, generated)
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
				uuid.Fix(t, generated)
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
				uuid.Fix(t, generated)
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
				uuid.Fix(t, generated)
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "prod", forUpdate).
					Return(&models.Tag{ID: "prod-id", Name: "prod", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true, UsageLimit: 10,
					Tags: []string{"prod"}, TagIDs: []string{"prod-id"}, CreatedBy: "000000000000000000000000",
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
			description: "creates a tag the key names that the namespace lacks",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", Tags: []string{"edge"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuid.Fix(t, generated)
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "edge", forUpdate).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("TagCreate", ctx, &models.Tag{Name: "edge", TenantID: tenant}).
					Return("edge-id", nil).Once()
				storeMock.On("ProvisioningKeyCreate", ctx, matchCreate(&models.ProvisioningKey{
					ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
					Tags: []string{"edge"}, TagIDs: []string{"edge-id"}, CreatedBy: "000000000000000000000000",
				})).Return(hashedKey, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyIDResolver, hashedKey).
					Return(&models.ProvisioningKey{
						ID: hashedKey, Name: "ci", TenantID: tenant, Reusable: true,
						Tags: []string{"edge"}, TagIDs: []string{"edge-id"}, CreatedBy: "000000000000000000000000",
					}, nil).Once()
			},
			expectedKey: plain,
		},
		{
			description: "fails without creating the key when a tag it names cannot be resolved",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", Tags: []string{"edge"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuid.Fix(t, generated)
				storeMock.On("ProvisioningKeyConflicts", ctx, scope.MustBounded(tenant), &models.ProvisioningKeyConflicts{ID: hashedKey, Name: "ci"}).
					Return([]string{}, false, nil).Once()
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "edge", forUpdate).
					Return(nil, errors.New("tag error")).Once()
			},
			expectedErr: errors.New("tag error"),
		},
		{
			description: "creates a single-use key when the usage limit is one",
			req:         &requests.CreateProvisioningKey{UserID: "000000000000000000000000", TenantID: tenant, Name: "ci", UsageLimit: 1},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				uuid.Fix(t, generated)
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
				uuid.Fix(t, generated)
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

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), WithClock(clk))

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
		Return([]models.ProvisioningKey{{Name: "ci", TenantID: tenant}}, int64(1), nil).Once()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	keys, count, err := s.ListProvisioningKeys(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Equal(t, []models.ProvisioningKey{{Name: "ci", TenantID: tenant}}, keys)

	storeMock.AssertExpectations(t)
}

func TestUpdateProvisioningKey(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Maybe()
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	queryOptionsMock.On("ForUpdate").Return(nil)
	forUpdate := mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 1 })

	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewFixed(now)

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
					Return(nil, store.ErrNoDocuments).Once()
			},
			expectedErr: NewErrNamespaceNotFound(tenant, store.ErrNoDocuments),
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
			description: "rejects changing the legacy key's enrollment mode",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Mode: &modeAutomatic},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyForbidden(),
		},
		{
			description: "rejects disabling the legacy key",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "legacy", Disabled: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "legacy").
					Return(&models.ProvisioningKey{ID: "hash", Name: "legacy", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy, Reusable: true, Mode: models.ProvisioningKeyModeManual}, nil).Once()
			},
			expectedErr: NewErrProvisioningKeyForbidden(),
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
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "prod", forUpdate).
					Return(&models.Tag{ID: "prod-id", Name: "prod", TenantID: tenant}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "runners", TenantID: tenant, Tags: []string{"prod"}, TagIDs: []string{"prod-id"}}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "leaves the tags the key holds when the request names none",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Disabled: &truePtr},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Tags: []string{"prod"}, TagIDs: []string{"prod-id"}}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Tags: []string{"prod"}, Disabled: true}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "drops every tag the key holds when the request names an empty list",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Tags: []string{}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Tags: []string{"prod"}, TagIDs: []string{"prod-id"}}, nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Tags: []string{}, TagIDs: []string{}}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "creates a tag the namespace lacks",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Tags: []string{"edge"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "edge", forUpdate).
					Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("TagCreate", ctx, &models.Tag{Name: "edge", TenantID: tenant}).
					Return("edge-id", nil).Once()
				storeMock.On("ProvisioningKeyUpdate", ctx, &models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant, Tags: []string{"edge"}, TagIDs: []string{"edge-id"}}).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			description: "fails without updating the key when a tag it names cannot be resolved",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Tags: []string{"edge"}},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).
					Return(namespace, nil).Once()
				storeMock.On("ProvisioningKeyResolve", ctx, mock.Anything, store.ProvisioningKeyNameResolver, "ci").
					Return(&models.ProvisioningKey{ID: "hash", Name: "ci", TenantID: tenant}, nil).Once()
				storeMock.On("TagResolve", ctx, scope.MustBounded(tenant), store.TagNameResolver, "edge", forUpdate).
					Return(nil, errors.New("tag error")).Once()
			},
			expectedErr: errors.New("tag error"),
		},
		{
			description: "resolves no tag when another field is rejected",
			req:         &requests.UpdateProvisioningKey{TenantID: tenant, CurrentName: "ci", Tags: []string{"edge"}, UsageLimit: &limitTwo},
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

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), WithClock(clk))

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
		expectedCount  int64
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
					Return(events, int64(1), nil).Once()
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

func TestListTenantRegistrations(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	const tenant = "00000000-0000-4000-0000-000000000000"

	events := []models.ProvisioningKeyEvent{{ID: "e1", ProvisioningKeyID: "tenant-digest", TenantID: tenant, Hostname: "web-01"}}
	errStore := errors.New("store failure")
	page := func() *requests.ListTenantRegistrations {
		return &requests.ListTenantRegistrations{TenantID: tenant, Paginator: query.Paginator{Page: 1, PerPage: 10}}
	}

	cases := []struct {
		description    string
		req            *requests.ListTenantRegistrations
		requiredMocks  func(ctx context.Context)
		expectedEvents []models.ProvisioningKeyEvent
		expectedCount  int64
		expectedErr    error
	}{
		{
			description:    "fails for an empty tenant",
			req:            &requests.ListTenantRegistrations{},
			requiredMocks:  func(context.Context) {},
			expectedEvents: nil,
			expectedErr:    ErrForbidden,
		},
		{
			description: "answers an empty page when the namespace has no tenant-only key",
			req:         page(),
			requiredMocks: func(ctx context.Context) {
				storeMock.On("ProvisioningKeyResolveSystem", ctx, scope.MustBounded(tenant)).Return(nil, store.ErrNoDocuments).Once()
			},
			expectedEvents: []models.ProvisioningKeyEvent{},
		},
		{
			description: "fails with the store's error when the tenant-only key cannot be read",
			req:         page(),
			requiredMocks: func(ctx context.Context) {
				storeMock.On("ProvisioningKeyResolveSystem", ctx, scope.MustBounded(tenant)).Return(nil, errStore).Once()
			},
			expectedEvents: nil,
			expectedErr:    errStore,
		},
		{
			description: "lists the tenant-only key's history, sorted by creation with an id tiebreak by default",
			req:         page(),
			requiredMocks: func(ctx context.Context) {
				storeMock.On("ProvisioningKeyResolveSystem", ctx, scope.MustBounded(tenant)).
					Return(&models.ProvisioningKey{ID: "tenant-digest", TenantID: tenant, Type: models.ProvisioningKeyTypeLegacy}, nil).Once()
				queryOptionsMock.On("Sort", &query.Sorter{By: "created_at", Tiebreak: "id"}).Return(nil).Once()
				queryOptionsMock.On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).Return(nil).Once()
				storeMock.On("ProvisioningKeyEventList", ctx, scope.MustBounded(tenant), "tenant-digest", mock.AnythingOfType("[]store.QueryOption")).
					Return(events, int64(1), nil).Once()
			},
			expectedEvents: events,
			expectedCount:  1,
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(storeMock, privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			result, count, err := s.ListTenantRegistrations(ctx, tc.req)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedCount, count)
			require.Equal(t, tc.expectedEvents, result)
		})
	}

	storeMock.AssertExpectations(t)
}
