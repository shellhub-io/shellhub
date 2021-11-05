package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
)

func HandlerSessionClose(tunnel *httptunnel.Tunnel) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
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
	}
}
