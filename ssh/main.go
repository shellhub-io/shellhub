package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/ssh/http"
	"github.com/shellhub-io/shellhub/ssh/pkg/dialer"
	ssh "github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/web"
	log "github.com/sirupsen/logrus"
)

const ListenAddress = ":8080"

func init() {
	loglevel.SetLogLevel()
	log.SetFormatter(&log.JSONFormatter{})
}

type Envs struct {
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RedisURI       string        `env:"REDIS_URI,default=redis://redis:6379"`
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool   `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
	WebEndpoints                 bool   `env:"SHELLHUB_WEB_ENDPOINTS,default=false"`
	WebEndpointsDomain           string `env:"SHELLHUB_WEB_ENDPOINTS_DOMAIN"`
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

	cli, err := internalclient.NewClient(nil, internalclient.WithAsynqWorker(env.RedisURI))
	if err != nil {
		log.WithError(err).
			Fatal("failed to create the tunnel")
	}

	d := dialer.NewDialer(cli)

	h := http.NewServer(d, cli, &http.Config{
		WebEndpoints:       env.WebEndpoints,
		WebEndpointsDomain: env.WebEndpointsDomain,
	})

	router := h.Router

	web.NewSSHServerBridge(router, cache)

	if envs.IsDevelopment() {
		runtime.SetBlockProfileRate(1)
		pprof.Register(router)

		log.Info("Profiling enabled at http://0.0.0.0:8080/debug/pprof/")
	}

	s := ssh.NewServer(d, cache, &ssh.Options{
		ConnectTimeout:               env.ConnectTimeout,
		AllowPublickeyAccessBelow060: env.AllowPublickeyAccessBelow060,
	})

	errs := make(chan error)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debugf("listen for HTTP server on %s paniced", ListenAddress)

				errs <- fmt.Errorf("listen for HTTP on %s paniced", ListenAddress)
			}
		}()

		log.WithField("address", ListenAddress).Info("starting HTTP server")
		if err := h.ListenAndServe(ListenAddress); err != nil {
			log.WithError(err).WithField("address", ListenAddress).Error("HTTP server failed")
			errs <- err
		} else {
			log.WithField("address", ListenAddress).Info("HTTP server stopped gracefully")
			errs <- nil
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debugf("listen for SSH server paniced")

				errs <- fmt.Errorf("listen for SSH server paniced")
			}
		}()

		errs <- s.ListenAndServe()
	}()

	if err := <-errs; err != nil {
		log.WithError(err).Fatal("a fatal error was send from HTTP or SSH server")
	}

	log.Warn("ssh service is closed")
}
