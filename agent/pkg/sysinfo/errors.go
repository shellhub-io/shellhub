package sysinfo

import "errors"

// ErrNoInterfaceFound is returned when the host exposes no usable network interface to
// derive a device identity from.
var ErrNoInterfaceFound = errors.New("no interface found")
