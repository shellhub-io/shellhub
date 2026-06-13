package models

import "time"

// ShareCreateRequest is the payload sent by the agent to register a new shareable terminal session.
type ShareCreateRequest struct {
	// Name is an optional human-friendly label for the share, shown in the namespace's list.
	Name string `json:"name"`
	// Command is the command being shared (informational; e.g. "bash" or "claude").
	Command string `json:"command"`
	// Writable, when true, lets guests type into the session (collaborative mode). Defaults to
	// read-only.
	Writable bool `json:"writable"`
	// TTLSeconds controls when the share token expires: 0 uses the server default, a negative value
	// means no expiry (the share only ends when the command exits), and a positive value sets a
	// custom lifetime in seconds.
	TTLSeconds int `json:"ttl_seconds"`
	// Term is the terminal type reported by the host (e.g. "xterm-256color").
	Term string `json:"term"`
	// Cols and Rows are the initial terminal dimensions.
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// ShareCreateResponse is returned to the agent after a shareable terminal session is created.
type ShareCreateResponse struct {
	// Token is the opaque, unguessable identifier used both to push the stream and to view it.
	Token string `json:"token"`
	// URL is the public, read-only address a guest can open to watch the session.
	URL string `json:"url"`
	// ExpiresAt is the moment the share token stops being valid.
	ExpiresAt time.Time `json:"expires_at"`
}

// ShareInfo describes an active shareable terminal session, listed for the namespace owner so they
// can see what is being shared and how many people are currently watching.
type ShareInfo struct {
	Token        string    `json:"token"`
	URL          string    `json:"url"`
	Name         string    `json:"name"`
	Command      string    `json:"command"`
	Writable     bool      `json:"writable"`
	DeviceUID    string    `json:"device_uid"`
	DeviceName   string    `json:"device_name"`
	DeviceOnline bool      `json:"device_online"`
	DeviceOS     string    `json:"device_os"`
	Viewers      int       `json:"viewers"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}
