package main

import (
	"net/http"
	"runtime"

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

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'.
	env, err := envs.ParseWithPrefix[server.Options]("SSH_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	tun := tunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tun.API = internalclient.NewClientWithAsynq(env.RedisURI)
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

	log.Fatal(server.NewServer(env, tun.Tunnel).ListenAndServe())
}
