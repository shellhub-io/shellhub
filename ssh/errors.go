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
		// revive:disable:error-strings
		//nolint:stylecheck,golint
		return errors.New("You cannot connect to this device because the namespace is not eligible for the free plan.\nPlease contact the namespace owner's to upgrade the plan.\nSee our pricing plans on https://www.shellhub.io/pricing to estimate the cost of your use cases on ShellHub Cloud or go to https://cloud.shellhub.io/settings/billing to upgrade the plan.")
		// revive:enable:error-strings
	default:
		return err
	}
}
