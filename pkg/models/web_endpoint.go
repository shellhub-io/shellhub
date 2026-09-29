package models

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
