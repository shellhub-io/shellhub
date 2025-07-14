package target

import (
	"errors"
	"strings"
)

var (
	ErrSplitTwoTarget   = errors.New("could not split the target into two parts")
	ErrSplitThreeTarget = errors.New("could not split the target into three parts")
	ErrSplitSSHID       = errors.New("could not split SSHID into namespace and hostname")
	ErrNotSSHID         = errors.New("target is not from SSHID type")
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
		return nil, ErrSplitTwoTarget
	}

	return &Target{Username: parts[USERNAME], Data: parts[DATA]}, nil
}

// IsSSHID checks if target is a SSHID or a device's ID.
func (t *Target) IsSSHID() bool {
	return strings.Contains(t.Data, ".")
}

type SSHID struct {
	Username  string
	Namespace string
	Device    string
	Container string
}

// HasContainer checks if the SSHID has a container part.
func (s *SSHID) HasContainer() bool {
	return s.Container != ""
}

func (s *SSHID) String() string {
	if s.Container != "" {
		return s.Username + "@" + s.Namespace + "." + s.Device + "." + s.Container
	}

	return s.Username + "@" + s.Namespace + "." + s.Device
}

const (
	SSHIDParts             = 2
	SSHIDPartsWitContainer = 3
)

// SSHMaxParts defines the maximum number of parts in an SSHID.
const SSHMaxParts = 3

// SplitSSHID splits the SSHID into namespace and hostname as lower strings.
func (t *Target) SplitSSHID() (*SSHID, error) {
	if !t.IsSSHID() {
		return nil, ErrNotSSHID
	}

	const NAMESPACE = 0
	const HOSTNAME = 1
	const CONTAINER = 2

	parts := strings.SplitN(t.Data, ".", SSHMaxParts)
	switch len(parts) {
	case SSHIDParts:
		return &SSHID{
			Username:  t.Username,
			Namespace: strings.ToLower(parts[NAMESPACE]),
			Device:    strings.ToLower(parts[HOSTNAME]),
		}, nil
	case SSHIDPartsWitContainer:
		return &SSHID{
			Username:  t.Username,
			Namespace: strings.ToLower(parts[NAMESPACE]),
			Device:    strings.ToLower(parts[HOSTNAME]),
			Container: strings.ToLower(parts[CONTAINER]),
		}, nil
	default:
		return nil, ErrSplitSSHID
	}
}
