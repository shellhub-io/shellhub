package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionList tests the SessionList method across all implementations
func (s *Suite) TestSessionList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when no sessions are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		sessions, count, err := st.SessionList(ctx,
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, 0, count)
	})

	t.Run("succeeds when sessions are found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test sessions
		s.CreateSession(t, WithSessionUser("user1"))
		s.CreateSession(t, WithSessionUser("user2"))
		s.CreateSession(t, WithSessionUser("user3"))
		s.CreateSession(t, WithSessionUser("user4"))

		// List all sessions
		sessions, count, err := st.SessionList(ctx,
			st.Options().Match(&query.Filters{}),
			st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
		)

		require.NoError(t, err)
		assert.Equal(t, 4, count)
		assert.Len(t, sessions, 4)
	})
}

// TestSessionResolve tests session resolution by UID
func (s *Suite) TestSessionResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when session not found by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		session, err := st.SessionResolve(ctx, store.SessionUIDResolver, "nonexistent")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, session)
	})

	t.Run("succeeds resolving session by UID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create test session
		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		sessionUID := s.CreateSession(t,
			WithSessionDevice(deviceUID),
			WithSessionUser("testuser"),
		)

		// Resolve by UID
		session, err := st.SessionResolve(ctx, store.SessionUIDResolver, string(sessionUID))
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

		// Create device first
		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		// Create session
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

		// Verify it was created
		created, err := st.SessionResolve(ctx, store.SessionUIDResolver, uid)
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

		// Create devices and session
		tenantID := s.CreateNamespace(t)
		oldDevice := s.CreateDevice(t, WithDeviceName("old-device"), WithTenantID(tenantID))
		newDevice := s.CreateDevice(t, WithDeviceName("new-device"), WithTenantID(tenantID))
		s.CreateSession(t, WithSessionDevice(oldDevice))

		// Update device UID
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

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Update session
		err := st.SessionUpdate(ctx, &models.Session{
			UID:           string(sessionUID),
			Authenticated: true,
		})
		require.NoError(t, err)

		// Verify update
		session, err := st.SessionResolve(ctx, store.SessionUIDResolver, string(sessionUID))
		require.NoError(t, err)
		assert.True(t, session.Authenticated)
	})

	t.Run("succeeds when setting Authenticated to true", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		tenantID := s.CreateNamespace(t)
		sessionUID := s.CreateSession(t, WithSessionUser("user2"))

		// Update session
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

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("user3"))

		// Update type
		err := st.SessionUpdate(ctx, &models.Session{
			UID:  string(sessionUID),
			Type: "exec",
		})
		require.NoError(t, err)
	})

	t.Run("succeeds when updating Recorded flag", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("user4"))

		// Update recorded flag
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

		// Create session with active flag
		sessionUID := s.CreateSession(t, WithSessionActive(true))

		// Delete active session
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
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, activeSession)
	})

	t.Run("succeeds when active session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create active session
		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		sessionUID := s.CreateSession(t,
			WithSessionDevice(deviceUID),
			WithSessionActive(true),
		)

		// Resolve active session
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
			LastSeen: time.Now(),
		})
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when active session is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create active session
		sessionUID := s.CreateSession(t, WithSessionActive(true))

		// Update last seen
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

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create session event
		event := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"output": "test output"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event)
		require.NoError(t, err)
	})

	t.Run("succeeds when creating multiple events for same session", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create multiple events
		for i := 0; i < 3; i++ {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"output": "test output"},
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

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events
		for i := 0; i < 3; i++ {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"output": "test output"},
				Seat:      1,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		// List events
		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, events, 3)
	})

	t.Run("succeeds filtering by seat", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events with different seats
		for seat := 1; seat <= 2; seat++ {
			for i := 0; i < 2; i++ {
				event := &models.SessionEvent{
					Session:   string(sessionUID),
					Type:      models.SessionEventTypePtyOutput,
					Timestamp: time.Now(),
					Data:      map[string]interface{}{"output": "test output"},
					Seat:      seat,
				}

				err := st.SessionEventsCreate(ctx, event)
				require.NoError(t, err)
			}
		}

		// List events for seat 1
		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, events, 2)
	})

	t.Run("succeeds filtering by event type", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events with different types
		event1 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"output": "test output"},
			Seat:      1,
		}

		event2 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyRequest,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"request": "test request"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event1)
		require.NoError(t, err)

		err = st.SessionEventsCreate(ctx, event2)
		require.NoError(t, err)

		// List only PtyOutput events
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

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events
		for i := 0; i < 3; i++ {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"output": "test output"},
				Seat:      1,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		// Delete events
		err := st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		// Verify deletion
		events, count, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, events)
	})

	t.Run("succeeds deleting only matching seat", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events with different seats
		for seat := 1; seat <= 2; seat++ {
			event := &models.SessionEvent{
				Session:   string(sessionUID),
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"output": "test output"},
				Seat:      seat,
			}

			err := st.SessionEventsCreate(ctx, event)
			require.NoError(t, err)
		}

		// Delete events for seat 1 only
		err := st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		// Verify seat 1 events are deleted
		events1, count1, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count1)
		assert.Empty(t, events1)

		// Verify seat 2 events still exist
		events2, count2, err := st.SessionEventsList(ctx, sessionUID, 2, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 1, count2)
		assert.Len(t, events2, 1)
	})

	t.Run("succeeds deleting only matching event type", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create session
		sessionUID := s.CreateSession(t, WithSessionUser("testuser"))

		// Create events with different types
		event1 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyOutput,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"output": "test output"},
			Seat:      1,
		}

		event2 := &models.SessionEvent{
			Session:   string(sessionUID),
			Type:      models.SessionEventTypePtyRequest,
			Timestamp: time.Now(),
			Data:      map[string]interface{}{"request": "test request"},
			Seat:      1,
		}

		err := st.SessionEventsCreate(ctx, event1)
		require.NoError(t, err)

		err = st.SessionEventsCreate(ctx, event2)
		require.NoError(t, err)

		// Delete only PtyOutput events
		err = st.SessionEventsDelete(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)

		// Verify PtyOutput events are deleted
		events1, count1, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyOutput)
		require.NoError(t, err)
		assert.Equal(t, 0, count1)
		assert.Empty(t, events1)

		// Verify PtyRequest events still exist
		events2, count2, err := st.SessionEventsList(ctx, sessionUID, 1, models.SessionEventTypePtyRequest)
		require.NoError(t, err)
		assert.Equal(t, 1, count2)
		assert.Len(t, events2, 1)
	})
}
