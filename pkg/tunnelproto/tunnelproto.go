// Package tunnelproto is the wire protocol between the agent and the server over the tunnel an
// agent holds open. Both sides import it, so a value here changes on both at once; a changed value
// is still a wire-incompatible change for released agents.
package tunnelproto

// The endpoints an agent opens its tunnel at. V1 carries a revdial connection, whose every logical
// stream needs one more HTTP handshake at RevdialPath; V2 carries yamux, whose streams are
// negotiated with multistream-select.
const (
	ConnectionV1Path = "/ssh/connection"
	ConnectionV2Path = "/agent/connection"
	RevdialPath      = "/ssh/revdial"
)

// The multistream protocols of a V2 stream.
const (
	ProtoSSHOpen   = "/ssh/open/1.0.0"
	ProtoSSHClose  = "/ssh/close/1.0.0"
	ProtoHTTPProxy = "/http/proxy/1.0.0"
)

// The request paths of a V1 stream. A session path ends with the session ID; the proxy path, sent
// with CONNECT, with the host:port the agent dials.
const (
	SSHOpenPathV1   = "/ssh/"
	SSHClosePathV1  = "/ssh/close/"
	HTTPProxyPathV1 = "/http/proxy/"
)

// Header is the JSON frame the server writes on a V2 stream before its payload. ID names the
// session or the proxied request; Host and Port are the proxy's target.
type Header struct {
	ID   string `json:"id"`
	Host string `json:"host,omitempty"`
	Port string `json:"port,omitempty"`
}

// StatusOK is the [Reply.Status] of a proxy stream whose target the agent reached.
const StatusOK = "ok"

// Reply is the JSON frame the agent answers a V2 proxy stream with before the proxied bytes: Status
// StatusOK, or the Error that kept it from reaching the target. The encoder ends the frame with a
// newline the reader must skip.
type Reply struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}
