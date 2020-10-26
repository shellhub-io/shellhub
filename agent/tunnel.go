package main

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Tunnel struct {
	router       *mux.Router
	srv          *http.Server
	connHandler  func(w http.ResponseWriter, r *http.Request)
	closeHandler func(w http.ResponseWriter, r *http.Request)
}

func NewTunnel() *Tunnel {
	router := mux.NewRouter()

	t := &Tunnel{
		router: router,
		srv: &http.Server{
			Handler: router,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c)
			},
		},
		connHandler: func(w http.ResponseWriter, r *http.Request) {
			panic("connHandler can not be nil")
		},
		closeHandler: func(w http.ResponseWriter, r *http.Request) {
			panic("closeHandler can not be nil")
		},
	}
	t.router.HandleFunc("/ssh/{id}", func(w http.ResponseWriter, r *http.Request) {
		t.connHandler(w, r)
	})
	t.router.HandleFunc("/ssh/close/{id}", func(w http.ResponseWriter, r *http.Request) {
		t.closeHandler(w, r)
	})

	return t
}

// Listen to reverse listener
func (t *Tunnel) Listen(l *revdial.Listener) error {
	return t.srv.Serve(l)
}
