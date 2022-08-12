package session

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/ssh/pkg/flow"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

var (
	ErrFindPublicKey      = fmt.Errorf("failed to get the public key from the server")
	ErrEvaluatePublicKey  = fmt.Errorf("failed to evaluate the public key in the server")
	ErrForbiddenPublicKey = fmt.Errorf("failed to use the public key for this action")
	ErrDataPublicKey      = fmt.Errorf("failed to parse the public key data")
	ErrSignaturePublicKey = fmt.Errorf("failed to decode the public key signature")
	ErrVerifyPublicKey    = fmt.Errorf("failed to verify the public key")
	ErrFindDevice         = fmt.Errorf("failed to find the device")
	ErrSignerPublicKey    = fmt.Errorf("failed to signer the public key")
	ErrDialSSH            = fmt.Errorf("failed to dial to connect to SSH server")
	ErrSession            = fmt.Errorf("failed to create the SSH session")
	ErrEnvIPAddress       = fmt.Errorf("failed to set the env virable of ip address to session")
	ErrEnvWS              = fmt.Errorf("failed to set the env virable of web socket to session")
	ErrPipe               = fmt.Errorf("failed to pipe session data from client to agent")
	ErrPty                = fmt.Errorf("failed to request the pty from agent")
	ErrShell              = fmt.Errorf("failed to get the shell from agent")
)

// WebData contains the data required by web termianl connection.
type WebData struct {
	// User is the device's user.
	User string
	// Password is the user's device password.
	// when Password is set, Fingerprint must not be set.
	Password string
	// Fingerprint is the public key fingerprint.
	// when Fingerprint is set, Password must not be set.
	Fingerprint string
	Signature   string
	// Columns is the width size of pty.
	Columns int
	// Rows is the height size of pty.
	Rows int
}

// NewWebData create a new WebData.
// WebData contains the data required by web termianl connection.
func NewWebData(socket *websocket.Conn) *WebData {
	get := func(socket *websocket.Conn, key string) string {
		return socket.Request().URL.Query().Get(key)
	}

	toInt := func(text string) int {
		integer, err := strconv.Atoi(text)
		if err != nil {
			log.WithError(err).Error("failed to convert the text to int")
		}

		return integer
	}

	return &WebData{
		User:        get(socket, "user"),
		Password:    get(socket, "passwd"),
		Fingerprint: get(socket, "fingerprint"),
		Signature:   get(socket, "signature"),
		Columns:     toInt(get(socket, "cols")),
		Rows:        toInt(get(socket, "rows")),
	}
}

// isPublicKey checks if connection is using public key method.
func (c *WebData) isPublicKey() bool { // nolint: unused
	return c.Fingerprint != "" && c.Signature != ""
}

// isPassword checks if connection is using password method.
func (c *WebData) isPassword() bool {
	return c.Password != ""
}

