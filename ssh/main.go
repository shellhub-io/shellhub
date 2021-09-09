package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"

	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
)

var magicKey *rsa.PrivateKey

type Options struct {
	Addr           string
	Broker         string
	ConnectTimeout time.Duration
}

func main() {
	tunnel := httptunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tunnel.ConnectionHandler = func(r *http.Request) (string, error) {
		return r.Header.Get(client.DeviceUIDHeader), nil
	}
	tunnel.CloseHandler = func(id string) {
		if err := client.NewClient().DevicesOffline(id); err != nil {
			logrus.Error(err)
		}
	}

	router := tunnel.Router().(*mux.Router)
	router.HandleFunc("/sessions/{uid}/close", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		decoder := json.NewDecoder(req.Body)
		var closeRequest struct {
			Device string `json:"device"`
		}

		if err := decoder.Decode(&closeRequest); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)

			return
		}

		conn, err := tunnel.Dial(context.Background(), closeRequest.Device)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)

			return
		}

		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", vars["uid"]), nil)
		if err := req.Write(conn); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)

			return
		}
	})
	router.Handle("/ws/ssh", websocket.Handler(HandlerWebsocket))

	go http.ListenAndServe(":8080", router) // nolint:errcheck

	var err error
	magicKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Fatal(NewServer(&Options{
		Addr:           ":2222",
		Broker:         "tcp://emq:1883",
		ConnectTimeout: 30 * time.Second,
	}, tunnel).ListenAndServe())
}
