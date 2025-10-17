package entity

import (
	"encoding/json"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions"`

	ID            string    `bun:"id,pk"`
	DeviceID      string    `bun:"device_id"`
	Username      string    `bun:"username"`
	IPAddress     string    `bun:"ip_address"`
	StartedAt     time.Time `bun:"started_at"`
	SeenAt        time.Time `bun:"seen_at"`
	Closed        bool      `bun:"closed"`
	Authenticated bool      `bun:"authenticated"`
	Recorded      bool      `bun:"recorded"`
	Type          string    `bun:"type"`
	Term          string    `bun:"term"`
	Longitude     float64   `bun:"longitude"`
	Latitude      float64   `bun:"latitude"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`

	Device *Device `bun:"rel:belongs-to,join:device_id=id"`
}

func SessionFromModel(model *models.Session) *Session {
	session := &Session{
		ID:            model.UID,
		DeviceID:      string(model.DeviceUID),
		Username:      model.Username,
		IPAddress:     model.IPAddress,
		StartedAt:     model.StartedAt,
		SeenAt:        model.LastSeen,
		Closed:        model.Closed,
		Authenticated: model.Authenticated,
		Recorded:      model.Recorded,
		Type:          model.Type,
		Term:          model.Term,
		Longitude:     model.Position.Longitude,
		Latitude:      model.Position.Latitude,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return session
}

func SessionToModel(entity *Session) *models.Session {
	session := &models.Session{
		UID:           entity.ID,
		DeviceUID:     models.UID(entity.DeviceID),
		Username:      entity.Username,
		IPAddress:     entity.IPAddress,
		StartedAt:     entity.StartedAt,
		LastSeen:      entity.SeenAt,
		Closed:        entity.Closed,
		Authenticated: entity.Authenticated,
		Recorded:      entity.Recorded,
		Type:          entity.Type,
		Term:          entity.Term,
		Position: models.SessionPosition{
			Longitude: entity.Longitude,
			Latitude:  entity.Latitude,
		},
	}

	if entity.Device != nil {
		session.Device = DeviceToModel(entity.Device)
		session.TenantID = entity.Device.NamespaceID
	}

	return session
}

type ActiveSession struct {
	bun.BaseModel `bun:"table:active_sessions"`

	SessionID string    `bun:"session_id,pk"`
	SeenAt    time.Time `bun:"seen_at"`
	CreatedAt time.Time `bun:"created_at"`

	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
}

func ActiveSessionFromModel(model *models.ActiveSession) *ActiveSession {
	return &ActiveSession{
		SessionID: string(model.UID),
		SeenAt:    model.LastSeen,
		CreatedAt: time.Now(),
	}
}

func ActiveSessionToModel(entity *ActiveSession) *models.ActiveSession {
	activeSession := &models.ActiveSession{
		UID:      models.UID(entity.SessionID),
		LastSeen: entity.SeenAt,
	}

	if entity.Session != nil && entity.Session.Device != nil {
		activeSession.TenantID = entity.Session.Device.NamespaceID
	}

	return activeSession
}

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

func SessionEventToModel(entity *SessionEvent) *models.SessionEvent {
	event := &models.SessionEvent{
		Session:   entity.SessionID,
		Type:      models.SessionEventType(entity.Type),
		Timestamp: entity.CreatedAt,
		Seat:      entity.Seat,
	}

	if entity.Data != "" {
		var data interface{}
		if err := json.Unmarshal([]byte(entity.Data), &data); err == nil {
			event.Data = data
		}
	}

	return event
}