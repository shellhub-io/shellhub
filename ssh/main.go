package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/web"
	log "github.com/sirupsen/logrus"
)

func init() {
	loglevel.SetLogLevel()
	log.SetFormatter(&log.JSONFormatter{})
}

type Envs struct {
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RecordURL      string        `env:"RECORD_URL,default=cloud-api:8080"`

	// RedisURI is the connection URI for accessing Redis.
	// Default value is "redis://redis:6379".
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`

	// RedisCachePoolSize sets the size of the connection pool available for Redis cache.
	// Default value is 0, meaning no pooling.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`

	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

func main() {
	env, err := envs.ParseWithPrefix[Envs]("SSH_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	cache, err := cache.NewRedisCache(env.RedisURI, env.RedisCachePoolSize)
	if err != nil {
		log.WithError(err).Error("Failed to configure redis store cache")
	}

	client := internalclient.New(internalclient.WithAsynq(env.RedisURI))
	defer client.Close()

	httptunnel := httptunnel.NewTunnel("/ssh/connection", "/ssh/revdial")

	tunnel := tunnel.New(httptunnel, cache, client)
	web.NewSSHServerBridge(tunnel.Router)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(tunnel.Router)

		log.Info("Profiling enabled at http://0.0.0.0:8080/debug/pprof/")
	}

	opts := &server.Options{
		ConnectTimeout:               env.ConnectTimeout,
		RecordURL:                    env.RecordURL,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
	}
	srv := server.NewServer(opts, tunnel.T)

	go func() {
		if err := http.ListenAndServe(":8080", tunnel.Router); err != nil {
			log.WithError(err).Fatal("failed to listen and serve the HTTP server")
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("failed to listen and serve the SSH server")
	}
}
