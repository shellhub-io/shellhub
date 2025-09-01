package dialer

// Multistream protocol identifiers used when negotiating per-stream
// application protocols over a V2 yamux connection.
//
// The agent and server must keep these values in sync. Changing a value
// is a wire incompatible change.
const (
	ProtoSSHOpen   = "/ssh/open/1.0.0"
	ProtoSSHClose  = "/ssh/close/1.0.0"
	ProtoHTTPProxy = "/http/proxy/1.0.0"
)
