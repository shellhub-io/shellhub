package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/session"
	"github.com/shellhub-io/shellhub/ssh/tunnel"
	websocketHelper "github.com/shellhub-io/shellhub/ssh/websocket"
	"golang.org/x/net/websocket"
)

const (
	SSHPort    = ":2222"
	TunnelPort = ":8080"
)

func main() {
	httpTunnel := tunnel.CreateTunnel("/ssh/connection", "/ssh/revdial")
	httpTunnelRouter := tunnel.GetTunnelRouter(httpTunnel)
	httpTunnelRouter.HandleFunc("/sessions/{uid}/close", session.HandlerSessionClose(httpTunnel))
	httpTunnelRouter.Handle("/ws/ssh", websocket.Handler(websocketHelper.HandlerWebsocket))

	go func() {
		err := http.ListenAndServe(TunnelPort, httpTunnelRouter)
		if err != nil {
			logrus.WithError(err).Fatal("could not init the server for tunneling")
		}
	}()

	if err := server.NewServer(&server.Options{
		Addr:           SSHPort,
		Broker:         "tcp://emq:1883",
		ConnectTimeout: 30 * time.Second,
	}, httpTunnel).ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("could not init the main server")
	}
}
