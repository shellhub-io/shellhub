package services

import (
	"context"
	"net"
	"testing"

	goerrors "errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
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
		name          string
		pagination    paginator.Query
		requiredMocks func(query paginator.Query)
		expected      Expected
	}{
		{
			name:       "ListSessions fails",
			pagination: paginator.Query{Page: 1, PerPage: 10},
			requiredMocks: func(query paginator.Query) {
				mock.On("SessionList", ctx, query).
					Return(nil, 0, goerrors.New("error")).Once()
			},
			expected: Expected{
				sessions: nil,
				count:    0,
				err:      goerrors.New("error"),
			},
		},
		{
			name:       "ListSessions succeeds",
			pagination: paginator.Query{Page: 1, PerPage: 10},
			requiredMocks: func(query paginator.Query) {
				sessions := []models.Session{
					{UID: "uid1"},
					{UID: "uid2"},
					{UID: "uid3"},
				}
				mock.On("SessionList", ctx, query).
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
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks(tc.pagination)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedSessions, count, err := service.ListSessions(ctx, tc.pagination)
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
			name: "GetSession fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionGet", ctx, models.UID("_uid")).
					Return(nil, goerrors.New("error")).Once()
			},
			expected: Expected{
				session: nil,
				err:     NewErrSessionNotFound(models.UID("_uid"), goerrors.New("error")),
			},
		},
		{
			name: "GetSession succeeds",
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
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

	Err := goerrors.New("error")

	cases := []struct {
		name          string
		session       requests.SessionCreate
		requiredMocks func()
		expected      Expected
	}{
		{
			name:    "CreateSession fails",
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
			name:    "CreateSession succeeds",
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)
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
			name: "DeactivateSession fails when session is not found",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
					Return(store.ErrNoDocuments).Once()
			},
			expected: NewErrSessionNotFound("_uid", store.ErrNoDocuments),
		},
		{
			name: "DeactivateSession fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
					Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name: "DeactivateSession succeeds",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("SessionDeleteActives", ctx, models.UID("uid")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.DeactivateSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestSetSessionAuthenticated(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		name          string
		uid           models.UID
		requiredMocks func()
		expected      error
	}{
		{
			name: "SetSessionAuthenticated fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionSetAuthenticated", ctx, models.UID("_uid"), true).
					Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name: "SetSessionAuthenticated succeeds",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("SessionSetAuthenticated", ctx, models.UID("uid"), true).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.SetSessionAuthenticated(ctx, tc.uid, true)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
