package main

import (
	"errors"
)

var (
	ErrInvalidSessionTarget = errors.New("invalid session target")
	ErrBillingBlock         = errors.New("reached the device limit, update to premium or choose up to 3 devices")
	ErrFirewallBlock        = errors.New("a firewall rule block this action")
	ErrFindDevice           = errors.New("cloud not find the device")
	ErrLookupDevice         = errors.New("could not lookup for device data")
)

// getExternalErrors returns a map with internal error as key and external error as value.
// watch out: on each new error, a new external must be created and associate on the returning map.
func getExternalErrors() map[error]error {
	// External errors are intended to be returned for the end user.
	var (
		ErrInvalidSessionTargetExternal = errors.New("internal error")
		ErrBillingBlockExternal         = errors.New("billing block")
		ErrFirewallBlockExternal        = errors.New("firewall block")
		ErrFindDeviceExternal           = errors.New("invalid device")
		ErrLookupDeviceExternal         = errors.New("device data error")
	)

	return map[error]error{
		ErrInvalidSessionTarget: ErrInvalidSessionTargetExternal,
		ErrBillingBlock:         ErrBillingBlockExternal,
		ErrFirewallBlock:        ErrFirewallBlockExternal,
		ErrFindDevice:           ErrFindDeviceExternal,
		ErrLookupDevice:         ErrLookupDeviceExternal,
	}
}

// ErrorExternal returns a external error from a internal one.
func ErrorExternal(err error) error {
	return getExternalErrors()[err]
}
