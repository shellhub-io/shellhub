package internalclient

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
)

type env struct {
	ServerPort string `env:"API_PORT,default=8080" validate:"required,numeric"`
}

func getEnvs() (*env, error) {
	env, err := envs.ParseWithPrefix[env]("API_")
	if err != nil {
		log.WithError(err).Error("Failed to parse environment variables with prefix 'api'")

		return nil, err
	}

	return env, nil
}
