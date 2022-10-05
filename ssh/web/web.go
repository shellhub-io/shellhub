package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/shellhub-io/shellhub/ssh/web/pkg/session"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Request struct {
	Device      string `json:"device"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Fingerprint string `json:"fingerprint"`
	Signature   string `json:"signature"`
}

type Response struct {
	Token string `json:"token"`
}

// NewSession create session's token to web terminal connection.
//
// It receives on request's body, the device's UID, the device's username and the device's password, to returns a
// JWT token that can be used to connect to the device. The JWT token is generated using a UUID as payload, and
// encrypted using a RSA private key.
//
// If a error occurs, it returns the error message and the HTTP status code related to the error.
func NewSession(res http.ResponseWriter, req *http.Request) {
	fail := func(response http.ResponseWriter, device, username string, status int, err error) {
		log.WithError(err).WithFields(log.Fields{
			"device":   device,
			"username": username,
			"status":   status,
		}).Error("failed to get the session's token")

		http.Error(response, err.Error(), status)
	}
	success := func(req http.ResponseWriter, device, username, token string) {
		log.WithFields(log.Fields{
			"device":   device,
			"username": username,
			"token":    token,
		}).Info("session's token generated successfully")

		req.WriteHeader(http.StatusOK)
		req.Header().Set("Content-Type", "application/json")

		json.NewEncoder(req).Encode(Response{Token: token}) // nolint: errcheck
	}

	ctx := req.Context()

	var data Request
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		fail(res, "", "", http.StatusBadRequest, errors.New("failed to decode the request body"))

		return
	}

	session, err := session.NewSession(ctx, data.Device, data.Username, data.Password, data.Fingerprint, data.Signature)
	if err != nil {
		fail(res, data.Device, data.Username, http.StatusInternalServerError, errors.New("failed to generate the session's token"))
	}

	success(res, session.Device, session.Username, session.Token)
}

// RestoreSession restore session's token to web terminal connection.
//
// It receives the session's token as a websocket's query parameter and verifies if the token is valid. If the token is
// valid, it calls the websocket handler to connect to the device. If the token is invalid, it returns an error message.
func RestoreSession(next func(socket *websocket.Conn, device, username, password, fingerprint, signature string)) websocket.Handler {
	return func(socket *websocket.Conn) {
		ctx := socket.Request().Context()

		fail := func(socket *websocket.Conn, internal, external error) {
			log.Error(internal.Error())

			socket.Write([]byte(fmt.Sprintf("%s\n", external.Error()))) // nolint: errcheck
		}

		get := func(socket *websocket.Conn, key string) (string, bool) {
			value := socket.Request().URL.Query().Get(key)

			return value, value != ""
		}

		token, ok := get(socket, "token")
		if !ok {
			fail(socket, errors.New("failed to get the session's token"), errors.New("failed to get the session's token"))

			return
		}

		session, err := session.Restore(ctx, token)
		if err != nil {
			fail(socket, errors.New("failed to restore the session"), errors.New("failed to restore the session"))

			return
		}

		next(socket, session.Device, session.Username, session.Password, session.Fingerprint, session.Signature)
	}
}
