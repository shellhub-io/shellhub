package responses

// Error is the body of every API error response.
//
// It is a projection, not a marshalling of the internal error type: that type carries a layer name
// and an internal code which must never reach a client.
type Error struct {
	// Message is a human-readable description of what failed. It is always present.
	Message string `json:"message"`
	// Fields maps a request field name to the reason it was rejected. It is present only when the
	// error carries per-field detail.
	Fields map[string]string `json:"fields,omitempty"`
}
