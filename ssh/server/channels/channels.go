package channels

const (
	// DirectTCPIPChannel is the channel type for direct-tcpip channels like "local port forwarding" and "dynamic
	// application-level port forwarding".
	//
	// Local port forwarding is used to forward a port from the client to the server, and dynamic application-level
	// is a method for securely tunneling and routing network traffic through an SSH connection to access remote
	// resources as if they were local.
	//
	// Example of local port forwarding: `ssh -L 8080:localhost:80 user@sshid`.
	//
	// Example of dynamic application-level port forwarding: `ssh -D 1080 user@sshid`.
	DirectTCPIPChannel = "direct-tcpip"
	SessionChannel     = "session"
)
