package websocket

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/server"
	"github.com/shellhub-io/shellhub/ssh/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

func HandlerWebsocket(ws *websocket.Conn) {
	user := GetFromQuery(ws, "user")
	passwd := GetFromQuery(ws, "passwd")
	fingerprint := GetFromQuery(ws, "fingerprint")
	signature := GetFromQuery(ws, "signature")
	cols, _ := strconv.Atoi(GetFromQuery(ws, "cols"))
	rows, _ := strconv.Atoi(GetFromQuery(ws, "rows"))

	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	if fingerprint != "" && signature != "" { //nolint:nestif
		parts := strings.SplitN(user, "@", 2)
		if len(parts) != 2 {
			util.Closes(ws)

			return
		}

		apiClient := internalclient.NewClient()

		device, err := apiClient.GetDevice(parts[1])
		if err != nil {
			return
		}

		key, err := apiClient.GetPublicKey(fingerprint, device.TenantID)
		if err != nil {
			util.WriteAndClose(ws, "permission denied")

			return
		}

		if ok, err := apiClient.EvaluateKey(fingerprint, device, parts[0]); !ok || err != nil {
			util.WriteAndClose(ws, "permission denied")

			return
		}

		pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data)
		if err != nil {
			return
		}

		digest, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			util.Closes(ws)

			return
		}

		err = pubKey.Verify([]byte(parts[0]), &ssh.Signature{
			Format: pubKey.Type(),
			Blob:   digest,
		})
		if err != nil {
			util.Closes(ws)

			return
		}

		signer, err := ssh.NewSignerFromKey(server.RSAKey)
		if err != nil {
			util.Closes(ws)

			return
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	}

	dial, err := ssh.Dial("tcp", "localhost:2222", config)
	if err != nil {
		util.Closes(ws)

		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	closeSession := func(s *ssh.Session) {
		err := s.Close()
		if err != nil {
			logrus.WithError(err).Error("could not close the SSH session")
		}
	}

	session, err := dial.NewSession()
	if err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	if err = session.Setenv("IP_ADDRESS", GetFromHeader(ws, "X-Real-Ip")); err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	sshOut, err := session.StdoutPipe()
	if err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	sshIn, err := session.StdinPipe()
	if err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	if err := session.Shell(); err != nil {
		closeSession(session)
		util.Closes(ws)

		return
	}

	doneCh := make(chan bool)

	go copyWorker(sshIn, ws, doneCh)

	go func() {
		err := redirToWs(sshOut, ws)
		if err != nil {
			logrus.WithError(err).Error("could not redirect to websocket")
		}

		doneCh <- true
	}()

	conn := &wsconn{
		pinger: time.NewTicker(pingInterval),
	}

	defer conn.pinger.Stop()

	go conn.keepAlive(ws)

	<-doneCh

	err = dial.Close()
	if err != nil {
		logrus.WithError(err).Error("could not closes the ssh dial internalclient")
	}

	util.Closes(ws)

	<-doneCh
}
