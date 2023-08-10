// Package mode defines the interfaces used by the server to determine how to handle authentication and sessions.
package modes

import gliderssh "github.com/gliderlabs/ssh"

// Features defines the features supported by the agent.
type Features uint32

// List of features supported by the agent using feature flags.
const (
	// FeatureLocalPortForwarding is the feature flag for local port forwarding.
	FeatureLocalPortForwarding Features = 1 << iota
	// FeatureReversePortForwarding is the feature flag for remote port forwarding.
	FeatureReversePortForwarding
	// FeatureShell is the feature flag for shell session.
	FeatureShell
	// FeatureHeredoc is the feature flag for heredoc session.
	FeatureHeredoc
	// FeatureExec is the feature flag for exec session.
	FeatureExec
	// FeatureSFTP is the feature flag for SFTP session.
	FeatureSFTP
)

// IsFeatureEnabled checks if the feature flag is enabled.
func (f Features) IsFeatureEnabled(flag Features) bool {
	return (f & flag) != 0
}

// Mode defines the SSH's server mode type.
type Mode string

const (
	// HostMode represents the SSH's server host mode.
	//
	// HostMode mode means that the SSH's server runs in the host machine, using the host `/etc/passwd`, `/etc/shadow`,
	// redirecting the SSH's connection to the device sdin, stdout and stderr and etc.
	HostMode Mode = "host"
)

// Authenticator defines the authentication methods used by the SSH's server.
type Authenticator interface {
	// Password must be implemented to deal with password authentication.
	Password(ctx gliderssh.Context, user string, password string) bool
	// PublicKey must be implemented to deal with public key authentication.
	PublicKey(ctx gliderssh.Context, user string, key gliderssh.PublicKey) bool
}

// Sessioner defines the session methods used by the SSH's server to deal wihth determining the type of session.
type Sessioner interface {
	Subsystemer
	// Shell must be implemented to deal with shell session.
	Shell(session gliderssh.Session) error
	// Heredoc must be implemented to deal with heredoc session.
	//
	// heredoc is special block of code that contains multi-line strings that will be redirected to a stdin of a shell.
	// It request a shell, but doesn't allocate a pty.
	Heredoc(session gliderssh.Session) error
	// Exec must be implemented to deal with exec session.
	Exec(session gliderssh.Session) error
}

type Subsystemer interface {
	// SFTP must be implemented to deal with SFTP session.
	SFTP(session gliderssh.Session) error
}
