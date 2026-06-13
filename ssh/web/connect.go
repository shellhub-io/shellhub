package web

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

// NewConnectBridge registers the routes for the direct connection bridge: a
// lightweight web terminal that dials an external SSH endpoint directly. Unlike
// [NewSSHServerBridge], it does NOT route through the agent/reverse-tunnel nor
// the device session machinery — it dials host:port itself and pipes the shell
// to the websocket. Used by saved "direct" connections.
//
// MVP: password authentication only. The credential is encrypted in transit and
// kept only for the short TTL of the token cache; it is never persisted.
func NewConnectBridge(router *echo.Echo) {
	const route = "/ws/connect"

	manager := newManager(30 * time.Second)

	// POST receives the connection credentials and returns a short-lived token.
	router.Add(http.MethodPost, route, echo.WrapHandler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
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

		tkn, err := token.NewToken(key)
		if err != nil {
			response(res, http.StatusBadRequest, Fail{Error: err.Error()})

			return
		}

		request.encryptPassword(key) //nolint:errcheck

		manager.save(tkn.ID, &request)

		response(res, http.StatusOK, Success{Token: tkn.ID})
	})))

	// GET upgrades to a websocket and pipes the shell of the dialed host.
	router.Add(http.MethodGet, route, echo.WrapHandler(websocket.Handler(func(wsconn *websocket.Conn) {
		defer wsconn.Close()

		exit := func(wsconn *websocket.Conn, err error) {
			log.WithError(err).Error("web connect terminal error")

			buffer, marshalErr := json.Marshal(Message{Kind: messageKindError, Data: err.Error()})
			if marshalErr != nil {
				return
			}

			wsconn.Write(buffer) //nolint:errcheck
		}

		tkn, err := getToken(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetToken)

			return
		}

		cols, rows, err := getDimensions(wsconn.Request())
		if err != nil {
			exit(wsconn, ErrWebSocketGetDimensions)

			return
		}

		creds, ok := manager.get(tkn)
		if !ok {
			exit(wsconn, ErrBridgeCredentialsNotFound)

			return
		}

		conn := NewConn(wsconn)
		defer conn.Close()

		go conn.KeepAlive()

		creds.decryptPassword(magickey.GetReference()) //nolint:errcheck

		if err := connectSession(conn, creds, Dimensions{cols, rows}); err != nil {
			exit(wsconn, err)

			return
		}
	})))
}

// connectSession dials the external SSH endpoint described by creds and pipes a
// shell to the websocket connection.
func connectSession(conn *Conn, creds *Credentials, dim Dimensions) error {
	logger := log.WithFields(log.Fields{
		"user": creds.Username,
		"host": creds.Host,
		"port": creds.Port,
	})

	logger.Info("handling direct connect request started")
	defer logger.Info("handling direct connect request end")

	addr := net.JoinHostPort(creds.Host, strconv.Itoa(creds.Port))

	client, err := gossh.Dial("tcp", addr, &gossh.ClientConfig{ //nolint: exhaustruct
		User:            creds.Username,
		Auth:            []gossh.AuthMethod{gossh.Password(creds.Password)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         30 * time.Second,
	})
	if err != nil {
		logger.WithError(err).Debug("failed to dial the direct host")

		return ErrAuthentication
	}

	defer client.Close()

	agent, err := client.NewSession()
	if err != nil {
		return ErrSession
	}

	defer agent.Close()

	stdin, err := agent.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := agent.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := agent.StderrPipe()
	if err != nil {
		return err
	}

	if err := agent.RequestPty("xterm", int(dim.Rows), int(dim.Cols), gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return ErrPty
	}

	if err := agent.Shell(); err != nil {
		return ErrShell
	}

	go func() {
		defer agent.Close()

		for {
			var message Message

			if _, err := conn.ReadMessage(&message); err != nil {
				if errors.Is(err, io.EOF) {
					return
				}

				logger.WithError(err).Error("failed to read the message from the client")

				return
			}

			switch message.Kind {
			case messageKindInput:
				buffer, ok := message.Data.(string)
				if !ok {
					continue
				}

				if _, err := stdin.Write([]byte(buffer)); err != nil {
					return
				}
			case messageKindResize:
				d, ok := message.Data.(Dimensions)
				if !ok {
					continue
				}

				if err := agent.WindowChange(int(d.Rows), int(d.Cols)); err != nil {
					return
				}
			}
		}
	}()

	go redirToWs(stdout, conn) //nolint:errcheck
	go io.Copy(conn, stderr)   //nolint:errcheck

	if err := agent.Wait(); err != nil {
		logger.WithError(err).Warning("client remote command returned an error")
	}

	return nil
}
