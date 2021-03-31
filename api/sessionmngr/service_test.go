package sessionmngr

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListSessions(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	sessions := []models.Session{
		{UID: "uid"},
	}

	query := paginator.Query{Page: 1, PerPage: 10}

	mock.On("SessionList", ctx, query).
		Return(sessions, len(sessions), nil).Once()

	returnedSessions, count, err := s.ListSessions(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, sessions, returnedSessions)
	assert.Equal(t, count, len(sessions))

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	session := &models.Session{UID: "uid"}
	mock.On("SessionGet", ctx, models.UID(session.UID)).
		Return(session, nil).Once()

	returnedSession, err := s.GetSession(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, session, returnedSession)

	mock.AssertExpectations(t)
}

func TestCreateSession(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	session := &models.Session{UID: "uid"}
	mock.On("SessionCreate", ctx, *session).
		Return(session, nil).Once()

	returnedSession, err := s.CreateSession(ctx, *session)
	assert.NoError(t, err)
	assert.Equal(t, session, returnedSession)

	mock.AssertExpectations(t)
}

func TestDeactivateSession(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	mock.On("SessionDeleteActives", ctx, models.UID("uid")).
		Return(nil).Once()

	err := s.DeactivateSession(ctx, models.UID("uid"))
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestSetSessionAuthenticated(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	mock.On("SessionSetAuthenticated", ctx, models.UID("uid"), true).
		Return(nil).Once()

	err := s.SetSessionAuthenticated(ctx, models.UID("uid"), true)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
