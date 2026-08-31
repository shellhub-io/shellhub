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
// connection is established and outlives it: Active says whether it is still running, and Closed
// distinguishes a session that ended cleanly from one whose device vanished.
type Session struct {
	UID       string  `json:"uid"`
	DeviceUID UID     `json:"device_uid,omitempty"`
	Device    *Device `json:"device"`
	TenantID  string  `json:"tenant_id"`
	Username  string  `json:"username"`
	// UserID is the ShellHub account that authorized this session via browser
	// approval. Empty for password/public-key logins and web-terminal sessions.
	UserID        string          `json:"user_id,omitempty"`
	IPAddress     string          `json:"ip_address"`
	StartedAt     time.Time       `json:"started_at"`
	LastSeen      time.Time       `json:"last_seen"`
	Active        bool            `json:"active"`
	Closed        bool            `json:"-"`
	Authenticated bool            `json:"authenticated"`
	Recorded      bool            `json:"recorded"`
	Type          string          `json:"type"`
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
	Recorded      *bool   `json:"recorded"`
	Authenticated *bool   `json:"authenticated"`
	Type          *string `json:"type"`
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
	// Session is the session UID where the event occurred.
	Session string `json:"session"`
	// Type of the session. Normally, it is the SSH request name.
	Type SessionEventType `json:"type"`
	// Timestamp contains the time when the event was logged.
	Timestamp time.Time `json:"timestamp"`
	// Data is a generic structure containing data of the event, normally the unmarshaling data of the request.
	Data any `json:"data"`
	// Seat is the seat where the event occurred.
	Seat int `json:"seat"`
}

// SessionEvents stores the events registered in a session.
type SessionEvents struct {
	// Types field is a set of sessions type to simplify the indexing on the database.
	Types []string `json:"types"`
	// Seats contains a list of seats of events.
	Seats []int `json:"seats"`
}

// SessionSeat stores a session's seat.
type SessionSeat struct {
	// ID is the identifier of session's seat.
	ID int `json:"id"`
}
