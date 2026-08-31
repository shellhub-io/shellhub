package storetest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionList tests the SessionList method across all implementations
func (s *Suite) TestSessionList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no sessions are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))

		require.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, 0, count)
	})

	t.Run("succeeds when sessions are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateSession(t, WithSessionUser("user1"))
		s.CreateSession(t, WithSessionUser("user2"))
		s.CreateSession(t, WithSessionUser("user3"))
		s.CreateSession(t, WithSessionUser("user4"))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))

		require.NoError(t, err)
		assert.Equal(t, 4, count)
		assert.Len(t, sessions, 4)
	})

	t.Run("returns all sessions across tenants without filter", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)

		device1 := s.CreateDevice(t, WithTenantID(tenant1))
		device2 := s.CreateDevice(t, WithTenantID(tenant2))

		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user1"))
		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user2"))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionUser("user3"))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics))
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, sessions, 3)
	})

	t.Run("succeeds when tenant filter applied", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)

		device1 := s.CreateDevice(t, WithTenantID(tenant1))
		device2 := s.CreateDevice(t, WithTenantID(tenant2))

		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user1"))
		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user2"))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionUser("user3"))

		sessions, count, err := st.SessionList(ctx, scope.MustBounded(tenant1))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, sessions, 2)

		for _, session := range sessions {
			assert.Equal(t, tenant1, session.TenantID)
		}
	})

	t.Run("returns no sessions from other tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)

		device1 := s.CreateDevice(t, WithTenantID(tenant1))
		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user1"))
		s.CreateSession(t, WithSessionDevice(device1), WithSessionUser("user2"))

		sessions, count, err := st.SessionList(ctx, scope.MustBounded(tenant2))
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, sessions)
	})

	t.Run("succeeds with pagination", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant := s.CreateNamespace(t)
		device := s.CreateDevice(t, WithTenantID(tenant))

		for i := range 5 {
			s.CreateSession(t, WithSessionDevice(device),
				WithSessionUser(fmt.Sprintf("user%d", i)))
		}

		sessions, count, err := st.SessionList(ctx, scope.MustBounded(tenant),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 2}))
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Len(t, sessions, 2)

		sessions, count, err = st.SessionList(ctx, scope.MustBounded(tenant),
			st.Options().Paginate(&query.Paginator{Page: 3, PerPage: 2}))
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Len(t, sessions, 1)
	})

	t.Run("filters by device_uid eq", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		sessA := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		sessB := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		require.NoError(t, st.ActiveSessionDelete(ctx, sessB)) // closes sessB

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		_ = sessA

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "device_uid",
					Operator: "eq",
					Value:    string(device1),
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, sessions, 2)
		for _, s := range sessions {
			assert.Equal(t, device1, s.DeviceUID)
		}
	})

	t.Run("filters by device_uid ne", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "device_uid",
					Operator: "ne",
					Value:    string(device1),
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, sessions, 2)

		for _, s := range sessions {
			assert.Equal(t, device2, s.DeviceUID)
		}
	})

	t.Run("filters by closed bool true", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		sessB := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		require.NoError(t, st.ActiveSessionDelete(ctx, sessB))

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "closed",
					Operator: "bool",
					Value:    true,
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, sessions, 1)
		for _, s := range sessions {
			assert.True(t, s.Closed)
		}
	})

	t.Run("filters by closed bool false", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		sessB := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		require.NoError(t, st.ActiveSessionDelete(ctx, sessB))

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "closed",
					Operator: "bool",
					Value:    false,
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, sessions, 3)
		for _, s := range sessions {
			assert.False(t, s.Closed)
		}
	})

	t.Run("filters by active bool true", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		sessB := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		require.NoError(t, st.ActiveSessionDelete(ctx, sessB))

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "active",
					Operator: "bool",
					Value:    true,
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Positive(t, count)
		assert.NotEmpty(t, sessions)
		assert.Equal(t, 2, count)
		assert.Len(t, sessions, 2)
		for _, s := range sessions {
			assert.True(t, s.Active)
		}
	})

	t.Run("filters by active bool false", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		device1 := s.CreateDevice(t)
		device2 := s.CreateDevice(t)

		s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		sessB := s.CreateSession(t, WithSessionDevice(device1), WithSessionActive(true))
		require.NoError(t, st.ActiveSessionDelete(ctx, sessB))

		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device2), WithSessionActive(false))

		sessions, count, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{Data: []query.Filter{
				{Type: query.FilterTypeProperty, Params: &query.FilterProperty{
					Name:     "active",
					Operator: "bool",
					Value:    false,
				}},
			}}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Positive(t, count)
		assert.NotEmpty(t, sessions)
		assert.Equal(t, 2, count)
		assert.Len(t, sessions, 2)
		for _, s := range sessions {
			assert.False(t, s.Active)
		}
	})
}

