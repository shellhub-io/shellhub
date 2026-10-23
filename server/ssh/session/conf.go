package session

import "time"

// Config carries the settings the session package needs from the SSH server.
type Config struct {
	// AllowPublickeyAccessBelow060 allows SSH to connect with an agent via a public key when the agent version is
	// less than 0.6.0. Agents 0.5.x or earlier do not validate the public key request and may panic.
	AllowPublickeyAccessBelow060 bool

	// ConsoleURL is the URL users reach the console at, with no trailing slash, which the browser
	// approval links shown in the terminal start with.
	ConsoleURL string

	// ConnectTimeout bounds the SSH handshake against the agent. Zero leaves it
	// unbounded.
	ConnectTimeout time.Duration
}

var sshconf = &Config{
	AllowPublickeyAccessBelow060: false,
	ConsoleURL:                   "http://localhost",
	ConnectTimeout:               0,
}

// Configure installs the package settings. It is called once from the SSH
// server's constructor, before any connection is served.
func Configure(cfg Config) {
	sshconf = &cfg
}
