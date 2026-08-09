package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// RealIPExtractor resolves the client address from the X-Real-IP header, but only for requests
// that arrive from a trusted peer — a loopback, link-local or private address. Anything else is
// attributed to the address it actually connected from.
//
// This is deliberately not [echo.ExtractIPFromRealIPHeader]. That helper checks whether the
// address *inside* the header is trusted, so a header naming a public client — which is the only
// interesting case — is always rejected and every request ends up attributed to the gateway
// container. Trusting the peer that set the header is what the topology calls for: the gateway is
// on a private network and is the only thing entitled to speak for the client.
func RealIPExtractor() echo.IPExtractor {
	return func(req *http.Request) string {
		peer, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			peer = req.RemoteAddr
		}

		if ip := net.ParseIP(peer); ip == nil || !trustedPeer(ip) {
			return peer
		}

		realIP := strings.TrimSuffix(strings.TrimPrefix(req.Header.Get(echo.HeaderXRealIP), "["), "]")
		if net.ParseIP(realIP) == nil {
			return peer
		}

		return realIP
	}
}

func trustedPeer(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate()
}
