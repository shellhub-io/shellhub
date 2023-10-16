package workers

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/envs"
)

type Envs struct {
	MongoURI                      string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	RedisURI                      string `envconfig:"redis_uri" default:"redis://redis:6379"`
	SessionRecordCleanupSchedule  string `envconfig:"session_record_cleanup_schedule" default:"@daily"`
	SessionRecordCleanupRetention int    `envconfig:"record_retention" default:"0"`
	// AsynqGroupMaxDelay is the maximum duration to wait before processing a group of tasks.
	//
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxDelay int `envconfig:"asynq_group_max_delay" default:"1"`
	// AsynqGroupGracePeriod is the grace period has configurable upper bound: you can set a maximum aggregation delay, after which Asynq server
	// will aggregate the tasks regardless of the remaining grace period.
	///
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupGracePeriod int64 `envconfig:"asynq_group_grace_period" default:"1"`
	// AsynqGroupMaxSize is the maximum number of tasks that can be aggregated together. If that number is reached, Asynq
	// server will aggregate the tasks immediately.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxSize int `envconfig:"asynq_group_max_size" default:"500"`
}

func getEnvs() (*Envs, error) {
	env, err := envs.ParseWithPrefix[Envs]("api")
	if err != nil {
		return nil, fmt.Errorf("failed to get the envs: %w", err)
	}

	return env, nil
}
