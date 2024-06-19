package main

import (
	"os"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	loglevel.SetLogLevel()

	if value, ok := os.LookupEnv("SHELLHUB_ENV"); ok && value == "development" {
		log.SetLevel(log.TraceLevel)
		log.Debug("Log level set to Trace")
	} else {
		log.Debug("Log level default")
	}
}

// Config provides the configuration for the API service.
// The values are load from the system environment variables.
type Config struct {
	// MongoDB connection string (URI format)
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	// Redis connection string (URI format)
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize is the pool size of connections available for Redis cache.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`
	// Enable GeoIP feature.
	//
	// GeoIP features enable the ability to get the logitude and latitude of the client from the IP address.
	// The feature is disabled by default. To enable it, it is required to have a `MAXMIND` database license and feed it
	// to `SHELLHUB_MAXMIND_LICENSE` with it, and `SHELLHUB_GEOIP=true`.
	GeoIP bool `env:"GEOIP,default=false"`
	// Session record cleanup worker schedule
	SessionRecordCleanupSchedule string `env:"SESSION_RECORD_CLEANUP_SCHEDULE,default=@daily"`
	// Sentry DSN.
	SentryDSN string `env:"SENTRY_DSN,default="`
}

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := envs.ParseWithPrefix[Config]("API_")
		if err != nil {
			log.WithError(err).Fatal("failed to read the environemental variables")
		}

		if err := Server(cmd.Context(), *cfg); err != nil {
			log.WithError(err).Fatal("failed to start the server")
		}

		return nil
	},
}

func main() {
	rootCmd := &cobra.Command{Use: "api"}

	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.ExecuteContext(rootCmd.Context()); err != nil {
		log.Fatal(err)
	}
}
