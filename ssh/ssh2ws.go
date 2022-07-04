package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/ssh/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

type Connection struct {
	User        string
	Password    string
	Fingerprint string
	Signature   string
	Columns     int
	Rows        int
}

func NewConnection(socket *websocket.Conn) *Connection {
	get := func(socket *websocket.Conn, key string) string {
		return socket.Request().URL.Query().Get(key)
	}

	toInt := func(text string) int {
		integer, err := strconv.Atoi(text)
		if err != nil {
			log.WithError(err).Errorln("could not convert the text to int")
		}

		return integer
	}

	return &Connection{
		User:        get(socket, "user"),
		Password:    get(socket, "passwd"),
		Fingerprint: get(socket, "fingerprint"),
		Signature:   get(socket, "signature"),
		Columns:     toInt(get(socket, "cols")),
		Rows:        toInt(get(socket, "rows")),
	}
}

// isPublicKey checks if connection is using public key method.
func (c *Connection) isPublicKey() bool { // nolint: unused
	return c.Fingerprint != "" && c.Signature != ""
}

// isPassword checks if connection is using password method.
func (c *Connection) isPassword() bool {
	return c.Password != ""
}

// getAuth gets the authentication methods from connection.
func (c *Connection) getAuth() ([]ssh.AuthMethod, error) {
	if c.isPassword() {
		return []ssh.AuthMethod{ssh.Password(c.Password)}, nil
	}

	tag, err := NewTarget(c.User)
	if err != nil {
		return nil, ErrInvalidSessionTarget
	}

	cli := client.NewClient()

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

	if err := pubKey.Verify([]byte(tag.Username), &ssh.Signature{
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

func HandlerWebsocket(socket *websocket.Conn) {
	exit := func(session *ssh.Session, socket *websocket.Conn, internal, external error) {
		finish := func(session *ssh.Session, socket *websocket.Conn) {
			if session != nil {
				session.Close()
			}

			if socket != nil {
				socket.Close()
			}
		}

		respond := func(socket *websocket.Conn, err error) {
			_, err = socket.Write([]byte(errors.GetExternal(err).Error()))
			if err != nil {
				log.WithError(err).Errorln("could not write the error to the socket")
			}
		}

		log.WithError(internal).Errorln(external)

		respond(socket, external)
		finish(session, socket)
	}

	connection := NewConnection(socket)

	auth, err := connection.getAuth()
	if err != nil {
		exit(nil, socket, nil, err)

		return
	}

	cli, err := ssh.Dial("tcp", "localhost:2222", &ssh.ClientConfig{
		User:            connection.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	})
	if err != nil {
		exit(nil, socket, err, ErrDialSSH)

		return
	}

	session, err := cli.NewSession()
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

	sshIn, err := session.StdinPipe()
	if err != nil {
		exit(session, socket, err, ErrPipeStdin)

		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		exit(session, socket, err, ErrPipeStdout)

		return
	}

	if err := session.RequestPty("xterm.js", connection.Rows, connection.Columns, ssh.TerminalModes{
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
		io.Copy(sshIn, socket) // nolint:errcheck
		done <- true
	}()

	go func() {
		redirToWs(sshOut, socket) // nolint:errcheck
		done <- true
	}()

	conn := &wsconn{
		pinger: time.NewTicker(pingInterval),
	}

	defer conn.pinger.Stop()

	go conn.keepAlive(socket)

	<-done

	cli.Close()
	socket.Close()

	<-done
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
