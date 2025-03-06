package session

import "fmt"

// Errors returned by the NewSession to the client.
var (
	ErrBillingBlock            = fmt.Errorf("Connection to this device is not available as your current namespace doesn't qualify for the free plan. To gain access, you'll need to contact the namespace owner to initiate an upgrade.\n\nFor a detailed estimate of costs based on your use-cases with ShellHub Cloud, visit our pricing page at https://www.shellhub.io/pricing. If you wish to upgrade immediately, navigate to https://cloud.shellhub.io/settings/billing. Your cooperation is appreciated.") //nolint:all
	ErrFirewallBlock           = fmt.Errorf("you cannot connect to this device because a firewall rule block your connection")
	ErrFirewallConnection      = fmt.Errorf("failed to communicate to the firewall")
	ErrFirewallUnknown         = fmt.Errorf("failed to evaluate the firewall rule")
	ErrHost                    = fmt.Errorf("failed to get the device address")
	ErrFindDevice              = fmt.Errorf("failed to find the device")
	ErrDial                    = fmt.Errorf("failed to connect to device agent, please check the device connection")
	ErrInvalidVersion          = fmt.Errorf("failed to parse device version")
	ErrUnsuportedPublicKeyAuth = fmt.Errorf("connections using public keys are not permitted when the agent version is 0.5.x or earlier")
	ErrUnexpectedAuthMethod    = fmt.Errorf("failed to authenticate the session due to a unexpected method")
	ErrEvaluatePublicKey       = fmt.Errorf("failed to evaluate the provided public key")
	ErrSeatAlreadySet          = fmt.Errorf("this seat was already set")
)
