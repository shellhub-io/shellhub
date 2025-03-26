package session

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
)

type config struct {
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

// sshconf is a global variable responsible for managing all environment configurations.
var sshconf *config

func init() {
	var err error

	sshconf, err = envs.Parse[config]()
	if err != nil {
		log.WithError(err).
			Error("failed to parse the environment variables")
	}
}
