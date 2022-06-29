package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/ssh/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

func createSigner(user, fingerprint, signature string) (ssh.Signer, error) {
	// split splits the user into two parts.
	// The first part is the username and the second part is the device identifier.
	split := func(user string) (string, string, error) {
		const USERNAME = 0
		const IDENTIFIER = 1

		parts := strings.SplitN(user, "@", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("the user does not have two parts")
		}

		return parts[USERNAME], parts[IDENTIFIER], nil
	}

	username, identifier, err := split(user)
	if err != nil {
		return nil, errors.Wrap(fmt.Errorf("not split the user into username and device id"), err)
	}

	// Creates a HTTP client to request the API.
	cli := client.NewClient()

	// Trys to get a device from the API.
	device, err := cli.GetDevice(identifier)
	if err != nil {
		return nil, errors.Wrap(ErrFindDevice, err)
	}

	// Trys to get a public key from the API.
	key, err := cli.GetPublicKey(fingerprint, device.TenantID)
	if err != nil {
		return nil, errors.Wrap(ErrFindPublicKey, err)
	}

	// Trys to evaluate the public key from the API.
	ok, err := cli.EvaluateKey(fingerprint, device, username)
	if err != nil {
		return nil, errors.Wrap(ErrEvaluatePublicKey, err)
	}

	if !ok {
		return nil, errors.Wrap(ErrForbiddenPublicKey, err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data) //nolint: dogsled
	if err != nil {
		return nil, errors.Wrap(ErrDataPublicKey, err)
	}

	digest, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, errors.Wrap(ErrSignaturePublicKey, err)
	}

	if err := pubKey.Verify([]byte(username), &ssh.Signature{
		Format: pubKey.Type(),
		Blob:   digest,
	}); err != nil {
		return nil, errors.Wrap(ErrVerifyPublicKey, err)
	}

	signer, err := ssh.NewSignerFromKey(magicKey)
	if err != nil {
		return nil, errors.Wrap(ErrSignerPublicKey, err)
	}

	return signer, nil
}

func HandlerWebsocket(socket *websocket.Conn) {
	// finish closes the ssh's session and/or websocket's connection.
	finish := func(session *ssh.Session, socket *websocket.Conn) {
		if session != nil {
			session.Close()
		}

		if socket != nil {
			socket.Close()
		}
	}

	// respond sends back to the websocket user the external error message when error is from errors.Error.
	respond := func(socket *websocket.Conn, err error) {
		_, err = socket.Write([]byte(errors.GetExternal(err).Error()))
		if err != nil {
			log.WithError(err).Errorln("could not write the error to the socket")
		}
	}

	// get gets the query variable passed by websocket connection.
	get := func(socket *websocket.Conn, key string) string {
		return socket.Request().URL.Query().Get(key)
	}

	// toInt converts a string to int. If the conversion return a error, toInt return 0 and log the error.
	toInt := func(text string) int {
		integer, err := strconv.Atoi(text)
		if err != nil {
			log.WithError(err).Errorln("could not convert the text to int")
		}

		return integer
	}

	// user is the user in the device system.
	user := get(socket, "user")
	// passwd is the password of the user in the device system.
	// passwd is empty when the user want to connect to device through a prublic key.
	passwd := get(socket, "passwd")
	// fingerprint is the fingerprint of the public key.
	// fingerprint is empty when the user's password is set.
	fingerprint := get(socket, "fingerprint")
	// signature is the private key signature. It should empty if passwd is set.
	// signature is empty when the user's password is set.
	signature := get(socket, "signature")

	// cols and rows are the terminal's dimentation.
	cols := toInt(get(socket, "cols"))
	rows := toInt(get(socket, "rows"))

	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	// checks if fingerprint and signature is set, it is connection from public key.
	if fingerprint != "" && signature != "" {
		signer, err := createSigner(user, fingerprint, signature)
		if err != nil {
			log.WithError(err).Errorln(ErrSigner)

			respond(socket, ErrSigner)
			finish(nil, socket)

			return
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		// if fingerprint and signature is not set, it is connection by password.
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	}

	cli, err := ssh.Dial("tcp", "localhost:2222", config)
	if err != nil {
		log.WithError(err).Errorln(ErrDialSSH)

		respond(socket, ErrDialSSH)
		finish(nil, socket)

		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	session, err := cli.NewSession()
	if err != nil {
		log.WithError(err).Errorln(ErrSession)

		respond(socket, ErrSession)
		finish(session, socket)

		return
	}

	if err = session.Setenv("IP_ADDRESS", socket.Request().Header.Get("X-Real-Ip")); err != nil {
		log.WithError(err).Errorln(ErrEnvIPAddress)

		respond(socket, ErrEnvIPAddress)
		finish(session, socket)

		return
	}

	if err = session.Setenv("WS", "true"); err != nil {
		log.WithError(err).Errorln(ErrEnvWS)

		respond(socket, ErrEnvWS)
		finish(session, socket)

		return
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		log.WithError(err).Errorln(ErrPipeStdin)

		respond(socket, ErrPipeStdin)
		finish(session, socket)

		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		log.WithError(err).Errorln(ErrPipeStdout)

		respond(socket, ErrPipeStdout)
		finish(session, socket)

		return
	}

	if err := session.RequestPty("xterm.js", rows, cols, modes); err != nil {
		log.WithError(err).Errorln(ErrPty)

		respond(socket, ErrPty)
		finish(session, socket)

		return
	}

	if err := session.Shell(); err != nil {
		log.WithError(err).Errorln(ErrShell)

		respond(socket, ErrShell)
		finish(session, socket)

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
