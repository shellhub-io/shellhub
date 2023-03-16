package services

import (
	"context"
	"errors"
	"net"
	"testing"

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
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	sessions := []models.Session{
		{UID: "uid1"},
		{UID: "uid2"},
		{UID: "uid3"},
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	Err := errors.New("error")

	type Expected struct {
		sessions []models.Session
		count    int
		err      error
	}

	cases := []struct {
		name          string
		pagination    paginator.Query
		requiredMocks func()
		expected      Expected
	}{
		{
			name:       "ListSessions fails",
			pagination: query,
			requiredMocks: func() {
				mock.On("SessionList", ctx, query).
					Return(nil, 0, Err).Once()
			},
			expected: Expected{
				sessions: nil,
				count:    0,
				err:      Err,
			},
		},
		{
			name:       "ListSessions succeeds",
			pagination: query,
			requiredMocks: func() {
				mock.On("SessionList", ctx, query).
					Return(sessions, len(sessions), nil).Once()
			},
			expected: Expected{
				sessions: sessions,
				count:    len(sessions),
				err:      nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedSessions, count, err := s.ListSessions(ctx, tc.pagination)
			assert.Equal(t, tc.expected, Expected{returnedSessions, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	type Expected struct {
		session *models.Session
		err     error
	}

	session := &models.Session{UID: "uid"}

	Err := errors.New("error")

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
					Return(nil, Err).Once()
			},
			expected: Expected{
				session: nil,
				err:     Err,
			},
		},
		{
			name: "GetSession succeeds",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("SessionGet", ctx, models.UID("uid")).
					Return(session, nil).Once()
			},
			expected: Expected{
				session: session,
				err:     nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedSession, err := s.GetSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestCreateSession(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

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
			returnedSession, err := s.CreateSession(ctx, tc.session)
			assert.Equal(t, tc.expected, Expected{returnedSession, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeactivateSession(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("error")

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
					Return(Err).Once()
			},
			expected: Err,
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
			err := s.DeactivateSession(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestSetSessionAuthenticated(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("error")

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
					Return(Err).Once()
			},
			expected: Err,
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
			err := s.SetSessionAuthenticated(ctx, tc.uid, true)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
