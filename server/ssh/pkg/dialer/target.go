package dialer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/multiformats/go-multistream"
	log "github.com/sirupsen/logrus"
)

type Target interface {
	prepare(conn net.Conn, version TransportVersion) (net.Conn, error)
}

// SSHOpenTarget prepares a connection for initiating a new SSH session
// with the agent.
type SSHOpenTarget struct{ SessionID string }

func (t SSHOpenTarget) prepare(conn net.Conn, version TransportVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case TransportVersion1:
		log.Debug("preparing SSH open target for transport version 1")

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", t.SessionID), nil)
		if err := req.Write(conn); err != nil {
			log.Errorf("failed to write HTTP request: %v", err)

			return nil, err
		}
	case TransportVersion2:
		log.Debug("preparing SSH open target for transport version 2")

		if err := multistream.SelectProtoOrFail(ProtoSSHOpen, conn); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(conn).Encode(map[string]string{"id": t.SessionID}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported transport version: %d", version)
	}

	return conn, nil
}

// SSHCloseTarget prepares a connection to request closing an existing SSH session.
type SSHCloseTarget struct{ SessionID string }

func (t SSHCloseTarget) prepare(conn net.Conn, version TransportVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case TransportVersion1:
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", t.SessionID), nil)
		if err := req.Write(conn); err != nil {
			return nil, err
		}
	case TransportVersion2:
		if err := multistream.SelectProtoOrFail(ProtoSSHClose, conn); err != nil {
			return nil, err
		}
		if err := json.NewEncoder(conn).Encode(map[string]string{"id": t.SessionID}); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported transport version: %d", version)
	}

	return conn, nil
}

// HTTPProxyTarget prepares a connection for proxying HTTP traffic to a
// device web endpoint. After preparation the caller should write the
// final HTTP request (with rewritten Host + URL) directly to the
// returned connection.
type HTTPProxyTarget struct {
	RequestID string
	Host      string
	Port      int
}

func (t HTTPProxyTarget) prepare(conn net.Conn, version TransportVersion) (net.Conn, error) { // nolint:ireturn
	switch version {
	case TransportVersion1:
		// Write initial handshake request and expect 200 OK.
		handshakeReq, _ := http.NewRequest(http.MethodConnect, fmt.Sprintf("/http/proxy/%s:%d", t.Host, t.Port), nil)
		if err := handshakeReq.Write(conn); err != nil {
			return nil, err
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), handshakeReq)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http proxy handshake failed: %s", resp.Status)
		}
	case TransportVersion2:
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

		// NOTE: limit the size of the response to avoid DoS via large payloads.
		const Limit = 512
		if err := json.NewDecoder(io.LimitReader(conn, Limit)).Decode(&result); err != nil {
			return nil, err
		}
		if result["status"] != "ok" {
			return nil, fmt.Errorf("http proxy negotiation failed: %s", result["message"])
		}
	default:
		return nil, fmt.Errorf("unsupported transport version: %d", version)
	}

	return conn, nil
}
