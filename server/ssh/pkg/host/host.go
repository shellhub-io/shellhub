package host

import "net"

type Host struct {
	Host string
}

func NewHost(address string) (*Host, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	return &Host{Host: host}, nil
}

// IsLocalhost checks if host address is localhost.
func (h *Host) IsLocalhost() bool {
	return h.Host == "127.0.0.1" || h.Host == "::1"
}
