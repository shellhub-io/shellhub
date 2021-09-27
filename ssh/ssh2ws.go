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
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	io.Copy(dst, src) // nolint:errcheck
	doneCh <- true
}

func HandlerWebsocket(ws *websocket.Conn) {
	user := ws.Request().URL.Query().Get("user")
	passwd := ws.Request().URL.Query().Get("passwd")
	fingerprint := ws.Request().URL.Query().Get("fingerprint")
	signature := ws.Request().URL.Query().Get("signature")
	cols, _ := strconv.Atoi(ws.Request().URL.Query().Get("cols"))
	rows, _ := strconv.Atoi(ws.Request().URL.Query().Get("rows"))

	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	if fingerprint != "" && signature != "" { //nolint:nestif
		parts := strings.SplitN(user, "@", 2)
		if len(parts) != 2 {
			ws.Close()

			return
		}

		apiClient := client.NewClient()

		device, err := apiClient.GetDevice(parts[1])
		if err != nil {
			return
		}

		key, err := apiClient.GetPublicKey(fingerprint, device.TenantID)
		if err != nil {
			fmt.Println(err)                          //nolint:forbidigo
			ws.Write([]byte("Permission denied\r\n")) // nolint:errcheck
			ws.Close()

			return
		}

		if ok, err := apiClient.EvaluateKey(fingerprint, device, parts[0]); !ok || err != nil {
			if err != nil {
				fmt.Println(err) //nolint:forbidigo
			}
			ws.Write([]byte("Permission denied\r\n")) // nolint:errcheck
			ws.Close()

			return
		}

		pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data)
		if err != nil {
			return
		}

		digest, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			ws.Close()

			return
		}

		err = pubKey.Verify([]byte(parts[0]), &ssh.Signature{
			Format: pubKey.Type(),
			Blob:   digest,
		})
		if err != nil {
			fmt.Println(err) //nolint:forbidigo
			ws.Close()

			return
		}

		signer, err := ssh.NewSignerFromKey(magicKey)
		if err != nil {
			ws.Close()

			return
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	}

	client, err := ssh.Dial("tcp", "localhost:2222", config)
	if err != nil {
		fmt.Println(err) //nolint:forbidigo
		ws.Close()

		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	session, err := client.NewSession()
	if err != nil {
		session.Close()
		ws.Close()

		return
	}

	if err = session.Setenv("IP_ADDRESS", ws.Request().Header.Get("X-Real-Ip")); err != nil {
		session.Close()
		ws.Close()

		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		ws.Close()

		return
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		session.Close()
		ws.Close()

		return
	}

	if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
		session.Close()
		ws.Close()

		return
	}

	if err := session.Shell(); err != nil {
		session.Close()
		ws.Close()

		return
	}

	doneCh := make(chan bool)

	go copyWorker(sshIn, ws, doneCh)

	go func() {
		redirToWs(sshOut, ws) // nolint:errcheck
		doneCh <- true
	}()

	conn := &wsconn{
		pinger: time.NewTicker(pingInterval),
	}

	defer conn.pinger.Stop()

	go conn.keepAlive(ws)

	<-doneCh

	client.Close()
	ws.Close()

	<-doneCh
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
