package services

import (
	"context"
	"net"
	"testing"

	goerrors "errors"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSessions(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	type Expected struct {
		sessions []models.Session
		count    int
		err      error
	}

	cases := []struct {
		description   string
		req           *requests.ListSessions
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails",
			req: &requests.ListSessions{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Paginator: query.Paginator{Page: 1, PerPage: 10},
			},
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
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
					On("InNamespace", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &query.Sorter{By: "started_at", Order: query.OrderDesc}).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 10}).
					Return(nil).
					Once()
				storeMock.On("SessionList", ctx, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
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
	}

	service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedSessions, count, err := service.ListSessions(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{returnedSessions, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := new(storemock.Store)

	ctx := context.TODO()

	type Expected struct {
		session *models.Session
		err     error
	}

	cases := []struct {
		name          string
		ctx           context.Context
		uid           models.UID
		requiredMocks func()
		expected      Expected
	}{
		{
			name: "fails when session is not found",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "_uid").
					Return(nil, goerrors.New("error")).Once()
			},
			expected: Expected{
				session: nil,
				err:     NewErrSessionNotFound(models.UID("_uid"), goerrors.New("error")),
			},
		},
		{
			name: "succeeds",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				session := &models.Session{UID: "uid"}

				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "uid").
					Return(session, nil).Once()
			},
			expected: Expected{
				session: &models.Session{UID: "uid"},
				err:     nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedSession, err := service.GetSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestCreateSession(t *testing.T) {
	mock := new(storemock.Store)

	ctx := context.TODO()

	locator := &mocksGeoIp.Locator{}

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
				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "uid").
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))
			returnedSession, err := service.CreateSession(ctx, tc.session)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeactivateSession(t *testing.T) {
	mock := new(storemock.Store)

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
				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "_uid").
					Return(nil, goerrors.New("get error")).Once()
			},
			expected: NewErrSessionNotFound("_uid", goerrors.New("get error")),
		},
		{
			name: "fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "_uid").
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
				mock.On("SessionResolve", ctx, store.SessionUIDResolver, "_uid").
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.DeactivateSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateSession(t *testing.T) {
	mockStore := new(storemock.Store)
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
				mockStore.On("SessionResolve", ctx, store.SessionUIDResolver, string(uid)).
					Return(nil, goerrors.New("get error")).Once()
			},
			expectedErr: NewErrSessionNotFound(uid, goerrors.New("get error")),
		},
		{
			description: "fails when SessionUpdate returns error",
			requiredMocks: func() {
				mockStore.On("SessionResolve", ctx, store.SessionUIDResolver, string(uid)).
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
				mockStore.On("SessionResolve", ctx, store.SessionUIDResolver, string(uid)).
					Return(sess, nil).Once()
				mockStore.On("ActiveSessionCreate", ctx, sess).
					Return(nil).Once()
				mockStore.On("SessionUpdate", ctx, sess).
					Return(nil).Once()
			},
			expectedErr: nil,
		},
	}

	service := NewService(store.Store(mockStore), privateKey, publicKey, storecache.NewNullCache(), clientMock)
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			err := service.UpdateSession(ctx, uid, updateModel)
			assert.Equal(t, tc.expectedErr, err)
		})
	}

	mockStore.AssertExpectations(t)
}
