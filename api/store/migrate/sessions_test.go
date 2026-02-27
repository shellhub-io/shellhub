package migrate

import (
	"testing"
	"time"

	"github.com/google/uuid" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSession(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	lastSeen := now.Add(time.Hour)

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoSession{
			UID:           "session-uid-123",
			DeviceUID:     "device-uid-123",
			Username:      "root",
			IPAddress:     "192.168.1.1",
			StartedAt:     now,
			LastSeen:      lastSeen,
			Closed:        true,
			Authenticated: true,
			Recorded:      true,
			Type:          "exec",
			Term:          "xterm",
			Position:      &mongoSessionPos{Longitude: -46.6, Latitude: -23.5},
		}

		result := convertSession(doc)

		assert.Equal(t, "session-uid-123", result.ID)
		assert.Equal(t, "device-uid-123", result.DeviceID)
		assert.Equal(t, "root", result.Username)
		assert.Equal(t, "192.168.1.1", result.IPAddress)
		assert.Equal(t, now, result.StartedAt)
		assert.Equal(t, lastSeen, result.SeenAt)
		assert.True(t, result.Closed)
		assert.True(t, result.Authenticated)
		assert.True(t, result.Recorded)
		assert.Equal(t, "exec", result.Type)
		assert.Equal(t, "xterm", result.Term)
		assert.Equal(t, now, result.CreatedAt)
		assert.Equal(t, lastSeen, result.UpdatedAt)
		assert.InDelta(t, -46.6, result.Longitude, 0.001)
		assert.InDelta(t, -23.5, result.Latitude, 0.001)
	})

	t.Run("defaults for empty type", func(t *testing.T) {
		doc := mongoSession{
			UID: "session-uid-456",
		}

		result := convertSession(doc)

		assert.Equal(t, "shell", result.Type)
	})

	t.Run("nil position", func(t *testing.T) {
		doc := mongoSession{
			UID: "session-uid-789",
		}

		result := convertSession(doc)

		assert.Zero(t, result.Longitude)
		assert.Zero(t, result.Latitude)
	})
}

func TestConvertSessionEvent(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("with data", func(t *testing.T) {
		doc := mongoSessionEvent{
			Session:   "session-123",
			Type:      "pty-output",
			Timestamp: now,
			Data:      map[string]any{"output": "hello"},
			Seat:      1,
		}

		result := convertSessionEvent(doc)

		_, err := uuid.Parse(result.ID)
		require.NoError(t, err)
		assert.Equal(t, "session-123", result.SessionID)
		assert.Equal(t, "pty-output", result.Type)
		assert.Equal(t, 1, result.Seat)
		assert.Equal(t, now, result.CreatedAt)
		assert.Contains(t, result.Data, `"output":"hello"`)
	})

	t.Run("nil data", func(t *testing.T) {
		doc := mongoSessionEvent{
			Session:   "session-456",
			Type:      "pty-output",
			Timestamp: now,
		}

		result := convertSessionEvent(doc)

		assert.Empty(t, result.Data)
	})

	t.Run("unique IDs", func(t *testing.T) {
		doc := mongoSessionEvent{Session: "s1", Type: "t1", Timestamp: now}
		a := convertSessionEvent(doc)
		b := convertSessionEvent(doc)
		assert.NotEqual(t, a.ID, b.ID)
	})
}
