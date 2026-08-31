package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/shellhub-io/shellhub/server/ssh/web/pkg/token"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

const (
	// WebsocketSSHBridgeRoute is the WebSocket upgrade that carries the terminal
	// stream. It is gated by a single-use token rather than authenticated.
	WebsocketSSHBridgeRoute = "/ws/ssh"
	// WebSessionRoute is the credential/token POST. It is authenticated, so the
	// bridge learns the logged-in ShellHub account (used in identity mode).
	WebSessionRoute = "/ws/ssh/session"
)

// exitLogLevel returns logrus.WarnLevel for expected/banner-derived errors that
// indicate normal client-side rejections (connection refused, access denied,
// invalid SSHID, authentication failure, device not found, bad credentials,
// or a malformed WebSocket handshake parameter) and logrus.ErrorLevel for
// genuine server faults or any unrecognised error.
func exitLogLevel(err error) log.Level {
	switch {
	case errors.Is(err, ErrConnect),
		errors.Is(err, ErrAccessDenied),
		errors.Is(err, ErrInvalidSSHID),
		errors.Is(err, ErrAuthentication),
		errors.Is(err, ErrGetAuth),
		errors.Is(err, ErrFindDevice),
		errors.Is(err, ErrForbiddenPublicKey),
		errors.Is(err, ErrBridgeCredentialsNotFound),
		errors.Is(err, ErrWebSocketGetToken),
		errors.Is(err, ErrWebSocketGetDimensions),
		errors.Is(err, ErrWebSocketGetIP):
		return log.WarnLevel
	default:
		return log.ErrorLevel
	}
}

// NewSSHServerBridge creates routes into a [echo.Router] to connect a webscoket to SSH using Shell session.
//
// authn is the API's authenticator, used to declare the WebSocket upgrade as
// reachable without a credential. It may be nil in tests.
func NewSSHServerBridge(router *echo.Echo, authn *routesmiddleware.Authenticator, service services.Service, handoff *webhandoff.Store) {
	manager := newManager(30 * time.Second)

	if authn != nil {
		authn.AllowAnonymous(http.MethodGet, WebsocketSSHBridgeRoute)
	}

	router.Add(http.MethodPost, WebSessionRoute, echo.WrapHandler(
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
				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(status)

				encoder.Encode(data) //nolint:errcheck
			}

			var request Credentials
			if err := decoder.Decode(&request); err != nil {
				response(res, http.StatusBadRequest, Fail{Error: err.Error()})

				return
			}

			request.UserID = req.Header.Get("X-ID")

			key := magickey.GetReference()

			token, err := token.NewToken(key)
			if err != nil {
				response(res, http.StatusBadRequest, Fail{Error: err.Error()})

				return
			}

			request.encryptPassword(key) //nolint:errcheck

			manager.save(token.ID, &request)

			response(res, http.StatusOK, Success{Token: token.ID})
		})),
	)

	router.Add(http.MethodGet, WebsocketSSHBridgeRoute, echo.WrapHandler(websocket.Handler(func(wsconn *websocket.Conn) {
		defer wsconn.Close() //nolint:errcheck

		exit := func(wsconn *websocket.Conn, err error) {
			log.WithError(err).Log(exitLogLevel(err), "web terminal error")

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

			return
		}

		conn := NewConn(wsconn)
		defer conn.Close() //nolint:errcheck

		go conn.KeepAlive()

		creds.decryptPassword(magickey.GetReference()) //nolint:errcheck

		if err := newSession(
			wsconn.Request().Context(),
			service,
			handoff,
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
