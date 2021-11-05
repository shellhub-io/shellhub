package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/sirupsen/logrus"
)

// CreateTunnel creates and return a http tunnel server between the fromPath to the dialerPath.
func CreateTunnel(fromPath string, dialerPath string) *httptunnel.Tunnel {
	tunnel := httptunnel.NewTunnel(fromPath, dialerPath)
	tunnel.ConnectionHandler = func(r *http.Request) (string, error) {
		return r.Header.Get(client.DeviceUIDHeader), nil
	}
	tunnel.CloseHandler = func(id string) {
		if err := client.NewClient().DevicesOffline(id); err != nil {
			logrus.Error(err)
		}
	}
	tunnel.KeepAliveHandler = func(id string) {
		if err := client.NewClient().DevicesHeartbeat(id); err != nil {
			logrus.Error(err)
		}
	}

	return tunnel
}

// GetTunnelRouter convert and return what return from the tunnel.Router(), http.Handle, to the mux.Router.
func GetTunnelRouter(tunnel *httptunnel.Tunnel) *mux.Router {
	router, ok := tunnel.Router().(*mux.Router)
	if !ok {
		logrus.Error("could convert the tunnel's router to the mux's router")

		return nil
	}

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

	return router
}
