package request

// SessionIDParam is a parameter that can be used to validate a session ID.
type SessionIDParam struct {
	// UID is the session's uid.
	UID string `param:"uid" validate:"required"`
}

// SessionGet is the structure for the request data at get session endpoint.
type SessionGet struct {
	SessionIDParam
}

// SessionAuthenticatedSet is the structure for the request data at set authenticated session endpoint.
type SessionAuthenticatedSet struct {
	SessionIDParam
	Authenticated bool `json:"authenticated" validate:"required"`
}

// SessionCreate is the structure for the request data at create session endpoint.
type SessionCreate struct {
	UID       string `json:"uid" validate:"required"`
	DeviceUID string `json:"device_uid" validate:"required"`
	Username  string `json:"username" validate:"required"`
	IPAddress string `json:"ip_address" validate:"required"`
	Type      string `json:"type" validate:"required"`
	Term      string `json:"term" validate:"required"`
}

// SessionFinish is the structure for the request data at finish session endpoint.
type SessionFinish struct {
	SessionIDParam
}

// SessionFinish is the structure for the request data at keep alive session endpoint.
type SessionKeepAlive struct {
	SessionIDParam
}