// TestSessionResolve tests session resolution by UID
func (s *Suite) TestSessionResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when session not found by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		session, err := st.SessionResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.SessionUIDResolver, "nonexistent")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, session)
	})

	t.Run("succeeds resolving session by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		sessionUID := s.CreateSession(t,
			WithSessionDevice(deviceUID),
			WithSessionUser("testuser"),
		)

		session, err := st.SessionResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.SessionUIDResolver, string(sessionUID))
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, string(sessionUID), session.UID)
		assert.Equal(t, tenantID, session.TenantID)
	})
}

// TestSessionCreate tests session creation
func (s *Suite) TestSessionCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		session := models.Session{
			Username:      "testuser",
			UID:           "test-session-uid",
			TenantID:      tenantID,
			DeviceUID:     deviceUID,
			IPAddress:     "192.168.1.1",
			Authenticated: true,
		}

		uid, err := st.SessionCreate(ctx, session)
		require.NoError(t, err)
		assert.NotEmpty(t, uid)

		created, err := st.SessionResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.SessionUIDResolver, uid)
		require.NoError(t, err)
		assert.Equal(t, tenantID, created.TenantID)
	})
}

// TestSessionUpdateDeviceUID tests updating session device UID
func (s *Suite) TestSessionUpdateDeviceUID(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when device is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		oldUID := models.UID("0000000000000000000000000000000000000000000000000000000000000000")
		newUID := models.UID("1111111111111111111111111111111111111111111111111111111111111111")

		err := st.SessionUpdateDeviceUID(ctx, oldUID, newUID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when device is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		oldDevice := s.CreateDevice(t, WithDeviceName("old-device"), WithTenantID(tenantID))
		newDevice := s.CreateDevice(t, WithDeviceName("new-device"), WithTenantID(tenantID))
		s.CreateSession(t, WithSessionDevice(oldDevice))

		err := st.SessionUpdateDeviceUID(ctx, oldDevice, newDevice)
		require.NoError(t, err)
	})
}

// TestSessionUpdate tests session updates
func (s *Suite) TestSessionUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		err := st.SessionUpdate(ctx, &models.Session{
			UID:           string(sessionUID),
			Authenticated: true,
		})
		require.NoError(t, err)

		session, err := st.SessionResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.SessionUIDResolver, string(sessionUID))
		require.NoError(t, err)
		assert.True(t, session.Authenticated)
	})

	t.Run("fails when session is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.SessionUpdate(ctx, &models.Session{
			UID:           "nonexistent-session-uid",
			Authenticated: true,
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when setting Authenticated to true", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		sessionUID := s.CreateSession(t, WithSessionUser("user2"))

		err := st.SessionUpdate(ctx, &models.Session{
			UID:           string(sessionUID),
			Authenticated: true,
			StartedAt:     time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			TenantID:      tenantID,
		})
		require.NoError(t, err)
	})

	t.Run("succeeds when updating Type field", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("user3"))

		err := st.SessionUpdate(ctx, &models.Session{
			UID:  string(sessionUID),
			Type: "exec",
		})
		require.NoError(t, err)
	})

	t.Run("succeeds when updating Recorded flag", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("user4"))

		err := st.SessionUpdate(ctx, &models.Session{
			UID:      string(sessionUID),
			Recorded: true,
		})
		require.NoError(t, err)
	})
}

// TestActiveSessionDelete tests active session deletion
func (s *Suite) TestActiveSessionDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when session is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.ActiveSessionDelete(ctx, models.UID("nonexistent"))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionActive(true))

		err := st.ActiveSessionDelete(ctx, sessionUID)
		require.NoError(t, err)
	})
}

// TestActiveSessionResolve tests active session resolution
func (s *Suite) TestActiveSessionResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when active session is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		activeSession, err := st.ActiveSessionResolve(ctx, store.SessionUIDResolver, "nonexistent")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, activeSession)
	})

	t.Run("succeeds when active session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		sessionUID := s.CreateSession(t,
			WithSessionDevice(deviceUID),
			WithSessionActive(true),
		)

		activeSession, err := st.ActiveSessionResolve(ctx, store.SessionUIDResolver, string(sessionUID))
		require.NoError(t, err)
		require.NotNil(t, activeSession)
		assert.Equal(t, sessionUID, activeSession.UID)
		assert.Equal(t, tenantID, activeSession.TenantID)
	})
}

