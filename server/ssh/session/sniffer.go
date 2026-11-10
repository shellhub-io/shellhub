package session

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"net"
	"slices"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/models"
)

const maxVersionLine = 255

type versionSniffer struct {
	net.Conn

	mu      sync.Mutex
	line    []byte
	version []byte
}

func sniffVersion(conn net.Conn) *versionSniffer {
	return &versionSniffer{Conn: conn}
}

func (v *versionSniffer) Read(p []byte) (int, error) {
	n, err := v.Conn.Read(p)

	v.mu.Lock()
	defer v.mu.Unlock()

	for _, b := range p[:n] {
		if v.version != nil {
			break
		}

		if b != '\n' {
			if len(v.line) < maxVersionLine {
				v.line = append(v.line, b)
			}

			continue
		}

		if bytes.HasPrefix(v.line, []byte("SSH-")) {
			v.version = bytes.TrimSuffix(v.line, []byte("\r"))
		}

		v.line = nil
	}

	return n, err
}

func (v *versionSniffer) admitsGatewayLogin() bool {
	return v.announces(models.AgentCapabilityGatewayLogin)
}

func (v *versionSniffer) startsShellOnOpen() bool {
	return v.announces(models.AgentCapabilityShellOnOpen)
}

func (v *versionSniffer) announces(capability string) bool {
	if v == nil {
		return false
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	_, comment, ok := bytes.Cut(v.version, []byte(" "))
	if !ok {
		return false
	}

	return slices.ContainsFunc(bytes.Fields(comment), func(word []byte) bool {
		return string(word) == capability
	})
}

func newAgentLogin() string {
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)

	return base64.RawURLEncoding.EncodeToString(secret)
}
