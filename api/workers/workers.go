package workers

import "github.com/shellhub-io/shellhub/pkg/envs"

type Envs struct {
	MongoURI                      string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	RedisURI                      string `env:"REDIS_URI,default=redis://redis:6379"`
	SessionRecordCleanupSchedule  string `env:"SESSION_RECORD_CLEANUP_SCHEDULE,default=@daily"`
	SessionRecordCleanupRetention int    `env:"RECORD_RETENTION,default=0"`
	// AsynqGroupMaxDelay is the maximum duration to wait before processing a group of tasks.
	//
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxDelay int `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod is the grace period has configurable upper bound: you can set a maximum aggregation delay, after which Asynq server
	// will aggregate the tasks regardless of the remaining grace period.
	///
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupGracePeriod int64 `env:"ASYNQ_GROUP_GRACE_PERIOD,default=1"`
	// AsynqGroupMaxSize is the maximum number of tasks that can be aggregated together. If that number is reached, Asynq
	// server will aggregate the tasks immediately.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxSize int `env:"ASYNQ_GROUP_MAX_SIZE,default=500"`
}

func getEnvs() (*Envs, error) {
	env, err := envs.ParseWithPrefix[Envs]("API_")
	if err != nil {
		return nil, err
	}

	return env, nil
}
