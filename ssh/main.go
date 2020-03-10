package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
)

func main() {
	opts := &Options{
		Addr:           ":2222",
		Broker:         "tcp://emq:1883",
		ConnectTimeout: 30 * time.Second,
	}

	tunnel := httptunnel.NewTunnel("/ssh/connection", "/ssh/revdial")
	tunnel.ConnectionHandler = func(r *http.Request) (string, error) {
		uid := r.Header.Get("X-Device-UID")
		return uid, nil
	}

	router := tunnel.Router().(*mux.Router)
	router.HandleFunc("/api/session/{uid}/close", func(res http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		decoder := json.NewDecoder(req.Body)
		var closeRequest struct {
			Device string `json:"device"`
		}
		err := decoder.Decode(&closeRequest)
		if err != nil {
			http.Error(res, err, http.StatusBadRequest)
			return
		}

		conn, err := tunnel.Dial(context.Background(), closeRequest.Device)
		if err != nil {
			http.Error(res, err, http.StatusBadRequest)
			return
		}
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", vars["uid"]), nil)
		if err := req.Write(conn); err != nil {
			http.Error(res, err, http.StatusBadRequest)
			return
		}

	})
	go http.ListenAndServe(":8080", router)

	server := NewServer(opts, tunnel)

	logrus.Fatal(server.ListenAndServe())
}