// TestActiveSessionUpdate tests active session updates
func (s *Suite) TestActiveSessionUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when active session is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.ActiveSessionUpdate(ctx, &models.ActiveSession{
			UID:      "nonexistent",
			LastSeen: clock.Now(),
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when active session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionActive(true))

		newTime := time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC)
		err := st.ActiveSessionUpdate(ctx, &models.ActiveSession{
			UID:      sessionUID,
			LastSeen: newTime,
		})
		require.NoError(t, err)
	})
}

// TestSessionEventsCreate tests session event creation
func (s *Suite) TestSessionEventsCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when creating a session event", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		event := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: clock.Now(),
			Data:      map[string]any{"output": "test output"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event)
		require.NoError(t, err)
	})

	t.Run("succeeds when creating multiple events for same session", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		for i := range 3 {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: clock.Now(),
				Data:      map[string]any{"output": "test output"},
				Seat:      i,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}
	})
}

// TestSessionEventsList tests session events listing
func (s *Suite) TestSessionEventsList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no events found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		events, count, err := st.SessionEventsList(ctx, "nonexistent", 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Empty(t, events)
		assert.Equal(t, 0, count)
	})

	t.Run("succeeds when events are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		for range 3 {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: clock.Now(),
				Data:      map[string]any{"output": "test output"},
				Seat:      1,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, events, 3)
	})

	t.Run("succeeds filtering by seat", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		for seat := 1; seat <= 2; seat++ {
			for range 2 {
				event := &models.SessionEvent{
					Session:   string(sessionUID),
					Type:      models.SessionEventTypePtyOutput,
					Timestamp: clock.Now(),
					Data:      map[string]any{"output": "test output"},
					Seat:      seat,
				}

				err := st.SessionEventsCreate(ctx, event)
				require.NoError(t, err)
			}
		}

		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, events, 2)
	})

	t.Run("succeeds filtering by event type", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		event1 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: clock.Now(),
			Data:      map[string]any{"output": "test output"},
			Seat:      1,
		}

		event2 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyRequest,
			Timestamp: clock.Now(),
			Data:      map[string]any{"request": "test request"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event1)
		require.NoError(t, err)

		err = st.SessionEventsCreate(ctx, event2)
		require.NoError(t, err)

		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, events, 1)
	})
}

// TestSessionEventsDelete tests session events deletion
func (s *Suite) TestSessionEventsDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no events exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		err := st.SessionEventsDelete(ctx, "nonexistent", 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
	})

	t.Run("succeeds when deleting events", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		for range 3 {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: clock.Now(),
				Data:      map[string]any{"output": "test output"},
				Seat:      1,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		err := st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, events)
	})

	t.Run("succeeds deleting only matching seat", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		for seat := 1; seat <= 2; seat++ {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: clock.Now(),
				Data:      map[string]any{"output": "test output"},
				Seat:      seat,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		err := st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		events1, count1, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count1)
		assert.Empty(t, events1)

		events2, count2, err := st.SessionEventsList(ctx, sessionUID, 2, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 1, count2)
		assert.Len(t, events2, 1)
	})

	t.Run("succeeds deleting only matching event type", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		event1 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: clock.Now(),
			Data:      map[string]any{"output": "test output"},
			Seat:      1,
		}

		event2 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyRequest,
			Timestamp: clock.Now(),
			Data:      map[string]any{"request": "test request"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event1)
		require.NoError(t, err)

		err = st.SessionEventsCreate(ctx, event2)
		require.NoError(t, err)

		err = st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		events1, count1, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count1)
		assert.Empty(t, events1)

		events2, count2, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyRequest)
		require.NoError(t, err)
		assert.Equal(t, 1, count2)
		assert.Len(t, events2, 1)
	})
}

