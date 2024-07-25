package web

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type BannerError struct {
	Message string
}

func NewBannerError(message string) *BannerError {
	return &BannerError{
		Message: message,
	}
}

func (b *BannerError) Error() string {
	return b.Message
}

// getAuth gets the authentication methods from credentials.
func getAuth(creds *Credentials, magicKey *rsa.PrivateKey) ([]ssh.AuthMethod, error) {
	if creds.isPassword() {
		return []ssh.AuthMethod{ssh.Password(creds.Password)}, nil
	}

	cli, err := internalclient.NewClient()
	if err != nil {
		return nil, err
	}

	// Trys to get a device from the API.
	device, err := cli.GetDevice(creds.Device)
	if err != nil {
		return nil, ErrFindDevice
	}

	// Trys to get a public key from the API.
	key, err := cli.GetPublicKey(creds.Fingerprint, device.TenantID)
	if err != nil {
		return nil, ErrFindPublicKey
	}

	// Trys to evaluate the public key from the API.
	ok, err := cli.EvaluateKey(creds.Fingerprint, device, creds.Username)
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

	digest, err := base64.StdEncoding.DecodeString(creds.Signature)
	if err != nil {
		return nil, ErrSignaturePublicKey
	}

	if err := pubKey.Verify([]byte(creds.Username), &ssh.Signature{ //nolint: exhaustruct
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

func newSession(conn *Conn, creds *Credentials, dim Dimensions, info Info) error {
	log.WithFields(log.Fields{
		"user":   creds.Username,
		"device": creds.Device,
		"cols":   dim.Cols,
		"rows":   dim.Rows,
	}).Info("handling web client request started")

	defer log.WithFields(log.Fields{
		"user":   creds.Username,
		"device": creds.Device,
		"cols":   dim.Cols,
		"rows":   dim.Rows,
	}).Info("handling web client request end")

	user := fmt.Sprintf("%s@%s", creds.Username, creds.Device)
	auth, err := getAuth(creds, magickey.GetRerefence())
	if err != nil {
		return ErrGetAuth
	}

	connection, err := ssh.Dial("tcp", "localhost:2222", &ssh.ClientConfig{ //nolint: exhaustruct
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		BannerCallback: func(message string) error {
			if message != "" {
				return NewBannerError(message)
			}

			return nil
		},
	})
	if err != nil {
		var e *BannerError

		// NOTICE: if the connection return a banner, wrap that message into an error and return to the session.
		if errors.As(err, &e) {
			return e
		}

		return ErrAuthentication
	}

	defer connection.Close()

	agent, err := connection.NewSession()
	if err != nil {
		return ErrSession
	}

	defer agent.Close()

	if err = agent.Setenv("IP_ADDRESS", info.IP); err != nil {
		return ErrEnvIPAddress
	}

	// NOTICE: when a SSH web shell is initialized, we set a env variable to the end user identify its origin.
	if err = agent.Setenv("WS", "true"); err != nil {
		return ErrEnvWS
	}

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

	if err := agent.RequestPty("xterm", dim.Rows, dim.Cols, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
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

				log.WithFields(
					log.Fields{
						"user":   creds.Username,
						"device": creds.Device,
						"ip":     info.IP,
					}).WithError(err).Error("failed to read the message from the client")

				return
			}

			switch message.Kind {
			case messageKindInput:
				buffer := message.Data.([]byte)

				if _, err := stdin.Write(buffer); err != nil {
					log.WithError(err).Error("failed to write the message data on the SSH session")

					return
				}
			case messageKindResize:
				dim := message.Data.(Dimensions)

				if err := agent.WindowChange(dim.Rows, dim.Cols); err != nil {
					log.WithFields(
						log.Fields{
							"user":   creds.Username,
							"device": creds.Device,
							"ip":     info.IP,
							"cols":   dim.Cols,
							"rows":   dim.Rows,
						},
					).WithError(err).Error("failed to change the seze of window for terminal session")

					return
				}
			}
		}
	}()

	go redirToWs(stdout, conn) // nolint:errcheck
	go io.Copy(conn, stderr)   //nolint:errcheck

	if err := agent.Wait(); err != nil {
		log.WithError(err).Warning("client remote command returned a error")
	}

	return nil
}

func redirToWs(rd io.Reader, ws io.ReadWriter) error {
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
