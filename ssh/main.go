package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
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
	RedisURI                     string        `env:"REDIS_URI,default=redis://redis:6379"`
	RecordURL                    string        `env:"RECORD_URL,default=cloud-api:8080"`
	ConnectTimeout               time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	AllowPublickeyAccessBelow060 bool          `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'.
	env, err := envs.ParseWithPrefix[Envs]("SSH_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	cache, err := cache.NewRedisCache(env.RedisURI, 0)
	if err != nil {
		log.WithError(err).
			Fatal("failed to connect to redis cache")
	}

	tun, err := tunnel.NewTunnel("/ssh/connection", "/ssh/revdial", env.RedisURI)
	if err != nil {
		log.WithError(err).
			Fatal("failed to create the internalclient")
	}

	router := tun.GetRouter()

	web.NewSSHServerBridge(router, cache)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(router)

		log.Info("Profiling enabled at http://0.0.0.0:8080/debug/pprof/")
	}

	go http.ListenAndServe(":8080", router) // nolint:errcheck,gosec

	log.Fatal(server.NewServer(&server.Options{
		ConnectTimeout:               env.ConnectTimeout,
		RecordURL:                    env.RecordURL,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
	}, tun.Tunnel, cache).ListenAndServe())
}
