package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionFromModel(t *testing.T) {
	now := time.Now()

	clockMock := clockmock.NewClock(t)
	oldClock := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	t.Cleanup(func() { clock.DefaultBackend = oldClock })

	tests := []struct {
		name  string
		model *models.Session
		check func(t *testing.T, result *Session)
	}{
		{
			name: "full fields",
			model: &models.Session{
				UID:           "session-uid-1",
				DeviceUID:     "device-uid-1",
				Username:      "root",
				IPAddress:     "192.168.1.1",
				StartedAt:     now,
				LastSeen:      now,
				Closed:        false,
				Authenticated: true,
				Recorded:      true,
				Type:          "shell",
				Term:          "xterm-256color",
				Position: models.SessionPosition{
					Longitude: 1.23,
					Latitude:  4.56,
				},
			},
			check: func(t *testing.T, result *Session) {
				assert.Equal(t, "session-uid-1", result.ID)
				assert.Equal(t, "device-uid-1", result.DeviceID)
				assert.Equal(t, "root", result.Username)
				assert.Equal(t, "192.168.1.1", result.IPAddress)
				assert.Equal(t, now, result.StartedAt)
				assert.Equal(t, now, result.SeenAt)
				assert.False(t, result.Closed)
				assert.True(t, result.Authenticated)
				assert.True(t, result.Recorded)
				assert.Equal(t, "shell", result.Type)
				assert.Equal(t, "xterm-256color", result.Term)
				assert.InDelta(t, 1.23, result.Longitude, 0.001)
				assert.InDelta(t, 4.56, result.Latitude, 0.001)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, now, result.UpdatedAt)
			},
		},
		{
			name: "empty Type defaults to shell",
			model: &models.Session{
				UID:  "session-uid-2",
				Type: "",
			},
			check: func(t *testing.T, result *Session) {
				assert.Equal(t, "shell", result.Type)
				assert.InDelta(t, 0.0, result.Longitude, 0.001)
				assert.InDelta(t, 0.0, result.Latitude, 0.001)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clockMock.On("Now").Return(now).Once()
			result := SessionFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestSessionToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		entity *Session
		check  func(t *testing.T, result *models.Session)
	}{
		{
			name: "full fields with Device",
			entity: &Session{
				ID:            "session-uid-1",
				DeviceID:      "device-uid-1",
				Username:      "root",
				IPAddress:     "192.168.1.1",
				StartedAt:     now,
				SeenAt:        now,
				Active:        true,
				Closed:        false,
				Authenticated: true,
				Recorded:      true,
				Type:          "shell",
				Term:          "xterm",
				Longitude:     1.23,
				Latitude:      4.56,
				EventTypes:    "pty-output,window-change",
				EventSeats:    "0,1",
				Device: &Device{
					ID:          "device-uid-1",
					NamespaceID: "ns-id-1",
					Status:      "accepted",
				},
			},
			check: func(t *testing.T, result *models.Session) {
				assert.Equal(t, "session-uid-1", result.UID)
				assert.Equal(t, models.UID("device-uid-1"), result.DeviceUID)
				assert.Equal(t, "root", result.Username)
				assert.Equal(t, "192.168.1.1", result.IPAddress)
				assert.Equal(t, now, result.StartedAt)
				assert.Equal(t, now, result.LastSeen)
				assert.True(t, result.Active)
				assert.False(t, result.Closed)
				assert.True(t, result.Authenticated)
				assert.True(t, result.Recorded)
				assert.Equal(t, "shell", result.Type)
				assert.Equal(t, "xterm", result.Term)
				assert.InDelta(t, 1.23, result.Position.Longitude, 0.001)
				assert.InDelta(t, 4.56, result.Position.Latitude, 0.001)
				assert.Equal(t, []string{"pty-output", "window-change"}, result.Events.Types)
				assert.Equal(t, []int{0, 1}, result.Events.Seats)
				assert.NotNil(t, result.Device)
				assert.Equal(t, "ns-id-1", result.TenantID)
			},
		},
		{
			name: "nil Device",
			entity: &Session{
				ID:       "session-uid-2",
				DeviceID: "device-uid-2",
				Type:     "exec",
				Device:   nil,
			},
			check: func(t *testing.T, result *models.Session) {
				assert.Nil(t, result.Device)
				assert.Equal(t, "", result.TenantID)
			},
		},
		{
			name: "whitespace trimming on ID",
			entity: &Session{
				ID:       "  session-uid-3  ",
				DeviceID: "  device-uid-3  ",
				Type:     "shell",
			},
			check: func(t *testing.T, result *models.Session) {
				assert.Equal(t, "session-uid-3", result.UID)
				assert.Equal(t, models.UID("device-uid-3"), result.DeviceUID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SessionToModel(tt.entity)
			tt.check(t, result)
		})
	}
}

func TestActiveSessionFromModel(t *testing.T) {
	now := time.Now()

	clockMock := clockmock.NewClock(t)
	oldClock := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	t.Cleanup(func() { clock.DefaultBackend = oldClock })

	tests := []struct {
		name  string
		model *models.ActiveSession
		check func(t *testing.T, result *ActiveSession)
	}{
		{
			name: "full fields",
			model: &models.ActiveSession{
				UID:      "active-session-1",
				LastSeen: now,
			},
			check: func(t *testing.T, result *ActiveSession) {
				assert.Equal(t, "active-session-1", result.SessionID)
				assert.Equal(t, now, result.SeenAt)
				assert.Equal(t, now, result.CreatedAt)
			},
		},
		{
			name: "zero-value LastSeen",
			model: &models.ActiveSession{
				UID: "active-session-2",
			},
			check: func(t *testing.T, result *ActiveSession) {
				assert.Equal(t, "active-session-2", result.SessionID)
				assert.True(t, result.SeenAt.IsZero())
				assert.Equal(t, now, result.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clockMock.On("Now").Return(now).Once()
			result := ActiveSessionFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestActiveSessionToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		entity *ActiveSession
		check  func(t *testing.T, result *models.ActiveSession)
	}{
		{
			name: "with Session and Device loaded",
			entity: &ActiveSession{
				SessionID: "active-session-1",
				SeenAt:    now,
				Session: &Session{
					Device: &Device{
						NamespaceID: "ns-id-1",
					},
				},
			},
			check: func(t *testing.T, result *models.ActiveSession) {
				assert.Equal(t, models.UID("active-session-1"), result.UID)
				assert.Equal(t, now, result.LastSeen)
				assert.Equal(t, "ns-id-1", result.TenantID)
			},
		},
		{
			name: "nil Session",
			entity: &ActiveSession{
				SessionID: "active-session-2",
				SeenAt:    now,
				Session:   nil,
			},
			check: func(t *testing.T, result *models.ActiveSession) {
				assert.Equal(t, models.UID("active-session-2"), result.UID)
				assert.Equal(t, "", result.TenantID)
			},
		},
		{
			name: "Session with nil Device",
			entity: &ActiveSession{
				SessionID: "active-session-3",
				SeenAt:    now,
				Session: &Session{
					Device: nil,
				},
			},
			check: func(t *testing.T, result *models.ActiveSession) {
				assert.Equal(t, models.UID("active-session-3"), result.UID)
				assert.Equal(t, "", result.TenantID)
			},
		},
		{
			name: "whitespace trimming on SessionID",
			entity: &ActiveSession{
				SessionID: "  active-session-4  ",
				SeenAt:    now,
			},
			check: func(t *testing.T, result *models.ActiveSession) {
				assert.Equal(t, models.UID("active-session-4"), result.UID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ActiveSessionToModel(tt.entity)
			tt.check(t, result)
		})
	}
}

func TestSessionEventFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		model *models.SessionEvent
		check func(t *testing.T, result *SessionEvent)
	}{
		{
			name: "full fields with Data",
			model: &models.SessionEvent{
				Session:   "session-1",
				Type:      models.SessionEventTypePtyOutput,
				Timestamp: now,
				Seat:      0,
				Data:      map[string]interface{}{"output": "hello"},
			},
			check: func(t *testing.T, result *SessionEvent) {
				assert.Equal(t, "session-1", result.SessionID)
				assert.Equal(t, "pty-output", result.Type)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, 0, result.Seat)
				assert.Equal(t, `{"output":"hello"}`, result.Data)
			},
		},
		{
			name: "nil Data",
			model: &models.SessionEvent{
				Session:   "session-2",
				Type:      models.SessionEventTypeShell,
				Timestamp: now,
				Seat:      1,
				Data:      nil,
			},
			check: func(t *testing.T, result *SessionEvent) {
				assert.Equal(t, "session-2", result.SessionID)
				assert.Equal(t, "shell", result.Type)
				assert.Equal(t, "", result.Data)
				assert.Equal(t, 1, result.Seat)
				assert.Equal(t, now, result.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SessionEventFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestSessionEventToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		entity *SessionEvent
		check  func(t *testing.T, result *models.SessionEvent)
	}{
		{
			name: "full fields with Data string",
			entity: &SessionEvent{
				SessionID: "session-1",
				Type:      "pty-output",
				CreatedAt: now,
				Seat:      0,
				Data:      `{"output":"hello"}`,
			},
			check: func(t *testing.T, result *models.SessionEvent) {
				assert.Equal(t, "session-1", result.Session)
				assert.Equal(t, models.SessionEventTypePtyOutput, result.Type)
				assert.Equal(t, now, result.Timestamp)
				assert.Equal(t, 0, result.Seat)
				require.NotNil(t, result.Data)
				assert.Equal(t, map[string]interface{}{"output": "hello"}, result.Data)
			},
		},
		{
			name: "empty Data",
			entity: &SessionEvent{
				SessionID: "session-2",
				Type:      "shell",
				CreatedAt: now,
				Data:      "",
			},
			check: func(t *testing.T, result *models.SessionEvent) {
				assert.Nil(t, result.Data)
			},
		},
		{
			name: "invalid JSON",
			entity: &SessionEvent{
				SessionID: "session-3",
				Type:      "exec",
				CreatedAt: now,
				Data:      "not-json{",
			},
			check: func(t *testing.T, result *models.SessionEvent) {
				assert.Nil(t, result.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SessionEventToModel(tt.entity)
			tt.check(t, result)
		})
	}
}

func TestParseEventTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single",
			input:    "pty-output",
			expected: []string{"pty-output"},
		},
		{
			name:     "multiple",
			input:    "pty-output,window-change,shell",
			expected: []string{"pty-output", "window-change", "shell"},
		},
		{
			name:     "trailing commas",
			input:    "pty-output,window-change,",
			expected: []string{"pty-output", "window-change"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEventTypes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEventSeats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "empty",
			input:    "",
			expected: []int{},
		},
		{
			name:     "single",
			input:    "0",
			expected: []int{0},
		},
		{
			name:     "multiple",
			input:    "0,1,2",
			expected: []int{0, 1, 2},
		},
		{
			name:     "invalid number skipped",
			input:    "0,abc,2",
			expected: []int{0, 2},
		},
		{
			name:     "trailing commas",
			input:    "0,1,",
			expected: []int{0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEventSeats(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
