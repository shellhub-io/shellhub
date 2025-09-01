package dialer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/multiformats/go-multistream"
)

type Target interface {
	prepare(conn net.Conn, version ConnectionVersion) (net.Conn, error)
}

// SSHOpenTarget prepares a connection for initiating a new SSH session
// with the agent.
type SSHOpenTarget struct{ SessionID string }

func (t SSHOpenTarget) prepare(conn net.Conn, version ConnectionVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case ConnectionVersion1:
		// Legacy: HTTP GET /ssh/<id>
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", t.SessionID), nil)
		if err := req.Write(conn); err != nil {
			return nil, err
		}
	case ConnectionVersion2:
		if err := multistream.SelectProtoOrFail(ProtoSSHOpen, conn); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(conn).Encode(map[string]string{"id": t.SessionID}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported connection version: %d", version)
	}

	return conn, nil
}

// SSHCloseTarget prepares a connection to request closing an existing SSH session.
type SSHCloseTarget struct{ SessionID string }

func (t SSHCloseTarget) prepare(conn net.Conn, version ConnectionVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case ConnectionVersion1:
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", t.SessionID), nil)
		if err := req.Write(conn); err != nil {
			return nil, err
		}
	case ConnectionVersion2:
		if err := multistream.SelectProtoOrFail(ProtoSSHClose, conn); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(conn).Encode(map[string]string{"id": t.SessionID}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported connection version: %d", version)
	}

	return conn, nil
}

// HTTPProxyTarget prepares a connection for proxying HTTP traffic to a
// device web endpoint. After preparation the caller should write the
// final HTTP request (with rewritten Host + URL) directly to the
// returned connection.
type HTTPProxyTarget struct {
	RequestID        string
	Host             string
	Port             int
	HandshakeRequest *http.Request // original inbound request used for V1 CONNECT-style handshake
}

func (t HTTPProxyTarget) prepare(conn net.Conn, version ConnectionVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case ConnectionVersion1:
		// Write initial handshake request and expect 200 OK.
		if err := t.HandshakeRequest.Write(conn); err != nil {
			return nil, err
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), t.HandshakeRequest)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http proxy handshake failed: %s", resp.Status)
		}
	case ConnectionVersion2:
		if err := multistream.SelectProtoOrFail(ProtoHTTPProxy, conn); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(conn).Encode(map[string]string{
			"id":   t.RequestID,
			"host": t.Host,
			"port": strconv.Itoa(t.Port),
		}); err != nil {
			return nil, err
		}
		result := map[string]string{}
		if err := json.NewDecoder(conn).Decode(&result); err != nil {
			return nil, err
		}
		if result["status"] != "ok" {
			return nil, fmt.Errorf("http proxy negotiation failed: %s", result["message"])
		}
	default:
		return nil, fmt.Errorf("unsupported connection version: %d", version)
	}

	return conn, nil
}
