package main

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
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
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RedisURI       string        `env:"REDIS_URI,default=redis://redis:6379"`
	RecordURL      string        `env:"RECORD_URL,default=cloud-api:8080"`
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'.
	env, err := envs.ParseWithPrefix[Envs]("SSH_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	tun := tunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tun.API = internalclient.New()
	if tun.API == nil {
		log.Fatal("failed to create internal client")
	}

	router := tun.GetRouter()

	web.NewSSHServerBridge(router)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(router)

		log.Info("Profiling enabled at http://0.0.0.0:8080/debug/pprof/")
	}

	go http.ListenAndServe(":8080", router) // nolint:errcheck

	log.Fatal(server.NewServer(&server.Options{
		ConnectTimeout:               env.ConnectTimeout,
		RecordURL:                    env.RecordURL,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
	}, tun.Tunnel).ListenAndServe())
}
