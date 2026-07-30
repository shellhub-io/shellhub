package models

// WebEndpointTLS carries how the HTTP proxy should reach the backend behind a web
// endpoint.
type WebEndpointTLS struct {
	Enabled bool `json:"enabled"`
	Verify  bool `json:"verify"`
	// Domain doubles as the Host header override and, with TLS enabled, the SNI
	// sent during the handshake.
	Domain string `json:"domain"`
}

// WebEndpoint is what the HTTP proxy needs to route a request to a device: which
// namespace and device to dial, and which backend address to ask that device for. It is
// deliberately narrower than the stored endpoint — the proxy has no use for its
// expiration or creation time.
type WebEndpoint struct {
	Address   string         `json:"address"`
	Namespace string         `json:"namespace"`
	DeviceUID string         `json:"device_uid"`
	Host      string         `json:"host"`
	Port      int            `json:"port"`
	TLS       WebEndpointTLS `json:"tls"`
}

// FirewallConnection describes the SSH connection attempt a firewall rule matches
// against.
type FirewallConnection struct {
	// Namespace is the namespace name, not its tenant ID.
	Namespace string `json:"namespace"`
	// Hostname is the device name within the namespace.
	Hostname string `json:"hostname"`
	// Username is the user being requested on the device, not the ShellHub user.
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
}
