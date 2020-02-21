package main

import (
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

	go http.ListenAndServe(":8080", tunnel.Router())

	server := NewServer(opts, tunnel)

	logrus.Fatal(server.ListenAndServe())
}
