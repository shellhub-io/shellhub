package main

import "errors"

var (
	ErrInvalidSessionTarget = errors.New("invalid session target")
	ErrBillingBlock         = errors.New("reached the device limit")
	ErrFirewallBlock        = errors.New("a firewall rule block this action")
	ErrFindDevice           = errors.New("cloud not find the device")
	ErrLookupDevice         = errors.New("could not lookup for device data")
)

// getExternalErrors converts an internal error for external one.
func getExternalError(err error) error {
	// External errors are intended to be returned for the end user.
	switch err {
	case ErrBillingBlock:
		return errors.New("reached the device limit, update to premium or choose up to 3 devices")
	default:
		return err
	}
}
