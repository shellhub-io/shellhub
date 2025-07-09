package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	sshTTYSpeed      = 14400
	utf8MinRuneBytes = 6
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
func getAuth(conn *Conn, creds *Credentials) ([]ssh.AuthMethod, error) {
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

	signer := &Signer{
		conn:      conn,
		publicKey: &pubKey,
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}

type Signer struct {
	conn      *Conn
	publicKey *ssh.PublicKey
}

func (s *Signer) PublicKey() ssh.PublicKey {
	return *s.publicKey
}

func (s *Signer) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	dataB64 := base64.StdEncoding.EncodeToString(data)
	if _, err := s.conn.WriteMessage(&Message{Kind: messageKindSignature, Data: dataB64}); err != nil {
		return nil, err
	}

	var msg Message
	if _, err := s.conn.ReadMessage(&msg); err != nil {
		return nil, fmt.Errorf("invalid signature response")
	}

	signed, ok := msg.Data.(string)
	if !ok {
		return nil, fmt.Errorf("data isn't a signed string")
	}

	blob, err := base64.StdEncoding.DecodeString(signed)
	if err != nil {
		return nil, err
	}

	return &ssh.Signature{
		Format: s.PublicKey().Type(),
		Blob:   blob,
	}, nil
}

func newSession(ctx context.Context, cache cache.Cache, conn *Conn, creds *Credentials, dim Dimensions, info Info) error {
	logger := log.WithFields(log.Fields{
		"user":   creds.Username,
		"device": creds.Device,
		"cols":   dim.Cols,
		"rows":   dim.Rows,
		"ip":     info.IP,
	})

	logger.Info("handling web client request started")

	defer logger.Info("handling web client request end")

	uuid := uuid.Generate()

	user := fmt.Sprintf("%s@%s", creds.Username, uuid)
	auth, err := getAuth(conn, creds)
	if err != nil {
		logger.WithError(err).Debug("failed to get the credentials")

		return ErrGetAuth
	}

	if err := cache.Set(ctx, "web-ip/"+user, fmt.Sprintf("%s:%s", creds.Device, info.IP), 1*time.Minute); err != nil {
		logger.WithError(err).Debug("failed to set the session IP on the cache")

		return err
	}

	defer cache.Delete(ctx, "web-ip/"+user) //nolint:errcheck

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

		// NOTE: if the connection return a error banner, wrap that message into an error and return to the session.
		if errors.As(err, &e) {
			logger.WithError(e).Debug("failed to receive the connection banner")

			return e
		}

		// NOTE: Otherwise, any other error from the [ssh.Dial] process, we assume it was an authentication error,
		// keeping the real error internally to avoid exposing some sensitive data.
		logger.WithError(err).Debug("failed to dial to the ssh server")

		return ErrAuthentication
	}

	defer connection.Close()

	agent, err := connection.NewSession()
	if err != nil {
		logger.WithError(err).Debug("failed to create a new session")

		return ErrSession
	}

	defer agent.Close()

	stdin, err := agent.StdinPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stdin pipe")

		return err
	}

	stdout, err := agent.StdoutPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stdout pipe")

		return err
	}

	stderr, err := agent.StderrPipe()
	if err != nil {
		logger.WithError(err).Debug("failed to create the stderr pipe")

		return err
	}

	if err := agent.RequestPty("xterm", dim.Rows, dim.Cols, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: sshTTYSpeed,
		ssh.TTY_OP_OSPEED: sshTTYSpeed,
	}); err != nil {
		logger.WithError(err).Debug("failed to request the pty on session")

		return ErrPty
	}

	if err := agent.Shell(); err != nil {
		logger.WithError(err).Debug("failed to request the shell on session")

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
				buffer := message.Data.([]byte)

				if _, err := stdin.Write(buffer); err != nil {
					logger.WithError(err).Error("failed to write the message data on the SSH session")

					return
				}
			case messageKindResize:
				dim := message.Data.(Dimensions)

				if err := agent.WindowChange(dim.Rows, dim.Cols); err != nil {
					logger.WithError(err).Error("failed to change the size of window for terminal session")

					return
				}
			}
		}
	}()

	go redirToWs(stdout, conn) // nolint:errcheck
	go io.Copy(conn, stderr)   //nolint:errcheck

	if err := agent.Wait(); err != nil {
		logger.WithError(err).Warning("client remote command returned a error")
	}

	return nil
}

func redirToWs(rd io.Reader, ws *Conn) error {
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

			if buflen-end >= utf8MinRuneBytes {
				end = nr

				break
			}
		}

		if _, err = ws.WriteBinary([]byte(string(bytes.Runes(buf[0:end])))); err != nil {
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
