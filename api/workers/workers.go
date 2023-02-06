package workers

import (
	"github.com/kelseyhightower/envconfig"
)

type Envs struct {
	MongoURI                      string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	RedisURI                      string `envconfig:"redis_uri" default:"redis://redis:6379"`
	SessionRecordCleanupSchedule  string `envconfig:"session_record_cleanup_schedule" default:"@daily"`
	SessionRecordCleanupRetention int    `envconfig:"record_retention" default:"0"`
}

func getEnvs() (*Envs, error) {
	var envs Envs
	if err := envconfig.Process("api", &envs); err != nil {
		return nil, err
	}

	return &envs, nil
}
