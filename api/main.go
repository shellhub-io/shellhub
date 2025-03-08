package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type env struct {
	// MongoURI specifies the connection string for MongoDB.
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`

	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUser specifies the username for authenticate PostgreSQL.
	PostgresUser string `env:"POSTGRES_USER,default=admin"`
	// PostgresUser specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDB especifica o nome do banco de dados PostgreSQL a ser utilizado.
	PostgresDB string `env:"POSTGRES_DB,default=main"`

	// RedisURI specifies the connection string for Redis.
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize defines the maximum number of concurrent connections to Redis cache.
	// Set to 0 for unlimited connections.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`

	// AsynqGroupMaxDelay specifies the maximum time (in seconds) to wait before
	// processing a group of tasks, regardless of other conditions.
	AsynqGroupMaxDelay int `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod defines the grace period (in seconds) before task aggregation.
	// Tasks arriving within this period will be aggregated with existing tasks in the group.
	AsynqGroupGracePeriod int64 `env:"ASYNQ_GROUP_GRACE_PERIOD,default=2"`
	// AsynqGroupMaxSize specifies the maximum number of tasks that can be aggregated in a group.
	// When this limit is reached, the group will be processed immediately.
	AsynqGroupMaxSize int `env:"ASYNQ_GROUP_MAX_SIZE,default=1000"`
	// AsynqUniquenessTimeout defines how long (in hours) a unique job remains locked in the queue.
	// If a job doesn't complete within this period, its lock is released, allowing a new instance
	// to be enqueued and executed.
	AsynqUniquenessTimeout int `env:"ASYNQ_UNIQUENESS_TIMEOUT,default=24"`

	// SentryDSN specifies the Data Source Name for Sentry error tracking.
	// Leave empty to disable Sentry integration.
	SentryDSN string `env:"SENTRY_DSN,default="`

	// GeoipMirror specifies an alternative URL for downloading GeoIP databases.
	// When configured, this takes precedence over GeoipMaxmindLicense.
	GeoipMirror string `env:"MAXMIND_MIRROR,default="`
	// GeoipMaxmindLicense is the MaxMind license key for downloading GeoIP databases directly.
	// This is used as a fallback when GeoipMirror is not configured.
	GeoipMaxmindLicense string `env:"MAXMIND_LICENSE,default="`
}

func main() {
	loglevel.UseEnvs()

	rootCmd := &cobra.Command{Use: "api"}
	rootCmd.AddCommand(&cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := envs.ParseWithPrefix[env]("API_")
			if err != nil {
				log.WithError(err).
					Error("Failed to load environment variables")

				return err
			}

			server := &Server{env: env}

			if err := server.Setup(cmd.Context()); err != nil {
				log.WithError(err).
					Error("failed to setup the server")

				return err
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigs
				log.WithField("signal", sig).
					Info("shutting down the server")

				server.Shutdown()
				os.Exit(0)
			}()

			if err := server.Start(); err != nil {
				log.WithError(err).
					Error("failed too start the server")

				return err
			}

			return nil
		},
	})

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.WithError(err).
			Error("failed to execute command")

		os.Exit(1)
	}
}
