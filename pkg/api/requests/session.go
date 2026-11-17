package requests

import "github.com/shellhub-io/shellhub/pkg/api/query"

// SessionIDParam is a structure to represent and validate a session UID as path param.
type SessionIDParam struct {
	// UID is the session's UID.
	UID string `param:"uid" json:"-" validate:"required"`
}

// ListSessions is the request to page through a namespace's sessions, live and finished alike.
type ListSessions struct {
	TenantID string `identity:"tenant" json:"-"`
	query.Paginator
	query.Filters
}

// SessionGet is the request data for the get session endpoint. Timeline asks for the session's
// event timeline; the handler sets it when the caller may replay recordings or is on the instance
// admin console, so it is never read from the request and a client cannot ask for a timeline it may
// not read.
type SessionGet struct {
	SessionIDParam
	Timeline bool `json:"-" query:"-"`
}

// SessionCreate is the structure to represent the request data for create session endpoint.
type SessionCreate struct {
	UID       string `json:"uid" validate:"required"`
	DeviceUID string `json:"device_uid" validate:"required"`
	Username  string `json:"username" validate:"required"`
	IPAddress string `json:"ip_address" validate:"required"`
	Web       bool   `json:"web" validate:""`
	// UserID is the ShellHub account that authorized the session via browser
	// approval. Empty for password/public-key and web-terminal sessions.
	UserID string `json:"user_id" validate:""`
	// APIKeyID is the API key the session acts as, when an automation opened it. It is
	// separate from UserID because the two name different tables.
	APIKeyID string `json:"api_key_id" validate:""`
}
