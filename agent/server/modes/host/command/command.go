package command

// SFTPServerMode selects how an SFTP subsystem request is served.
type SFTPServerMode string

// Native re-executes the agent itself as the SFTP server; Docker runs it inside the
// container the session targets.
const (
	SFTPServerModeNative SFTPServerMode = "native"
	SFTPServerModeDocker SFTPServerMode = "docker"
)
