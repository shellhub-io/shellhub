package session

import "time"

// Config carries the settings the session package needs from the SSH server.
type Config struct {
	// AllowPublickeyAccessBelow060 allows SSH to connect with an agent via a public key when the agent version is
	// less than 0.6.0. Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool

	// Domain is the instance's base domain, used to build the browser approval URL
	// shown in the terminal when a namespace requires SSH login approval.
	Domain string

	// AutoSSL reports whether the instance serves the console over HTTPS; it
	// selects the scheme of the approval URL.
	AutoSSL bool

	// ConnectTimeout bounds the SSH handshake against the agent. Zero leaves it
	// unbounded.
	ConnectTimeout time.Duration
}

// sshconf holds the settings installed by [Configure].
//
// It is never nil: a test that does not configure the package reads these
// defaults rather than dereferencing nothing.
var sshconf = &Config{
	AllowPublickeyAccessBelow060: false,
	Domain:                       "localhost",
	AutoSSL:                      false,
	ConnectTimeout:               0,
}

// Configure installs the package settings. It is called once from the SSH
// server's constructor, before any connection is served.
func Configure(cfg Config) {
	sshconf = &cfg
}
