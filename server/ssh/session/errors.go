package session

import (
	"errors"
	"fmt"
)

// Errors returned by the NewSession to the client.
var (
	ErrBillingBlock            = fmt.Errorf("Connection to this device is not available as your current namespace doesn't qualify for the free plan. To gain access, you'll need to contact the namespace owner to initiate an upgrade.\n\nFor a detailed estimate of costs based on your use-cases with ShellHub Cloud, visit our pricing page at https://www.shellhub.io/pricing. If you wish to upgrade immediately, navigate to https://cloud.shellhub.io/settings/billing. Your cooperation is appreciated.") //nolint:all
	ErrFirewallBlock           = errors.New("you cannot connect to this device because a firewall rule block your connection")
	ErrFirewallUnknown         = errors.New("failed to evaluate the firewall rule")
	ErrHost                    = errors.New("failed to get the device address")
	ErrFindDevice              = errors.New("failed to find the device")
	ErrDial                    = errors.New("failed to connect to device agent, please check the device connection")
	ErrInvalidVersion          = errors.New("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = errors.New("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
	ErrUnexpectedAuthMethod    = errors.New("failed to authenticate the session due to a unexpected method")
	ErrEvaluatePublicKey       = errors.New("failed to evaluate the provided public key")
	ErrSeatAlreadySet          = errors.New("this seat was already set")
	ErrSeatNotFound            = errors.New("this seat has no agent channel")
	ErrWebData                 = errors.New("failed to parse the web session data")
	ErrLicenseBlock            = fmt.Errorf("Connection blocked: your ShellHub instance has exceeded the maximum number of devices allowed by your license. Please contact support or remove unused devices.") //nolint:all
)
