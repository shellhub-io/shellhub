package tunnel

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
	HTTPHandler  func(w http.ResponseWriter, r *http.Request)
	ConnHandler  func(w http.ResponseWriter, r *http.Request)
	CloseHandler func(w http.ResponseWriter, r *http.Request)
}

func NewTunnel() *Tunnel {
	router := mux.NewRouter()

	t := &Tunnel{
		router: router,
		srv: &http.Server{
			Handler: router,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c) //nolint:revive
			},
		},
		HTTPHandler: func(w http.ResponseWriter, r *http.Request) {
			panic("HTTPHandler can not be nil")
		},
		ConnHandler: func(w http.ResponseWriter, r *http.Request) {
			panic("connHandler can not be nil")
		},
		CloseHandler: func(w http.ResponseWriter, r *http.Request) {
			panic("closeHandler can not be nil")
		},
	}
	t.router.HandleFunc("/ssh/http", func(w http.ResponseWriter, r *http.Request) {
		t.HTTPHandler(w, r)
	})
	t.router.HandleFunc("/ssh/{id}", func(w http.ResponseWriter, r *http.Request) {
		t.ConnHandler(w, r)
	})
	t.router.HandleFunc("/ssh/close/{id}", func(w http.ResponseWriter, r *http.Request) {
		t.CloseHandler(w, r)
	})

	return t
}

// Listen to reverse listener.
func (t *Tunnel) Listen(l *revdial.Listener) error {
	return t.srv.Serve(l)
}
