package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	sshTunnel "github.com/shellhub-io/shellhub/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	"github.com/shellhub-io/shellhub/ssh/web"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/cache"
	log "github.com/sirupsen/logrus"
)

func init() {
	loglevel.SetLogLevel()
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	// Populates configuration based on environment variables prefixed with 'SSH_'.
	env, err := envs.ParseWithPrefix[server.Options]("ssh")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	if err := cache.ConnectRedis(env.RedisURI); err != nil {
		log.WithError(err).Fatal("Failed to connect to redis")
	}

	options, err := asynq.ParseRedisURI(env.RedisURI)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse redis uri")
	}

	client := asynq.NewClient(options)
	if client == nil {
		log.WithError(err).Fatal("Failed to create asynq client")
	}

	tunnel := sshTunnel.NewTunnel("/ssh/connection", "/ssh/revdial")

	// withAsynq is a configuration function that sets the Asynq client to the API internal client.
	withAsynq := func(o *internalclient.Options) error {
		o.Asynq = client

		return nil
	}

	tunnel.API = internalclient.NewClient(withAsynq)

	router := tunnel.GetRouter()
	router.Any("/sessions/:uid/close", func(c echo.Context) error {
		exit := func(status int, err error) error {
			log.WithError(err).WithField("status", status).Error("failed to close the session")

			return c.JSON(status, err.Error())
		}

		uid := c.Param("uid")
		var closeRequest struct {
			Device string `json:"device"`
		}
		if err := c.Bind(&closeRequest); err != nil {
			return exit(http.StatusBadRequest, err)
		}

		conn, err := tunnel.Dial(context.Background(), closeRequest.Device)
		if err != nil {
			return exit(http.StatusInternalServerError, err)
		}

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", uid), nil)
		if err != nil {
			return exit(http.StatusInternalServerError, err)
		}

		if err := req.Write(conn); err != nil {
			return exit(http.StatusInternalServerError, err)
		}

		return c.NoContent(http.StatusOK)
	})

	router.Any("/ssh/http", func(c echo.Context) error {
		replyError := func(err error, msg string, code int) error {
			log.WithError(err).WithFields(log.Fields{
				"remote":  c.Request().RemoteAddr,
				"address": c.Request().Header.Get("X-Public-Address"),
				"path":    c.Request().Header.Get("X-Path"),
			}).Error(msg)

			return c.String(code, msg)
		}

		dev, err := tunnel.API.GetDeviceByPublicURLAddress(c.Request().Header.Get("X-Public-URL-Address"))
		if err != nil {
			return replyError(err, "failed to get device data", http.StatusInternalServerError)
		}

		if !dev.PublicURL {
			return replyError(err, "this device is not accessible via public URL", http.StatusForbidden)
		}

		in, err := tunnel.Dial(c.Request().Context(), dev.UID)
		if err != nil {
			return replyError(err, "failed to connect to device", http.StatusInternalServerError)
		}

		defer in.Close()

		if err := c.Request().Write(in); err != nil {
			return replyError(err, "failed to write request to device", http.StatusInternalServerError)
		}

		ctr := http.NewResponseController(c.Response())
		out, _, err := ctr.Hijack()
		if err != nil {
			return replyError(err, "failed to hijack response", http.StatusInternalServerError)
		}

		defer out.Close()
		if _, err := io.Copy(out, in); errors.Is(err, io.ErrUnexpectedEOF) {
			return replyError(err, "failed to copy response from device service to client", http.StatusInternalServerError)
		}

		return nil
	})

	// TODO: add `/ws/ssh` route to OpenAPI repository.
	router.GET("/ws/ssh", echo.WrapHandler(web.HandlerRestoreSession(web.RestoreSession, handler.WebSession)))
	router.POST("/ws/ssh", echo.WrapHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		web.HandlerCreateSession(web.CreateSession)(res, req)
	})))

	router.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	go http.ListenAndServe(":8080", router) // nolint:errcheck

	log.Fatal(server.NewServer(env, tunnel.Tunnel).ListenAndServe())
}