// GetAuth gets the authentication methods from connection.
func (c *WebData) GetAuth(magicKey *rsa.PrivateKey) ([]ssh.AuthMethod, error) {
	if c.isPassword() {
		return []ssh.AuthMethod{ssh.Password(c.Password)}, nil
	}

	tag, err := target.NewTarget(c.User)
	if err != nil {
		return nil, ErrTarget
	}

	cli := internalclient.NewClient()

	// Trys to get a device from the API.
	device, err := cli.GetDevice(tag.Data)
	if err != nil {
		return nil, ErrFindDevice
	}

	// Trys to get a public key from the API.
	key, err := cli.GetPublicKey(c.Fingerprint, device.TenantID)
	if err != nil {
		return nil, ErrFindPublicKey
	}

	// Trys to evaluate the public key from the API.
	ok, err := cli.EvaluateKey(c.Fingerprint, device, tag.Username)
	if err != nil {
		return nil, ErrEvaluatePublicKey
	}

	if !ok {
		return nil, ErrForbiddenPublicKey
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data) //nolint: dogsled
	if err != nil {
		return nil, ErrDataPublicKey
	}

	digest, err := base64.StdEncoding.DecodeString(c.Signature)
	if err != nil {
		return nil, ErrSignaturePublicKey
	}

	if err := pubKey.Verify([]byte(tag.Username), &ssh.Signature{ //nolint: exhaustruct
		Format: pubKey.Type(),
		Blob:   digest,
	}); err != nil {
		return nil, ErrVerifyPublicKey
	}

	signer, err := ssh.NewSignerFromKey(magicKey)
	if err != nil {
		return nil, ErrSignerPublicKey
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

// WebSession is the session's handler for connection coming from the web terminal.
func WebSession(socket *websocket.Conn) {
	log.Info("handling web session request started")

	// exit responds and finish SSH's session when something goes wrong during session handling.
	// Internal error is the error from Go code, and external one is what going to be send to end user.
	exit := func(session *ssh.Session, socket *websocket.Conn, internal, external error) {
		log.WithFields(log.Fields{
			"internal": internal,
			"external": external,
		}).Error("failed to handler the web session")

		// finish close the SSH's session and websocket connection.
		finish := func(session *ssh.Session, socket *websocket.Conn) {
			if session != nil {
				err := session.Close()
				if err != nil {
					log.WithError(err).Error("failed to finish the web session")
				}
			}

			if socket != nil {
				err := socket.Close()
				if err != nil {
					log.WithError(err).Error("failed to close the web socket")
				}
			}
		}

		respond := func(socket *websocket.Conn, err error) {
			_, err = socket.Write([]byte(err.Error()))
			if err != nil {
				log.WithError(err).Error("failed to write the error to the socket")
			}
		}

		respond(socket, external)
		finish(session, socket)
	}

	data := NewWebData(socket)

	auth, err := data.GetAuth(magickey.GetRerefence())
	if err != nil {
		exit(nil, socket, nil, err)

		return
	}

	server, err := ssh.Dial("tcp", "localhost:2222", &ssh.ClientConfig{ //nolint: exhaustruct
		User:            data.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	})
	if err != nil {
		exit(nil, socket, err, ErrDialSSH)

		return
	}

	session, err := server.NewSession()
	if err != nil {
		exit(session, socket, err, ErrSession)

		return
	}

	if err = session.Setenv("IP_ADDRESS", socket.Request().Header.Get("X-Real-Ip")); err != nil {
		exit(session, socket, err, ErrEnvIPAddress)

		return
	}

	if err = session.Setenv("WS", "true"); err != nil {
		exit(session, socket, err, ErrEnvWS)

		return
	}

	flow, _ := flow.NewFlow(session)
	if err != nil {
		exit(session, socket, err, ErrPipe)

		return
	}

	if err := session.RequestPty("xterm", data.Rows, data.Columns, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		exit(session, socket, err, ErrPty)

		return
	}

	if err := session.Shell(); err != nil {
		exit(session, socket, err, ErrShell)

		return
	}

	done := make(chan bool)

	go func() {
		io.Copy(flow.Stdin, socket) // nolint:errcheck

		done <- true
	}()

	go func() {
		redirToWs(flow.Stdout, socket) // nolint:errcheck

		done <- true
	}()

	conn := &wsconn{
		pinger: time.NewTicker(pingInterval),
	}

	defer conn.pinger.Stop()

	go conn.keepAlive(socket)

	<-done

	server.Close()
	socket.Close()

	<-done

	log.Info("handling web session request closed")
}

func redirToWs(rd io.Reader, ws *websocket.Conn) error {
	var buf [32 * 1024]byte
	var start, end, buflen int

	for {
		nr, err := rd.Read(buf[start:])
		if err != nil {
			return err
		}

		buflen = start + nr
		for end = buflen - 1; end >= 0; end-- {
			if utf8.RuneStart(buf[end]) {
				ch, width := utf8.DecodeRune(buf[end:buflen])
				if ch != utf8.RuneError {
					end += width
				}

				break
			}

			if buflen-end >= 6 {
				end = nr

				break
			}
		}

		if _, err = ws.Write([]byte(string(bytes.Runes(buf[0:end])))); err != nil {
			return err
		}

		start = buflen - end

		if start > 0 {
			// copy remaning read bytes from the end to the beginning of a buffer
			// so that we will get normal bytes
			for i := 0; i < start; i++ {
				buf[i] = buf[end+i]
			}
		}
	}
}

const pingInterval = time.Second * 30

type wsconn struct {
	pinger *time.Ticker
}

func (w *wsconn) keepAlive(ws *websocket.Conn) {
	for {
		if err := ws.SetDeadline(clock.Now().Add(pingInterval * 2)); err != nil {
			return
		}

		if fw, err := ws.NewFrameWriter(websocket.PingFrame); err != nil {
			return
		} else if _, err = fw.Write([]byte{}); err != nil {
			return
		}

		if _, running := <-w.pinger.C; !running {
			return
		}
	}
}
