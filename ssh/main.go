package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/ssh/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	loglevel.SetLogLevel()
	log.SetFormatter(&log.JSONFormatter{})
}

type Envs struct {
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RedisURI       string        `env:"REDIS_URI,default=redis://redis:6379"`
}

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'.
	env, err := envs.ParseWithPrefix[Envs]("SSH_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	tun := tunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tun.API = internalclient.NewClientWithAsynq(env.RedisURI)
	if tun.API == nil {
		log.Fatal("failed to create internal client")
	}

	router := tun.GetRouter().(*echo.Echo)

	router.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(router)

		log.Info("Profiling enabled at http://0.0.0.0:8080/debug/pprof/")
	}

	cherr := make(chan error)
	go func() {
		if err := http.ListenAndServe(":8080", router); err != nil {
			cherr <- err
		}

		cherr <- nil
	}()

	go func() {
		srv := server.NewServer(&server.Options{
			ConnectTimeout: env.ConnectTimeout,
			RedisURI:       env.RedisURI,
		}, tun.Tunnel)
		if err := srv.ListenAndServe(); err != nil {
			cherr <- err
		}

		cherr <- nil
	}()

	// NOTICE: the HTTP server and SSH server must run in parallel. The first one is responsible for receiving the
	// device's requests and the second one is responsible for receiving the SSH connections per se. They cannot run
	// without each other. Due to this, as soon as one of them fails, the whole service must be stopped.
	err = <-cherr
	if err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}

	log.Info("Server stopped")
}
