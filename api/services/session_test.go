package services

import (
	"context"
	"net"
	"testing"

	goerrors "errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
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
			name: "fails when session is not found",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
					Return(store.ErrNoDocuments).Once()
			},
			expected: NewErrSessionNotFound("_uid", store.ErrNoDocuments),
		},
		{
			name: "fails",
			uid:  models.UID("_uid"),
			requiredMocks: func() {
				mock.On("SessionDeleteActives", ctx, models.UID("_uid")).
					Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name: "succeeds",
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

func TestUpdateSession(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	theTrue := true

	cases := []struct {
		name          string
		uid           models.UID
		model         models.SessionUpdate
		requiredMocks func()
		expected      error
	}{
		{
			name:  "fails whne cannot get the sessioni",
			uid:   models.UID("_uid"),
			model: models.SessionUpdate{},
			requiredMocks: func() {
				mock.On("SessionGet", ctx, models.UID("_uid")).Return(nil, goerrors.New("error")).Once()
			},
			expected: NewErrSessionNotFound(models.UID("_uid"), goerrors.New("error")),
		},
		{
			name:  "fails to update the session",
			uid:   models.UID("_uid"),
			model: models.SessionUpdate{},
			requiredMocks: func() {
				sess := &models.Session{}

				mock.On("SessionGet", ctx, models.UID("_uid")).Return(sess, nil).Once()

				mock.On("SessionUpdate", ctx, models.UID("_uid"), sess).Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name:  "success to update the session",
			uid:   models.UID("_uid"),
			model: models.SessionUpdate{},
			requiredMocks: func() {
				sess := &models.Session{}

				mock.On("SessionGet", ctx, models.UID("_uid")).Return(sess, nil).Once()

				mock.On("SessionUpdate", ctx, models.UID("_uid"), sess).Return(nil).Once()
			},
			expected: nil,
		},
		{
			name: "fails to update the session when authenticated field is updated",
			uid:  models.UID("_uid"),
			model: models.SessionUpdate{
				Authenticated: &theTrue,
			},
			requiredMocks: func() {
				sess := &models.Session{}

				mock.On("SessionGet", ctx, models.UID("_uid")).Return(sess, nil).Once()
				mock.On("SessionUpdate", ctx, models.UID("_uid"), sess).Return(nil).Once()
				mock.On("SessionActiveCreate", ctx, models.UID("_uid"), sess).Return(goerrors.New("error")).Once()
			},
			expected: goerrors.New("error"),
		},
		{
			name: "success to update the session when authenticated field is updated",
			uid:  models.UID("_uid"),
			model: models.SessionUpdate{
				Authenticated: &theTrue,
			},
			requiredMocks: func() {
				sess := &models.Session{}

				mock.On("SessionGet", ctx, models.UID("_uid")).Return(sess, nil).Once()
				mock.On("SessionUpdate", ctx, models.UID("_uid"), sess).Return(nil).Once()
				mock.On("SessionActiveCreate", ctx, models.UID("_uid"), sess).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateSession(ctx, tc.uid, tc.model)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
