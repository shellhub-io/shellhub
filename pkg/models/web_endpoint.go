package models

// FirewallConnection describes an SSH connection attempt: firewall rules match
// IPAddress, the client's address, and the licensed regions apply to
// DeviceIPAddress, the address the device's agent last authenticated from.
type FirewallConnection struct {
	// Namespace is the namespace name, not its tenant ID.
	Namespace string `json:"namespace"`
	// Hostname is the device name within the namespace.
	Hostname string `json:"hostname"`
	// Username is the user being requested on the device, not the ShellHub user.
	Username        string `json:"username"`
	IPAddress       string `json:"ip_address"`
	DeviceIPAddress string `json:"device_ip_address"`
}
