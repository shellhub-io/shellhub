package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTag(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	ctx := context.TODO()

	type Expected struct {
		insertedID string
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.CreateTag
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				insertedID: "",
				err:        NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag name conflicts",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "production"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: Expected{
				insertedID: "",
				err:        NewErrTagDuplicated([]string{"name"}, nil),
			},
		},
		{
			description: "fails when tag create fails",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "production"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagCreate", ctx, &models.Tag{Name: "production", TenantID: "tenant1"}).
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{
				insertedID: "",
				err:        errors.New("error"),
			},
		},
		{
			description: "succeeds creating tag",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "production"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagCreate", ctx, &models.Tag{Name: "production", TenantID: "tenant1"}).
					Return("000000000000000000000000", nil).
					Once()
			},
			expected: Expected{
				insertedID: "000000000000000000000000",
				err:        nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			insertedID, err := service.CreateTag(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{insertedID, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_PushTagTo(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	ctx := context.TODO()

	cases := []struct {
		description   string
		target        store.TagTarget
		req           *requests.PushTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			target:      store.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			target:      store.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when tag push fails",
			target:      store.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				storeMock.
					On("TagPushToTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", store.TagTargetDevice, "device_00000000-0000-4000-0000-000000000000").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds pushing tag",
			target:      store.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				storeMock.
					On("TagPushToTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", store.TagTargetDevice, "device_00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.PushTagTo(ctx, tc.target, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_PullTagFrom(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	ctx := context.TODO()

	cases := []struct {
		description   string
		target        store.TagTarget
		req           *requests.PullTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			target:      store.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			target:      store.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when tag pull fails",
			target:      store.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				storeMock.
					On("TagPullFromTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", store.TagTargetDevice, []string{"device_00000000-0000-4000-0000-000000000000"}).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds pulling tag",
			target:      store.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "device_00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000"}, nil).
					Once()
				storeMock.
					On("TagPullFromTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", store.TagTargetDevice, []string{"device_00000000-0000-4000-0000-000000000000"}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.PullTagFrom(ctx, tc.target, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_ListTags(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	ctx := context.TODO()

	type Expected struct {
		tags       []models.Tag
		totalCount int64
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.ListTags
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.ListTags{
				TenantID: "tenant1",
				Paginator: query.Paginator{
					Page:    1,
					PerPage: 10,
				},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{},
				totalCount: 0,
				err:        NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag list fails",
			req: &requests.ListTags{
				TenantID:  "tenant1",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				queryOptionsMock.
					On("Match", &query.Filters{}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("TagList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, int64(0), errors.New("error")).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{},
				totalCount: 0,
				err:        errors.New("error"),
			},
		},
		{
			description: "succeeds listing tags",
			req: &requests.ListTags{
				TenantID: "tenant1",
				Paginator: query.Paginator{
					Page:    1,
					PerPage: 10,
				},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				queryOptionsMock.
					On("Match", &query.Filters{}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "created_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.
					On("TagList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Tag{{Name: "production", TenantID: "tenant1"}}, int64(1), nil).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{{Name: "production", TenantID: "tenant1"}},
				totalCount: 1,
				err:        nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			tags, count, err := service.ListTags(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{tags, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_UpdateTag(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	ctx := context.TODO()

	type Expected struct {
		tag *models.Tag
		err error
	}

	cases := []struct {
		description   string
		req           *requests.UpdateTag
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				err: NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag not found",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				err: NewErrTagNotFound("production", errors.New("error")),
			},
		},
		{
			description: "fails when new name conflicts",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "staging"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: Expected{
				err: NewErrTagDuplicated([]string{"name"}, nil),
			},
		},
		{
			description: "reports a conflict lookup failure as itself, not as a conflict",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(&models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "staging"}).
					Return(nil, false, errors.New("error")).
					Once()
			},
			expected: Expected{
				err: errors.New("error"),
			},
		},
		{
			description: "fails when tag update fails",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}
				updatedTag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "staging"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(tag, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "staging"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagUpdate", ctx, updatedTag).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				err: errors.New("error"),
			},
		},
		{
			description: "fails when reading back the renamed tag fails",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}
				updatedTag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "staging"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(tag, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "staging"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagUpdate", ctx, updatedTag).
					Return(nil).
					Once()
				storeMock.
					On("TagResolve", ctx, scope.MustBounded("tenant1"), store.TagIDResolver, "tag_00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				err: errors.New("error"),
			},
		},
		{
			description: "succeeds returning the tag as stored after the rename",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}
				updatedTag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "staging"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production").
					Return(tag, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, scope.MustBounded("tenant1"), &models.TagConflicts{Name: "staging"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagUpdate", ctx, updatedTag).
					Return(nil).
					Once()
				storeMock.
					On("TagResolve", ctx, scope.MustBounded("tenant1"), store.TagIDResolver, "tag_00000000-0000-4000-0000-000000000000").
					Return(&models.Tag{
						ID:        "tag_00000000-0000-4000-0000-000000000000",
						TenantID:  "tenant1",
						Name:      "staging",
						CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
					}, nil).
					Once()
			},
			expected: Expected{
				tag: &models.Tag{
					ID:        "tag_00000000-0000-4000-0000-000000000000",
					TenantID:  "tenant1",
					Name:      "staging",
					CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
				},
				err: nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	storeMock.
		On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Times(len(cases))

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			tag, err := service.UpdateTag(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{tag, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_DeleteTag(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	queryOptionsMock.On("ForUpdate").Return(nil).Maybe()

	ctx := context.TODO()

	cases := []struct {
		description   string
		req           *requests.DeleteTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when a public key filter holds the tag",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{
						{PublicKeyFields: models.PublicKeyFields{Name: "deploy", Filter: models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{"tag_other", tag.ID}}}}},
						{PublicKeyFields: models.PublicKeyFields{Name: "untagged"}},
					}, int64(2), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{}, int64(0), nil).
					Once()
			},
			expected: NewErrTagInUse([]string{`public key "deploy"`}),
		},
		{
			description: "fails when access policy filters hold the tag",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{}, int64(0), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{
						{Name: "operators", Filter: models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{tag.ID}}}},
						{Name: "admins", Filter: models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{tag.ID}}}},
					}, int64(2), nil).
					Once()
			},
			expected: NewErrTagInUse([]string{`access policy "operators"`, `access policy "admins"`}),
		},
		{
			description: "fails when the public keys cannot be listed",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return(nil, int64(0), errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when the access policies cannot be listed",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()
				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{}, int64(0), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return(nil, int64(0), errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when tag pull fails",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()

				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{{PublicKeyFields: models.PublicKeyFields{Name: "untagged"}}}, int64(1), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{{Name: "untagged"}}, int64(1), nil).
					Once()

				for _, target := range store.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", target).
						Return(errors.New("error")).
						Once()

					break
				}
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when tag delete fails",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()

				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{{PublicKeyFields: models.PublicKeyFields{Name: "untagged"}}}, int64(1), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{{Name: "untagged"}}, int64(1), nil).
					Once()

				for _, target := range store.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", target).
						Return(nil).
						Once()
				}

				storeMock.
					On("TagDelete", ctx, tag).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				tag := &models.Tag{ID: "tag_00000000-0000-4000-0000-000000000000", Name: "production"}

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagResolve", ctx, mock.Anything, store.TagNameResolver, "production", mock.AnythingOfType("[]store.QueryOption")).
					Return(tag, nil).
					Once()

				storeMock.
					On("PublicKeyList", ctx, mock.Anything).
					Return([]models.PublicKey{{PublicKeyFields: models.PublicKeyFields{Name: "untagged"}}}, int64(1), nil).
					Once()
				storeMock.
					On("AccessPolicyList", ctx, mock.Anything).
					Return([]models.AccessPolicy{{Name: "untagged"}}, int64(1), nil).
					Once()

				for _, target := range store.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tag_00000000-0000-4000-0000-000000000000", target).
						Return(nil).
						Once()
				}

				storeMock.
					On("TagDelete", ctx, tag).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil)

	storeMock.
		On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Times(len(cases))

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.DeleteTag(ctx, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListTags(t *testing.T) {
	t.Run("TagFilterFields rejects any field and operator", func(t *testing.T) {
		assert.False(t, TagFilterFields.Allows("name", "eq"))
		assert.False(t, TagFilterFields.Allows("name", "contains"))
		assert.False(t, TagFilterFields.Allows("unknown", "eq"))
	})

	t.Run("TagSortFields allows name", func(t *testing.T) {
		assert.True(t, TagSortFields.Allows("name"))
	})

	t.Run("TagSortFields allows created_at", func(t *testing.T) {
		assert.True(t, TagSortFields.Allows("created_at"))
	})

	t.Run("TagSortFields allows updated_at", func(t *testing.T) {
		assert.True(t, TagSortFields.Allows("updated_at"))
	})

	t.Run("TagSortFields rejects unknown field", func(t *testing.T) {
		assert.False(t, TagSortFields.Allows("unknown"))
	})
}
