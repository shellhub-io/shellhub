package models

import (
	"time"
)

// SessionPosition is where the client connected from, resolved from its address by GeoIP. It is
// recorded once at session start and not refreshed.
type SessionPosition struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Session is one SSH connection to one device, live or finished. It is created when the
// connection is established and outlives it: Active says whether it is still running.
type Session struct {
	UID       string  `json:"uid"`
	DeviceUID UID     `json:"device_uid,omitempty"`
	Device    *Device `json:"device"`
	TenantID  string  `json:"tenant_id"`
	Username  string  `json:"username"`
	// UserID is the ShellHub account that authorized this session via browser
	// approval. Empty for password/public-key logins and web-terminal sessions.
	// It is not serialized; readers get Principal.
	UserID string `json:"-"`
	// APIKeyID is the API key this session acts as, when an automation opened it. Like UserID, it
	// is written here and read through Principal.
	APIKeyID string `json:"-"`
	// Principal is who opened the session: the person or the API key that presented the
	// credential, projected from UserID and APIKeyID. It is absent, rather than empty, under the
	// legacy access model, where a session has no principal by design.
	Principal     *Principal      `json:"principal,omitempty"`
	IPAddress     string          `json:"ip_address"`
	StartedAt     time.Time       `json:"started_at"`
	LastSeen      time.Time       `json:"last_seen"`
	Active        bool            `json:"active"`
	Authenticated bool            `json:"authenticated"`
	Recorded      bool            `json:"recorded"`
	Term          string          `json:"term"`
	Web           bool            `json:"web"`
	Position      SessionPosition `json:"position"`
	Events        SessionEvents   `json:"events"`
}

// ActiveSession is the liveness half of a session, written on the keep-alive path. It is kept
// apart from Session so a heartbeat does not rewrite the whole record.
type ActiveSession struct {
	UID      UID       `json:"uid"`
	LastSeen time.Time `json:"last_seen"`
	TenantID string    `json:"tenant_id"`
}

// RecordedSession is one frame of a recorded terminal session. Recording is a cloud feature and
// the type lives there too; this copy exists because migrations reference it, so it cannot move
// until they no longer do.
type RecordedSession struct {
	UID      UID       `json:"uid"`
	Message  string    `json:"message"`
	TenantID string    `json:"tenant_id"`
	Time     time.Time `json:"time"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
}

// Status is the authentication state of a session, as the agent reports it back once the SSH
// handshake has completed.
type Status struct {
	Authenticated bool `json:"authenticated"`
}

// SessionUpdate is a partial update to a session: a nil field is left alone, which is why every
// field is a pointer.
type SessionUpdate struct {
	Recorded      *bool `json:"recorded"`
	Authenticated *bool `json:"authenticated"`
}

// SessionEventType names the SSH request an event came from. The values are the wire names from
// the SSH protocol, not names of our own, so they can be matched against a packet capture.
type SessionEventType string

// The event types a session can record. All but pty-output are SSH request names as they appear on
// the wire; pty-output is ours, carrying the bytes the terminal produced, which SSH itself sends as
// channel data rather than as a request.
const (
	SessionEventTypePtyOutput SessionEventType = "pty-output"

	SessionEventTypePtyRequest   SessionEventType = "pty-req"
	SessionEventTypeWindowChange SessionEventType = "window-change"
	SessionEventTypeExitCode     SessionEventType = "exit-code"

	SessionEventTypeExitStatus SessionEventType = "exit-status"
	SessionEventTypeExitSignal SessionEventType = "exit-signal"

	SessionEventTypeEnv       SessionEventType = "env"
	SessionEventTypeShell     SessionEventType = "shell"
	SessionEventTypeExec      SessionEventType = "exec"
	SessionEventTypeSubsystem SessionEventType = "subsystem"

	SessionEventTypeSignal       SessionEventType = "signal"
	SessionEventTypeTcpipForward SessionEventType = "tcpip-forward"
	SessionEventTypeAuthAgentReq SessionEventType = "auth-agent-req"
)

// SessionEvent represents a session event.
type SessionEvent struct {
	// Session is the session UID where the event occurred. It is not serialized.
	Session string `json:"-"`
	// Type of the session. Normally, it is the SSH request name.
	Type SessionEventType `json:"type"`
	// Timestamp contains the time when the event was logged.
	Timestamp time.Time `json:"timestamp"`
	// Data is the event's payload. Its shape follows the request: an object for one that carries
	// fields, and an empty string for one the protocol defines as having no body, such as shell.
	// It is omitted when the event carried no payload at all.
	Data any `json:"data,omitempty"`
	// Seat is the seat where the event occurred.
	Seat int `json:"seat"`
}

// SessionEvents stores the events registered in a session.
type SessionEvents struct {
	// Types field is a set of sessions type to simplify the indexing on the database.
	Types []string `json:"types"`
	// Seats contains a list of seats of events.
	Seats []int `json:"seats"`
	// First is the type of the event the session opened with, taken from the request types that
	// open a channel: pty-req, shell, exec and subsystem. It is empty when the session recorded
	// none of them.
	First SessionEventType `json:"first,omitempty"`
	// Items is the session's timeline, oldest first, carried by the by-uid request alone and
	// never by the list. It excludes terminal output and is capped, so it is not necessarily
	// every event the session recorded, and it is absent both when there is nothing to show and
	// when the read failed.
	Items []SessionEvent `json:"items,omitempty"`
}

// SessionSeat stores a session's seat.
type SessionSeat struct {
	// ID is the identifier of session's seat.
	ID int `json:"id"`
}
