package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/egress"
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
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(status)

			encoder.Encode(data) //nolint: errcheck,errchkjson
		}

		var request Credentials
		if err := decoder.Decode(&request); err != nil {
			// Keep the raw decode error out of the response; it can carry internal
			// detail and the client can't act on it anyway.
			log.WithError(err).Debug("failed to decode direct connect credentials")
			response(res, http.StatusBadRequest, Fail{Error: "invalid request body"})

			return
		}

		key := magickey.GetReference()

		tkn, err := token.NewToken(key)
		if err != nil {
			log.WithError(err).Error("failed to create direct connect token")
			response(res, http.StatusInternalServerError, Fail{Error: "failed to create token"})

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

		if err := connectSession(wsconn.Request().Context(), conn, creds, Dimensions{cols, rows}); err != nil {
			exit(wsconn, err)

			return
		}
	})))
}

// connectSession dials the external SSH endpoint described by creds and pipes a
// shell to the websocket connection. The context is tied to the websocket
// request, so a dropped client cancels an in-flight dial.
func connectSession(ctx context.Context, conn *Conn, creds *Credentials, dim Dimensions) error {
	logger := log.WithFields(log.Fields{
		"user": creds.Username,
		"host": creds.Host,
		"port": creds.Port,
	})

	logger.Info("handling direct connect request started")
	defer logger.Info("handling direct connect request end")

	addr := net.JoinHostPort(creds.Host, strconv.Itoa(creds.Port))

	// Public-key auth keeps the private key in the browser: the server advertises
	// the supplied public key and proxies each signing challenge over the
	// websocket via [Signer]. Falls back to password when no key is selected.
	var auth []gossh.AuthMethod
	if creds.isPublicKey() {
		pubKey, _, _, _, parseErr := gossh.ParseAuthorizedKey([]byte(creds.PublicKey))
		if parseErr != nil {
			logger.WithError(parseErr).Debug("failed to parse the direct connection public key")

			return ErrDataPublicKey
		}

		auth = []gossh.AuthMethod{gossh.PublicKeys(&Signer{conn: conn, publicKey: &pubKey})}
	} else {
		auth = []gossh.AuthMethod{gossh.Password(creds.Password)}
	}

	// creds come straight from the browser; reject an out-of-range port so the
	// uint16 conversion below is well-defined.
	if creds.Port < 1 || creds.Port > math.MaxUint16 {
		logger.WithField("port", creds.Port).Debug("rejected out-of-range port")

		return ErrAuthentication
	}

	// The host comes straight from the browser, so dial through an SSRF guardian:
	// guardian.Safe validates the real resolved IP (and the port) at the socket
	// layer right before connecting, so the server can't be used as a pivot to
	// reach internal/reserved addresses. Only the configured target port passes.
	dialer := egress.GuardedDialer(creds.Port)
	// Live session, so a longer timeout than the reachability probe's default.
	dialer.Timeout = 30 * time.Second

	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		// Distinguish a guardian rejection (host/port not a permitted target)
		// from an ordinary unreachable host.
		if egress.IsBlocked(err) {
			logger.WithError(err).Warn("blocked direct connect to a disallowed host")

			return ErrEgressBlocked
		}

		logger.WithError(err).Debug("failed to dial the direct host")

		return ErrUnreachableHost
	}

	// Verify the live host key against the one the browser confirmed (TOFU). A
	// mismatch means the host differs from what the user trusted, so abort. An
	// empty key would disable verification entirely (the value is browser-supplied,
	// so an attacker who can shape the request could force it), so refuse it.
	if creds.KnownHostKey == "" {
		netConn.Close() //nolint:errcheck
		logger.Warn("direct connect rejected: no verified host key")

		return ErrHostKeyUnverified
	}

	expected, _, _, _, parseErr := gossh.ParseAuthorizedKey([]byte(creds.KnownHostKey))
	if parseErr != nil {
		netConn.Close() //nolint:errcheck
		logger.WithError(parseErr).Warn("invalid known host key supplied")

		return ErrAuthentication
	}

	// gossh wraps the callback error, so errors.Is won't reliably recover the
	// sentinel after the handshake. Record the mismatch out of band instead.
	hostKeyMismatch := false
	expectedKey := expected.Marshal()
	hostKeyCallback := func(_ string, _ net.Addr, key gossh.PublicKey) error {
		if !bytes.Equal(key.Marshal(), expectedKey) {
			hostKeyMismatch = true

			return ErrHostKeyMismatch
		}

		return nil
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(netConn, addr, &gossh.ClientConfig{ //nolint: exhaustruct
		User:            creds.Username,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	})
	if err != nil {
		netConn.Close() //nolint:errcheck

		if hostKeyMismatch {
			logger.Warn("host key mismatch on direct connect")

			return ErrHostKeyMismatch
		}

		return ErrAuthentication
	}

	client := gossh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	agent, err := client.NewSession()
	if err != nil {
		return ErrSession
	}

	defer agent.Close()

	// Return a sentinel so the raw error (internal detail) isn't echoed to the browser.
	stdin, err := agent.StdinPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stdin pipe")

		return ErrSession
	}

	stdout, err := agent.StdoutPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stdout pipe")

		return ErrSession
	}

	stderr, err := agent.StderrPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stderr pipe")

		return ErrSession
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
