package target

import (
	"errors"
	"strings"
)

var (
	ErrSplitTarget = errors.New("could not split the target into two parts")
	ErrNotSSHID    = errors.New("target is not from SSHID type")
)

type Target struct {
	Username string
	Data     string
}

func NewTarget(sshid string) (*Target, error) {
	// SSHID could be either device ID or a SSHID.
	//
	// Example: username@namespace.00-00-00-00-00-00@localhost.
	// 'username' is the user on the device.
	// 'namespace' is the user's namespace in ShellHub.
	// '00-00-00-00-00' is the device's hostname in ShellHub.
	// 'localhost' is the server's address.
	const USERNAME = 0
	const DATA = 1

	parts := strings.SplitN(sshid, "@", 2)
	if len(parts) != 2 {
		return nil, ErrSplitTarget
	}

	return &Target{Username: parts[USERNAME], Data: parts[DATA]}, nil
}

// IsSSHID checks if target is a SSHID or a device's ID.
func (t *Target) IsSSHID() bool {
	return strings.Contains(t.Data, ".")
}

// SplitSSHID splits the SSHID into namespace and hostname as lower strings.
// Namespace is the device's namespace and hostname is the device's name.
func (t *Target) SplitSSHID() (string, string, error) {
	if !t.IsSSHID() {
		return "", "", ErrNotSSHID
	}

	const NAMESPACE = 0
	const HOSTNAME = 1

	parts := strings.SplitN(t.Data, ".", 2)
	if len(parts) != 2 {
		return "", "", ErrSplitTarget
	}

	return parts[NAMESPACE], parts[HOSTNAME], nil
}
