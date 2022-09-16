package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	sshTunnel "github.com/shellhub-io/shellhub/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/server/handler"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

const (
	ServerRouterAddress = ":8080"
	ServerSSHAddress    = "tcp://emq:1883"
	ServerSSHBroker     = ":2222"
)

func init() {
	loglevel.SetLogLevel()
}

func main() {
	tunnel := sshTunnel.NewTunnel("/ssh/connection", "/ssh/revdial")

	router := tunnel.GetRouter()
	router.HandleFunc("/sessions/{uid}/close", func(response http.ResponseWriter, request *http.Request) {
		exit := func(response http.ResponseWriter, status int, err error) {
			log.WithError(err).WithFields(log.Fields{
				"status": status,
			}).Error("failed to close the session")

			http.Error(response, err.Error(), status)
		}

		vars := mux.Vars(request)
		decoder := json.NewDecoder(request.Body)
		var closeRequest struct {
			Device string `json:"device"`
		}

		if err := decoder.Decode(&closeRequest); err != nil {
			exit(response, http.StatusBadRequest, err)

			return
		}

		conn, err := tunnel.Dial(context.Background(), closeRequest.Device)
		if err != nil {
			exit(response, http.StatusInternalServerError, err)

			return
		}

		request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", vars["uid"]), nil)
		if err := request.Write(conn); err != nil {
			exit(response, http.StatusInternalServerError, err)

			return
		}
	})
	router.Handle("/ws/ssh", websocket.Handler(handler.WebSession))

	go http.ListenAndServe(ServerRouterAddress, router) // nolint:errcheck

	log.Fatal(server.NewServer(&server.Options{
		Addr:           ServerSSHBroker,
		Broker:         ServerSSHAddress,
		ConnectTimeout: 30 * time.Second,
	}, tunnel.Tunnel).ListenAndServe())
}
