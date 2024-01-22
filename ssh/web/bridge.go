package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
	"golang.org/x/net/websocket"
)

type Bridge struct {
	// TODO: use something more generic as router.
	Router  *echo.Router
	Manager *Manager
}

// NewBridge creates a new web SSH bridge. The bridge is used by a web terminal to connect to device through SSH server.
func NewBridge(router *echo.Router) *Bridge {
	return &Bridge{
		Router:  router,
		Manager: NewManager(30 * time.Second),
	}
}

var (
	ErrWebSocketGetToken      = errors.New("failed to get the token from query")
	ErrWebSocketGetDimensions = errors.New("failed to get terminal dimensions from query")
	ErrWebSocketGetIP         = errors.New("failed to get IP from query")
)

var ErrBridgeCredentialsNotFound = errors.New("failed to find the credentials")

// Handle create endpoints for the web SSH bridge send its credentials securely, init session and handle it.
func (b *Bridge) Handle(
	callback func(
		conn *Conn,
		creds *Credentials,
		token string,
		cols, rows int,
		ip string,
	) error,
) { //nolint:whitespace
	// WebsocketSSHBridgeRoute is the route where the web ssh terminal access to send them credentials and connects to
	// the web socket itself.
	const WebsocketSSHBridgeRoute = "/ws/ssh"

	// NOTICE: this is the route that users send your credentials securely.
	b.Router.Add(http.MethodPost, WebsocketSSHBridgeRoute, echo.WrapHandler(
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

			key := magickey.GetRerefence()

			token, err := token.NewToken(key)
			if err != nil {
				response(res, http.StatusBadRequest, Fail{Error: err.Error()})

				return
			}

			request.EncryptPassword(key) //nolint:errcheck

			// NOTICE: saved credentials are delete after a time period.
			b.SaveCredentials(token.ID, request)

			response(res, http.StatusOK, Success{Token: token.ID})
		})),
	)

	b.Router.Add(http.MethodGet, WebsocketSSHBridgeRoute, echo.WrapHandler(websocket.Handler(func(wsconn *websocket.Conn) {
		defer wsconn.Close()

		// exit sends the error's message to the client on the browser.
		exit := func(wsconn *websocket.Conn, err error) {
			wsconn.Write([]byte(err.Error())) //nolint:errcheck
		}

		token, err := GetToken(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetToken)

			return
		}

		cols, rows, err := GetDimensions(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetDimensions)

			return
		}

		ip, err := GetIP(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetIP)

			return
		}

		creds, err := b.GetCredentials(token)
		if err != nil {
			exit(wsconn, ErrBridgeCredentialsNotFound)

			return
		}

		conn := NewConn(wsconn)
		defer conn.Close()

		go conn.KeepAlive()

		creds.DecryptPassword(magickey.GetRerefence()) //nolint:errcheck

		if err := callback(
			conn,
			creds,
			token,
			cols, rows,
			ip,
		); err != nil {
			exit(wsconn, err)

			return
		}
	})))
}

// SaveCredentials saves the credentials for a time period. After this, the credentials are deleted.
func (b *Bridge) SaveCredentials(id string, creds Credentials) {
	b.Manager.Save(id, &creds)
}

// GetCredentials gets the credentials if it time period have not ended.
func (b *Bridge) GetCredentials(id string) (*Credentials, error) {
	sess, ok := b.Manager.Get(id)
	if !ok {
		return sess, ErrBridgeCredentialsNotFound
	}

	return sess, nil
}
