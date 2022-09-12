package handler

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
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

// WebData contains the data required by web terminal connection.
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

// WebSession is the Client's handler for connection coming from the web terminal.
func WebSession(socket *websocket.Conn) {
	log.Info("handling web client request started")
	defer log.Info("handling web client request end")

	data := NewWebData(socket)

	auth, err := data.GetAuth(magickey.GetRerefence())
	if err != nil {
		sendAndInformError(socket, err, ErrGetAuth)

		return
	}

	connection, err := ssh.Dial("tcp", "localhost:2222", &ssh.ClientConfig{ //nolint: exhaustruct
		User:            data.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	})
	if err != nil {
		sendAndInformError(socket, err, ErrDialSSH)

		return
	}

	defer connection.Close()

	agent, err := connection.NewSession()
	if err != nil {
		sendAndInformError(socket, err, ErrSession)

		return
	}

	defer agent.Close()

	if err = agent.Setenv("IP_ADDRESS", socket.Request().Header.Get("X-Real-Ip")); err != nil {
		sendAndInformError(socket, err, ErrEnvIPAddress)

		return
	}

	if err = agent.Setenv("WS", "true"); err != nil {
		sendAndInformError(socket, err, ErrEnvWS)

		return
	}

	flw, err := flow.NewFlow(agent)
	if err != nil {
		sendAndInformError(socket, err, ErrPipe)

		return
	}

	defer flw.Close()

	if err := agent.RequestPty("xterm", data.Rows, data.Columns, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		sendAndInformError(socket, err, ErrPty)

		return
	}

	if err := agent.Shell(); err != nil {
		sendAndInformError(socket, err, ErrShell)

		return
	}

	done := make(chan bool)

	go flw.PipeIn(socket, done)
	go redirToWs(flw.Stdout, socket) // nolint:errcheck
	go flw.PipeErr(socket, nil)

	go func() {
		<-done

		agent.Close()
	}()

	conn := &wsconn{
		pinger: time.NewTicker(pingInterval),
	}

	defer conn.pinger.Stop()

	go conn.keepAlive(socket)

	if err := agent.Wait(); err != nil {
		log.WithError(err).Warning("client remote command returned a error")
	}
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
