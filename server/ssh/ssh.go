package ssh

import (
	"time"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/server/ssh/server"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	log "github.com/sirupsen/logrus"
)

type Env struct {
	ConnectTimeout               time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RedisURI                     string        `env:"REDIS_URI,default=redis://redis:6379"`
	AllowPublickeyAccessBelow060 bool          `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

type Server struct {
	env   *Env
	cache cache.Cache
	tun   *tunnel.Tunnel
}

func New(router *echo.Echo) (*Server, error) {
	env, err := envs.ParseWithPrefix[Env]("SSH_")
	if err != nil {
		return nil, err
	}

	cache, err := cache.NewRedisCache(env.RedisURI, 0)
	if err != nil {
		return nil, err
	}

	tun, err := tunnel.NewTunnel("/ssh/connection", "/ssh/revdial", env.RedisURI)
	if err != nil {
		return nil, err
	}

	web.NewSSHServerBridge(router, cache)

	if envs.IsDevelopment() {
		pprof.Register(router)
	}

	return &Server{
		env:   env,
		cache: cache,
		tun:   tun,
	}, nil
}

func (s *Server) Start() error {
	return server.NewServer(&server.Options{
		ConnectTimeout:               s.env.ConnectTimeout,
		AllowPublickeyAccessBelow060: s.env.AllowPublickeyAccessBelow060,
	}, s.tun.Tunnel, s.cache).ListenAndServe()
}

func (s *Server) Shutdown() {
	log.Info("SSH server shutdown (noop)")
}