// TestSessionCleanup tests the SessionListExpired/SessionDeleteMany pair across all
// implementations. They are exercised together because retention only ever uses them as a pair:
// list a batch, act on it, delete it.
//
// The cases below age sessions through pinClock rather than by setting StartedAt, because
// SessionCreate stamps started_at from the clock and ignores whatever the caller passed. Pinning
// the clock is the only way to build the age distribution retention is about.
func (s *Suite) TestSessionCleanup(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	cutoff := now.AddDate(0, 0, -180)

	listUIDs := func(t *testing.T) []string {
		t.Helper()

		sessions, _, err := st.SessionList(ctx, scope.NewUnbounded(reasonTestQueryMechanics),
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)

		uids := make([]string, 0, len(sessions))
		for _, session := range sessions {
			uids = append(uids, session.UID)
		}

		return uids
	}

	prune := func(t *testing.T, limit int) int64 {
		t.Helper()

		expired, err := st.SessionListExpired(ctx, cutoff, limit)
		require.NoError(t, err)

		uids := make([]string, len(expired))
		for i, session := range expired {
			uids[i] = session.UID
		}

		deleted, err := st.SessionDeleteMany(ctx, uids)
		require.NoError(t, err)

		return deleted
	}

	t.Run("finds nothing when every session is newer than the cutoff", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, 1))
		uid := s.CreateSession(t, WithSessionActive(false))
		clk.now = now

		assert.Equal(t, int64(0), prune(t, 100))
		assert.Equal(t, []string{string(uid)}, listUIDs(t))
	})

	t.Run("deletes only the sessions started before the cutoff", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -10))
		device := s.CreateDevice(t)
		s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))
		s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = cutoff.AddDate(0, 0, 10)
		kept := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = now

		assert.Equal(t, int64(2), prune(t, 100))
		assert.Equal(t, []string{string(kept)}, listUIDs(t))
	})

	t.Run("lists the oldest first and stops at the limit", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		device := s.CreateDevice(t)
		oldest := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = cutoff.AddDate(0, 0, -20)
		middle := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = cutoff.AddDate(0, 0, -10)
		newest := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = now

		expired, err := st.SessionListExpired(ctx, cutoff, 2)
		require.NoError(t, err)
		assert.Equal(t, []store.ExpiredSession{
			{UID: string(oldest), Recorded: false},
			{UID: string(middle), Recorded: false},
		}, expired)

		deleted, err := st.SessionDeleteMany(ctx, []string{string(oldest), string(middle)})
		require.NoError(t, err)
		assert.Equal(t, int64(2), deleted)
		assert.Equal(t, []string{string(newest)}, listUIDs(t))
	})

	t.Run("reports whether each expired session was recorded", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		device := s.CreateDevice(t)
		plain := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = cutoff.AddDate(0, 0, -20)
		recorded := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))
		require.NoError(t, st.SessionUpdate(ctx, &models.Session{UID: string(recorded), Recorded: true}))

		clk.now = now

		expired, err := st.SessionListExpired(ctx, cutoff, 100)
		require.NoError(t, err)
		assert.Equal(t, []store.ExpiredSession{
			{UID: string(plain), Recorded: false},
			{UID: string(recorded), Recorded: true},
		}, expired)
	})

	t.Run("leaves an active session in place however old it is", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		device := s.CreateDevice(t)
		active := s.CreateSession(t, WithSessionDevice(device), WithSessionActive(true))
		s.CreateSession(t, WithSessionDevice(device), WithSessionActive(false))

		clk.now = now

		assert.Equal(t, int64(1), prune(t, 100))
		assert.Equal(t, []string{string(active)}, listUIDs(t))
	})

	t.Run("takes the session's events with it", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		uid := s.CreateSession(t, WithSessionActive(false))

		require.NoError(t, st.SessionEventsCreate(ctx, &models.SessionEvent{
			Session:   string(uid),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: clk.now,
			Data:      map[string]any{"output": "test output"},
			Seat:      1,
		}))

		clk.now = now

		assert.Equal(t, int64(1), prune(t, 100))

		_, count, err := st.SessionEventsList(ctx, uid, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("lists nothing when the limit is not positive", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		uid := s.CreateSession(t, WithSessionActive(false))
		clk.now = now

		expired, err := st.SessionListExpired(ctx, cutoff, 0)
		require.NoError(t, err)
		assert.Empty(t, expired)
		assert.Equal(t, []string{string(uid)}, listUIDs(t))
	})

	t.Run("deleting none is a no-op", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		clk := pinClock(t, cutoff.AddDate(0, 0, -30))
		uid := s.CreateSession(t, WithSessionActive(false))
		clk.now = now

		deleted, err := st.SessionDeleteMany(ctx, []string{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
		assert.Equal(t, []string{string(uid)}, listUIDs(t))
	})
}
