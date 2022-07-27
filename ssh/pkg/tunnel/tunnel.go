package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	log "github.com/sirupsen/logrus"
)

type Tunnel struct {
	Tunnel *httptunnel.Tunnel
}

func NewTunnel(connection, dial string) *Tunnel {
	return &Tunnel{Tunnel: httptunnel.NewTunnel(connection, dial)}
}

func (t *Tunnel) SetConnectionHandler() {
	t.Tunnel.ConnectionHandler = func(request *http.Request) (string, error) {
		return request.Header.Get(internalclient.DeviceUIDHeader), nil
	}
}

func (t *Tunnel) SetCloseHandler() {
	t.Tunnel.CloseHandler = func(id string) {
		if err := internalclient.NewClient().DevicesOffline(id); err != nil {
			log.Error(err)
		}
	}
}

func (t *Tunnel) SetKeepAliveHandler() {
	t.Tunnel.KeepAliveHandler = func(id string) {
		if err := internalclient.NewClient().DevicesHeartbeat(id); err != nil {
			log.Error(err)
		}
	}
}

func (t *Tunnel) GetRouter() *mux.Router {
	router, ok := t.Tunnel.Router().(*mux.Router)
	if !ok {
		// TODO: should the SSH does not up when this assertion fail?
		log.Error("type assertion failed")
	}

	return router
}

func (t *Tunnel) Dial(ctx context.Context, id string) (net.Conn, error) {
	return t.Tunnel.Dial(ctx, id)
}
