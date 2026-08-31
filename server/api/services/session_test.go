package services

import (
	"context"
	goerrors "errors"
	"net"
	"reflect"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSessions(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	ctx := context.TODO()

	type Expected struct {
		sessions []models.Session
		count    int
		err      error
	}

	matchFilters := func(data []query.Filter) any {
		return mock.MatchedBy(func(f *query.Filters) bool {
			return reflect.DeepEqual(f.Data, data)
		})
	}

	deviceUIDFilterData := []query.Filter{
		{
			Type: query.FilterTypeProperty,
			Params: &query.FilterProperty{
				Name:     "device_uid",
				Operator: "eq",
				Value:    "uid1",
			},
		},
	}

	closedFilterData := []query.Filter{
		{
			Type: query.FilterTypeProperty,
			Params: &query.FilterProperty{
				Name:     "closed",
				Operator: "bool",
				Value:    true,
			},
		},
	}

	activeFilterData := []query.Filter{
		{
			Type: query.FilterTypeProperty,
			Params: &query.FilterProperty{
				Name:     "active",
				Operator: "bool",
				Value:    true,
			},
		},
	}

	cases := []struct {
		description   string
		sc            scope.Scope
		req           *requests.ListSessions
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.AnythingOfType("*query.Filters")).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return(nil, 0, goerrors.New("error")).Once()
			},
			expected: Expected{
				sessions: nil,
				count:    0,
				err:      goerrors.New("error"),
			},
		},
		{
			description: "succeeds",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				sessions := []models.Session{
					{UID: "uid1"},
					{UID: "uid2"},
					{UID: "uid3"},
				}
				queryOptionsMock.
					On("Match", mock.AnythingOfType("*query.Filters")).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return(sessions, len(sessions), nil).Once()
			},
			expected: Expected{
				sessions: []models.Session{
					{UID: "uid1"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				count: len([]models.Session{
					{UID: "uid1"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				err: nil,
			},
		},
		{
			description: "succeeds with device_uid filter",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{Data: deviceUIDFilterData},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", matchFilters(deviceUIDFilterData)).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Session{{UID: "uid1"}}, 1, nil).Once()
			},
			expected: Expected{
				sessions: []models.Session{{UID: "uid1"}},
				count:    1,
				err:      nil,
			},
		},
		{
			description: "succeeds with closed filter",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{Data: closedFilterData},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", matchFilters(closedFilterData)).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Session{}, 0, nil).Once()
			},
			expected: Expected{
				sessions: []models.Session{},
				count:    0,
				err:      nil,
			},
		},
		{
			description: "succeeds with active filter",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Filters:   query.Filters{Data: activeFilterData},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", matchFilters(activeFilterData)).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.Anything, mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Session{}, 0, nil).Once()
			},
			expected: Expected{
				sessions: []models.Session{},
				count:    0,
				err:      nil,
			},
		},
		{
			description: "succeeds with unbounded scope",
			sc:          scope.NewUnbounded("admin"),
			req: &requests.ListSessions{
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.AnythingOfType("*query.Filters")).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc, Tiebreak: "id"}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.MatchedBy(func(sc scope.Scope) bool {
					return !sc.IsBounded() && sc.IsValid()
				}), mock.AnythingOfType("[]store.QueryOption")).
					Return([]models.Session{{UID: "s1"}}, 1, nil).Once()
			},
			expected: Expected{
				sessions: []models.Session{{UID: "s1"}},
				count:    1,
				err:      nil,
			},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedSessions, count, err := service.ListSessions(ctx, tc.sc, tc.req)
			assert.Equal(t, tc.expected, Expected{returnedSessions, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	ctx := context.TODO()

	type Expected struct {
		session *models.Session
		err     error
	}

	boundedScope := scope.MustBounded("00000000-0000-4000-0000-000000000000")

	cases := []struct {
		name          string
		scope         scope.Scope
		uid           models.UID
		requiredMocks func(storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			name:  "fails when session is not found",
			scope: boundedScope,
			uid:   models.UID("_uid"),
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SessionResolve", ctx, boundedScope, store.SessionUIDResolver, "_uid").
					Return(nil, goerrors.New("error")).Once()
			},
			expected: Expected{
				session: nil,
				err:     NewErrSessionNotFound(models.UID("_uid"), goerrors.New("error")),
			},
		},
		{
			name:  "passes the caller's bounded scope straight to the store",
			scope: boundedScope,
			uid:   models.UID("uid"),
			requiredMocks: func(storeMock *storemock.MockStore) {
				session := &models.Session{UID: "uid", TenantID: "00000000-0000-4000-0000-000000000000"}
				storeMock.On("SessionResolve", ctx, boundedScope, store.SessionUIDResolver, "uid").
					Return(session, nil).Once()
			},
			expected: Expected{
				session: &models.Session{UID: "uid", TenantID: "00000000-0000-4000-0000-000000000000"},
				err:     nil,
			},
		},
		{
			name:  "returns not found for a session the bounded scope excludes",
			scope: scope.MustBounded("11111111-1111-4111-0000-000000000000"),
			uid:   models.UID("victim-uid"),
			requiredMocks: func(storeMock *storemock.MockStore) {
				storeMock.On("SessionResolve", ctx, scope.MustBounded("11111111-1111-4111-0000-000000000000"), store.SessionUIDResolver, "victim-uid").
					Return(nil, goerrors.New("not found")).Once()
			},
			expected: Expected{
				session: nil,
				err:     NewErrSessionNotFound(models.UID("victim-uid"), goerrors.New("not found")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			tc.requiredMocks(storeMock)

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache())
			returnedSession, err := service.GetSession(ctx, tc.scope, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
			storeMock.AssertExpectations(t)
		})
	}
}

func TestCreateSession(t *testing.T) {
	mock := storemock.NewMockStore(t)

	ctx := context.TODO()

	locator := mocksGeoIp.NewMockLocator(t)

	type Expected struct {
		session *models.Session
		err     error
	}

	req := requests.SessionCreate{UID: "uid"}
	model := models.Session{UID: "uid", Position: models.SessionPosition{
		Latitude:  0,
		Longitude: 0,
	}}

	Err := goerrors.New("error")

	cases := []struct {
		name          string
		session       requests.SessionCreate
		requiredMocks func()
		expected      Expected
	}{
		{
			name:    "fails",
			session: req,
			requiredMocks: func() {
				locator.On("GetPosition", net.ParseIP(model.IPAddress)).
					Return(geoip.Position{}, nil).Once()
				mock.On("SessionCreate", ctx, model).
					Return("", Err).Once()
			},
			expected: Expected{
				session: nil,
				err:     Err,
			},
		},
		{
			name:    "succeeds",
			session: req,
			requiredMocks: func() {
				locator.On("GetPosition", net.ParseIP(model.IPAddress)).
					Return(geoip.Position{}, nil).Once()
				mock.On("SessionCreate", ctx, model).
					Return("uid", nil).Once()
				mock.On("SessionResolve", ctx, scope.NewUnbounded("reading back the session this call just created, by its generated UID"), store.SessionUIDResolver, "uid").
					Return(&model, nil).Once()
			},
			expected: Expected{
				session: &model,
				err:     nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), WithLocator(locator))
			returnedSession, err := service.CreateSession(ctx, tc.session)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeactivateSession(t *testing.T) {
	mock := storemock.NewMockStore(t)

	ctx := context.TODO()

	cases := []struct {
		name          string
		uid           models.UID
		requiredMocks func()
		expected      error
	}{
		{
			name: "fails when session is not found",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, "_uid").
					Return(nil, goerrors.New("get error")).Once()
			},
			expected: NewErrSessionNotFound("_uid", goerrors.New("get error")),
		},
		{
			name: "fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, "_uid").
					Return(&models.Session{
						UID: "_uid",
					}, nil).Once()

				mock.On("ActiveSessionDelete", ctx, models.UID("_uid")).
					Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name: "succeeds",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, "_uid").
					Return(&models.Session{
						UID: "_uid",
					}, nil).Once()

				mock.On("ActiveSessionDelete", ctx, models.UID("_uid")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache())
			err := service.DeactivateSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateSession(t *testing.T) {
	mockStore := storemock.NewMockStore(t)
	ctx := context.Background()
	uid := models.UID("test-uid")
	updateModel := models.SessionUpdate{}
	theTrue := true
	updateModel.Authenticated = &theTrue

	sess := &models.Session{UID: string(uid)}

	cases := []struct {
		description   string
		requiredMocks func()
		expectedErr   error
	}{
		{
			description: "fails when SessionGet returns error",
			requiredMocks: func() {
				mockStore.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid)).
					Return(nil, goerrors.New("get error")).Once()
			},
			expectedErr: NewErrSessionNotFound(uid, goerrors.New("get error")),
		},
		{
			description: "fails when SessionUpdate returns error",
			requiredMocks: func() {
				mockStore.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid)).
					Return(sess, nil).Once()
				mockStore.On("ActiveSessionCreate", ctx, sess).
					Return(nil).Once()
				mockStore.On("SessionUpdate", ctx, sess).
					Return(goerrors.New("update error")).Once()
			},
			expectedErr: goerrors.New("update error"),
		},
		{
			description: "succeeds when no errors",
			requiredMocks: func() {
				mockStore.On("SessionResolve", ctx, scope.NewUnbounded(reasonInternalSessionMutation), store.SessionUIDResolver, string(uid)).
					Return(sess, nil).Once()
				mockStore.On("ActiveSessionCreate", ctx, sess).
					Return(nil).Once()
				mockStore.On("SessionUpdate", ctx, sess).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
	}

	service := NewService(store.Store(mockStore), privateKey, publicKey, storecache.NewNullCache())
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := service.UpdateSession(ctx, uid, updateModel)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	mockStore.AssertExpectations(t)
}
