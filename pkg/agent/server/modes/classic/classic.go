// Package classic defines handlers for authentication and session management for SSH connections when it is running in
// classic mode.

// Classic mode means that the Agent's SSH server runs for the host machine, being compiled or embedded inside a Docker
// Container, using the host "/etc/passwd", "/etc/shadow", redirecting the SSH's connection to the devices STDIN, STDOUT
// and STDERR and other things needed to run the SSH'sserver in the host machine.
package classic

type Mode struct {
	Authenticator
	Sessioner
}
