package workers

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Envs struct {
	MongoURI                      string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	RedisURI                      string `envconfig:"redis_uri" default:"redis://redis:6379"`
	SessionRecordCleanupSchedule  string `envconfig:"session_record_cleanup_schedule" default:"@daily"`
	SessionRecordCleanupRetention int    `envconfig:"record_retention" default:"0"`
}

var envs *Envs

func loadEnvs() error {
	var local Envs
	if envs == nil {
		if err := envconfig.Process("api", &local); err != nil {
			envs = &local

			return err
		}
	}

	return nil
}

func getEnvs() (*Envs, error) {
	if envs == nil {
		if err := loadEnvs(); err != nil {
			return nil, err
		}
	}

	return envs, nil
}

// Setup loads the essentials data to run the workers.
func Setup() error {
	if err := loadEnvs(); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	return nil
}
