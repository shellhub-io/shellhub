package entity

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// Session is a row of sessions — the record of one SSH connection, kept after the connection
// itself has ended.
type Session struct {
	bun.BaseModel `bun:"table:sessions"`

	ID            string    `bun:"id,pk"`
	NamespaceID   string    `bun:"namespace_id"`
	DeviceID      string    `bun:"device_id"`
	Username      string    `bun:"username"`
	UserID        string    `bun:"user_id,nullzero"`
	IPAddress     string    `bun:"ip_address"`
	StartedAt     time.Time `bun:"started_at"`
	SeenAt        time.Time `bun:"seen_at"`
	Closed        bool      `bun:"closed"`
	Authenticated bool      `bun:"authenticated"`
	Recorded      bool      `bun:"recorded"`
	Type          string    `bun:"type"`
	Term          string    `bun:"term"`
	Web           bool      `bun:"web"`
	Longitude     float64   `bun:"longitude"`
	Latitude      float64   `bun:"latitude"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
	// Active indicates if the session is currently active (computed from active_sessions table)
	Active bool `bun:"active,scanonly"`
	// EventTypes is a comma-separated list of unique event types
	EventTypes string `bun:"event_types,scanonly"`
	// EventSeats is a comma-separated list of unique seats as integers
	EventSeats string `bun:"event_seats,scanonly"`

	Device    *Device    `bun:"rel:belongs-to,join:device_id=id"`
	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
}

// SessionFromModel projects a session into its row form.
func SessionFromModel(model *models.Session) *Session {
	sessionType := model.Type
	if sessionType == "" {
		sessionType = "shell"
	}

	session := &Session{
		ID:            model.UID,
		NamespaceID:   model.TenantID,
		DeviceID:      string(model.DeviceUID),
		Username:      model.Username,
		UserID:        model.UserID,
		IPAddress:     model.IPAddress,
		StartedAt:     model.StartedAt,
		SeenAt:        model.LastSeen,
		Closed:        model.Closed,
		Authenticated: model.Authenticated,
		Recorded:      model.Recorded,
		Type:          sessionType,
		Term:          model.Term,
		Web:           model.Web,
		Longitude:     model.Position.Longitude,
		Latitude:      model.Position.Latitude,
		UpdatedAt:     clock.Now(),
	}

	return session
}

// SessionToModel rebuilds a session from its row.
func SessionToModel(entity *Session) *models.Session {
	session := &models.Session{
		UID:           strings.TrimSpace(entity.ID),
		TenantID:      entity.NamespaceID,
		DeviceUID:     models.UID(strings.TrimSpace(entity.DeviceID)),
		Username:      entity.Username,
		UserID:        entity.UserID,
		IPAddress:     entity.IPAddress,
		StartedAt:     entity.StartedAt,
		LastSeen:      entity.SeenAt,
		Active:        entity.Active,
		Closed:        entity.Closed,
		Authenticated: entity.Authenticated,
		Recorded:      entity.Recorded,
		Type:          entity.Type,
		Term:          entity.Term,
		Web:           entity.Web,
		Position: models.SessionPosition{
			Longitude: entity.Longitude,
			Latitude:  entity.Latitude,
		},
		Events: models.SessionEvents{
			Types: parseEventTypes(entity.EventTypes),
			Seats: parseEventSeats(entity.EventSeats),
		},
	}

	if entity.Device != nil {
		session.Device = DeviceToModel(entity.Device)
	}

	return session
}

// ActiveSession is a row of active_sessions, holding only the liveness of a session. It is a
// separate table because it is written on every keep-alive, while the session row is not.
type ActiveSession struct {
	bun.BaseModel `bun:"table:active_sessions"`

	SessionID string    `bun:"session_id,pk"`
	SeenAt    time.Time `bun:"seen_at"`
	// skipupdate: ActiveSessionFromModel always stamps clock.Now(), so without this every
	// ActiveSessionUpdate would reset created_at to "now".
	CreatedAt time.Time `bun:"created_at,skipupdate"`

	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
}

// ActiveSessionFromModel projects a liveness record into its row form, stamping the current time.
func ActiveSessionFromModel(model *models.ActiveSession) *ActiveSession {
	return &ActiveSession{
		SessionID: string(model.UID),
		SeenAt:    model.LastSeen,
		CreatedAt: clock.Now(),
	}
}

// ActiveSessionToModel rebuilds a liveness record from its row, taking the namespace from the
// session relation when it was loaded.
func ActiveSessionToModel(entity *ActiveSession) *models.ActiveSession {
	activeSession := &models.ActiveSession{
		UID:      models.UID(strings.TrimSpace(entity.SessionID)),
		LastSeen: entity.SeenAt,
	}

	if entity.Session != nil {
		activeSession.TenantID = entity.Session.NamespaceID
	}

	return activeSession
}

// SessionEvent is a row of session_events — one thing that happened during a session, which for
// a recorded session includes its output.
type SessionEvent struct {
	bun.BaseModel `bun:"table:session_events"`

	ID        string    `bun:"id,pk"`
	SessionID string    `bun:"session_id"`
	Type      string    `bun:"type"`
	Seat      int       `bun:"seat"`
	Data      string    `bun:"data"`
	CreatedAt time.Time `bun:"created_at"`

	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
}

// SessionEventFromModel projects an event into its row form.
func SessionEventFromModel(model *models.SessionEvent) *SessionEvent {
	event := &SessionEvent{
		SessionID: model.Session,
		Type:      string(model.Type),
		Seat:      model.Seat,
		CreatedAt: model.Timestamp,
	}

	if model.Data != nil {
		if dataBytes, err := json.Marshal(model.Data); err == nil {
			event.Data = string(dataBytes)
		}
	}

	return event
}

// SessionEventToModel rebuilds an event from its row.
func SessionEventToModel(entity *SessionEvent) *models.SessionEvent {
	event := &models.SessionEvent{
		Session:   entity.SessionID,
		Type:      models.SessionEventType(entity.Type),
		Timestamp: entity.CreatedAt,
		Seat:      entity.Seat,
	}

	if entity.Data != "" {
		var data any
		if err := json.Unmarshal([]byte(entity.Data), &data); err == nil {
			event.Data = data
		}
	}

	return event
}

func parseEventTypes(eventTypes string) []string {
	if eventTypes == "" {
		return []string{}
	}

	types := strings.Split(eventTypes, ",")
	result := make([]string, 0, len(types))

	for _, t := range types {
		if trimmed := strings.TrimSpace(t); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func parseEventSeats(eventSeats string) []int {
	if eventSeats == "" {
		return []int{}
	}

	seats := strings.Split(eventSeats, ",")
	result := make([]int, 0, len(seats))

	for _, s := range seats {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			if seat, err := strconv.Atoi(trimmed); err == nil {
				result = append(result, seat)
			}
		}
	}

	return result
}
