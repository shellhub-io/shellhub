package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type (
	functionHandleCreateSession  func(ctx context.Context, data *Input) (*Session, error)
	functionHandleRestoreSession func(ctx context.Context, data *Output) (*Session, error)
)

// HandlerCreateSession handles a HTTP request with the data to create a new web session.
//
// It receives on request's body the device's UID and the device's username, either the device's password or the
// device's fingerprint and the device's signature, to returns a JWT token that can be used to connect to the device.
// The JWT token is generated using a UUID as payload, and encrypted using a runtime generated RSA private key.
//
// If a error occurs, it logs on the server the error and returns the error message and the HTTP status code related to
// the error to the user.
func HandlerCreateSession(create functionHandleCreateSession) func(http.ResponseWriter, *http.Request) {
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

	success := func(req http.ResponseWriter, device, username, token string) {
		log.WithFields(log.Fields{
			"device":   device,
			"username": username,
			"token":    token,
		}).Info("session's token generated successfully")

		req.WriteHeader(http.StatusOK)
		req.Header().Set("Content-Type", "application/json")

		json.NewEncoder(req).Encode(Response{Token: token}) //nolint: errcheck,errchkjson
	}

	fail := func(response http.ResponseWriter, device, username string, status int, err error) {
		log.WithError(err).WithFields(log.Fields{
			"device":   device,
			"username": username,
			"status":   status,
		}).Error("failed to get the session's token")

		http.Error(response, err.Error(), status)
	}

	return func(res http.ResponseWriter, req *http.Request) {
		var request *Request
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			fail(res, "", "", http.StatusBadRequest, errors.New("failed to decode the request body"))
		}

		data := &Input{
			Device:      request.Device,
			Username:    request.Username,
			Password:    request.Password,
			Fingerprint: request.Fingerprint,
			Signature:   request.Signature,
		}

		session, err := create(req.Context(), data)
		if err != nil {
			fail(res, data.Device, data.Username, http.StatusInternalServerError, errors.New("failed to generate the session's token"))
		}

		success(res, session.Device, session.Username, session.Token)
	}
}

// HandlerCreateSession handles a websocket request with the data to restore web session.
//
// It receives the session's token as a websocket's query parameter and verifies if the token is valid. If the token is
// valid, it calls the websocket handler to connect to the device. If the token is invalid, it returns an error message.
//
// If any other error occurs, it logs on the server the error and returns the error message and the error to the user.
func HandlerRestoreSession(restore functionHandleRestoreSession, handler func(socket *websocket.Conn, session *Session)) websocket.Handler {
	return func(socket *websocket.Conn) {
		get := func(socket *websocket.Conn, key string) (string, bool) {
			value := socket.Request().URL.Query().Get(key)

			return value, value != ""
		}

		fail := func(socket *websocket.Conn, internal, external error) {
			log.Error(internal.Error())

			socket.Write([]byte(fmt.Sprintf("%s\n", external.Error()))) //nolint: errcheck
		}

		token, ok := get(socket, "token")
		if !ok {
			fail(socket, errors.New("failed to get the token from the websocket"), errors.New("failed to get the token from the websocket"))
		}

		session, err := restore(socket.Request().Context(), &Output{Token: token})
		if err != nil {
			fail(socket, err, errors.New("failed to get the session"))
		}

		handler(socket, session)
	}
}
