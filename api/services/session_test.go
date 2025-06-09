package services

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	mocker "github.com/stretchr/testify/mock"
)

func TestListSessions(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		sessions []models.Session
		count    int
		err      error
	}

	cases := []struct {
		description   string
		paginator     query.Paginator
		requiredMocks func(paginator query.Paginator)
		expected      Expected
	}{
		{
			description: "fails",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			requiredMocks: func(paginator query.Paginator) {
				mock.On("SessionList", ctx, paginator).
					Return(nil, 0, errors.New("error")).Once()
			},
			expected: Expected{
				sessions: nil,
				count:    0,
				err:      errors.New("error"),
			},
		},
		{
			description: "succeeds",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			requiredMocks: func(paginator query.Paginator) {
				sessions := []models.Session{
					{UID: "uid1"},
					{UID: "uid2"},
					{UID: "uid3"},
				}
				mock.On("SessionList", ctx, paginator).
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.paginator)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedSessions, count, err := service.ListSessions(ctx, tc.paginator)
			assert.Equal(t, tc.expected, Expected{returnedSessions, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := new(mocks.Store)

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
				mock.On("SessionGet", ctx, models.UID("_uid")).
					Return(nil, errors.New("error")).Once()
			},
			expected: Expected{
				session: nil,
				err:     NewErrSessionNotFound(models.UID("_uid"), errors.New("error")),
			},
		},
		{
			name: "succeeds",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				session := &models.Session{UID: "uid"}

				mock.On("SessionGet", ctx, models.UID("uid")).
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
	mock := new(mocks.Store)

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

	Err := errors.New("error")

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
					Return(nil, Err).Once()
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
	mock := new(mocks.Store)

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
				mock.On("SessionGet", ctx, models.UID("_uid")).
					Return(nil, goerrors.New("get error")).Once()
			},
			expected: NewErrSessionNotFound("_uid", goerrors.New("get error")),
		},
		{
			name: "fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionGet", ctx, models.UID("_uid")).
					Return(&models.Session{
						UID: "_uid",
					}, nil).Once()

				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
					Return(errors.New("error")).Once()
			},
			expected: errors.New("error"),
		},
		{
			name: "succeeds",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionGet", ctx, models.UID("_uid")).
					Return(&models.Session{
						UID: "_uid",
					}, nil).Once()

				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
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
	mockStore := new(mocks.Store)
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
				mockStore.On("SessionGet", ctx, uid).
					Return(nil, errors.New("get error")).Once()
			},
			expectedErr: NewErrSessionNotFound(uid, errors.New("get error")),
		},
		{
			description: "fails when SessionUpdate returns error",
			requiredMocks: func() {
				mockStore.On("SessionGet", ctx, uid).
					Return(sess, nil).Once()
				mockStore.On("SessionUpdate", ctx, uid, sess, &updateModel).
					Return(errors.New("update error")).Once()
			},
			expectedErr: errors.New("update error"),
		},
		{
			description: "succeeds when no errors",
			requiredMocks: func() {
				mockStore.On("SessionGet", ctx, uid).
					Return(sess, nil).Once()
				mockStore.On("SessionUpdate", ctx, uid, sess, &updateModel).
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

func TestListEvents(t *testing.T) {
	type Expected struct {
		events  []models.SessionEvent
		counter int
		err     error
	}

	mock := new(mocks.Store)

	tests := []struct {
		description   string
		uid           string
		paginator     query.Paginator
		sorter        query.Sorter
		filters       query.Filters
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "failed to get the session",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				nil, 0, NewErrSessionNotFound(models.UID("uid"), errors.New("error")),
			},
		},
		{
			description: "failed to list the events",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(&models.Session{
						UID: "uid",
					}, nil).
					Once()

				mock.On(
					"SessionListEvents",
					mocker.Anything,
					models.UID("uid"),
					query.Paginator{Page: 1, PerPage: 10},
					query.Filters{},
					query.Sorter{By: "timestamp", Order: "asc"},
				).
					Return(nil, 0, errors.New("error")).
					Once()
			},
			expected: Expected{
				nil, 0, errors.New("error"),
			},
		},
		{
			description: "success when session has no events",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(&models.Session{
						UID: "uid",
					}, nil).
					Once()

				mock.On(
					"SessionListEvents",
					mocker.Anything,
					models.UID("uid"),
					query.Paginator{Page: 1, PerPage: 10},
					query.Filters{},
					query.Sorter{By: "timestamp", Order: "asc"},
				).
					Return([]models.SessionEvent{}, 0, nil).
					Once()
			},
			expected: Expected{
				[]models.SessionEvent{}, 0, nil,
			},
		},
		{
			description: "success when session has one event",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(&models.Session{
						UID: "uid",
					}, nil).
					Once()

				mock.On(
					"SessionListEvents",
					mocker.Anything,
					models.UID("uid"),
					query.Paginator{Page: 1, PerPage: 10},
					query.Filters{},
					query.Sorter{By: "timestamp", Order: "asc"},
				).
					Return([]models.SessionEvent{
						{},
					}, 1, nil).
					Once()
			},
			expected: Expected{
				[]models.SessionEvent{
					{},
				}, 1, nil,
			},
		},
		{
			description: "success when session has many events",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(&models.Session{
						UID: "uid",
					}, nil).
					Once()

				mock.On(
					"SessionListEvents",
					mocker.Anything,
					models.UID("uid"),
					query.Paginator{Page: 1, PerPage: 10},
					query.Filters{},
					query.Sorter{By: "timestamp", Order: "asc"},
				).
					Return([]models.SessionEvent{
						{},
						{},
						{},
						{},
					}, 4, nil).
					Once()
			},
			expected: Expected{
				[]models.SessionEvent{
					{},
					{},
					{},
					{},
				}, 4, nil,
			},
		},
		{
			description: "success when session has many events and is paged",
			uid:         "uid",
			paginator:   query.Paginator{Page: 1, PerPage: 2},
			sorter:      query.Sorter{By: "timestamp", Order: "asc"},
			filters:     query.Filters{},
			requiredMocks: func() {
				mock.On(
					"SessionGet",
					mocker.Anything,
					models.UID("uid"),
				).
					Return(&models.Session{
						UID: "uid",
					}, nil).
					Once()

				mock.On(
					"SessionListEvents",
					mocker.Anything,
					models.UID("uid"),
					query.Paginator{Page: 1, PerPage: 2},
					query.Filters{},
					query.Sorter{By: "timestamp", Order: "asc"},
				).
					Return([]models.SessionEvent{
						{},
						{},
					}, 4, nil).
					Once()
			},
			expected: Expected{
				[]models.SessionEvent{
					{},
					{},
				}, 4, nil,
			},
		},
	}

	service := NewService(mock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			test.requiredMocks()

			events, counter, err := service.ListEventsSession(ctx, models.UID(test.uid), test.paginator, test.filters, test.sorter)
			assert.Equal(t, test.expected, Expected{
				events:  events,
				counter: counter,
				err:     err,
			})
		})
	}

	mock.AssertExpectations(t)
}
