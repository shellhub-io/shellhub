package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// NewSSHServerBridge creates routes into a [echo.Router] to connect a webscoket to SSH using Shell session.
func NewSSHServerBridge(router *echo.Echo, cache cache.Cache) {
	const WebsocketSSHBridgeRoute = "/ws/ssh"

	manager := newManager(30 * time.Second)

	// NOTICE: this is the route that users send your credentials securely.
	router.Add(http.MethodPost, WebsocketSSHBridgeRoute, echo.WrapHandler(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			type Success struct {
				Token string `json:"token"`
			}

			type Fail struct {
				Error string `json:"error"`
			}

			decoder := json.NewDecoder(req.Body)
			encoder := json.NewEncoder(res)

			response := func(res http.ResponseWriter, status int, data any) {
				res.WriteHeader(status)
				res.Header().Set("Content-Type", "application/json")

				encoder.Encode(data) //nolint: errcheck,errchkjson
			}

			var request Credentials
			if err := decoder.Decode(&request); err != nil {
				response(res, http.StatusBadRequest, Fail{Error: err.Error()})

				return
			}

			key := magickey.GetReference()

			token, err := token.NewToken(key)
			if err != nil {
				response(res, http.StatusBadRequest, Fail{Error: err.Error()})

				return
			}

			request.encryptPassword(key) //nolint:errcheck

			// NOTICE: saved credentials are delete after a time period.
			manager.save(token.ID, &request)

			response(res, http.StatusOK, Success{Token: token.ID})
		})),
	)

	router.Add(http.MethodGet, WebsocketSSHBridgeRoute, echo.WrapHandler(websocket.Handler(func(wsconn *websocket.Conn) {
		defer wsconn.Close()

		// exit sends the error's message to the client on the browser.
		exit := func(wsconn *websocket.Conn, err error) {
			log.WithError(err).Error("web terminal error")

			buffer, marshalErr := json.Marshal(Message{
				Kind: messageKindError,
				Data: err.Error(),
			})
			if marshalErr != nil {
				log.WithError(marshalErr).Error("failed to marshal error message")

				return
			}

			wsconn.Write(buffer) //nolint:errcheck
		}

		token, err := getToken(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetToken)

			return
		}

		cols, rows, err := getDimensions(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetDimensions)

			return
		}

		ip, err := getIP(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetIP)

			return
		}

		creds, ok := manager.get(token)
		if !ok {
			exit(wsconn, ErrBridgeCredentialsNotFound)
		}

		conn := NewConn(wsconn)
		defer conn.Close()

		go conn.KeepAlive()

		creds.decryptPassword(magickey.GetReference()) //nolint:errcheck

		if err := newSession(
			wsconn.Request().Context(),
			cache,
			conn,
			creds,
			Dimensions{cols, rows},
			Info{IP: ip},
		); err != nil {
			exit(wsconn, err)

			return
		}
	})))
}
