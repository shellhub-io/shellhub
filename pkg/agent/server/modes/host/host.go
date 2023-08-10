// Package host defines authentication and sessions handles when SSH's server is running in host mode.
package host

import "github.com/shellhub-io/shellhub/pkg/agent/server/modes"

// Features defines the features supported by the agent when running in host mode.
const Features modes.Features = modes.FeatureLocalPortForwarding |
	modes.FeatureShell |
	modes.FeatureHeredoc |
	modes.FeatureExec |
	modes.FeatureSFTP
