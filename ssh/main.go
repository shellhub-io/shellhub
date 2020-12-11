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

	"github.com/parnurzeal/gorequest"
	api "github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
)

var magicKey *rsa.PrivateKey

func main() {
	opts := &Options{
		Addr:           ":2222",
		Broker:         "tcp://emq:1883",
		ConnectTimeout: 30 * time.Second,
	}

	tunnel := httptunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tunnel.ConnectionHandler = func(r *http.Request) (string, error) {
		uid := r.Header.Get(api.DeviceUIDHeader)
		return uid, nil
	}

	router := tunnel.Router().(*mux.Router)
	router.HandleFunc("/sessions/{uid}/close", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		decoder := json.NewDecoder(req.Body)
		var closeRequest struct {
			Device string `json:"device"`
		}
		err := decoder.Decode(&closeRequest)
		if err != nil {
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

	go http.ListenAndServe(":8080", router)

	server := NewServer(opts, tunnel)

	go func() {
		for {
			id, online := tunnel.Online()
			if !online {
				_, _, _ = gorequest.New().Post(fmt.Sprintf("http://api:8080/internal/devices/%s/offline", id)).End()
			}
		}
	}()

	var err error
	magicKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Fatal(server.ListenAndServe())
}
