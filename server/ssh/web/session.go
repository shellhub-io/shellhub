package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type BannerError struct {
	Message string
	kind    banner.Kind
}

// NewBannerError constructs a BannerError and classifies the message exactly
// once at construction time. Use this constructor — never build a BannerError
// literal — so that Message and kind always agree.
func NewBannerError(message string) *BannerError {
	kind, _ := banner.Classify(message)

	return &BannerError{
		Message: message,
		kind:    kind,
	}
}

func (b *BannerError) Error() string {
	return b.Message
}

// Kind returns the banner classification computed when the error was created.
func (b *BannerError) Kind() banner.Kind {
	return b.kind
}

// mapBannerError converts a BannerError to the appropriate sentinel error for
// the web client based on the banner Kind.
func mapBannerError(e *BannerError) error {
	switch e.Kind() {
	case banner.KindConnectionFailed:
		return ErrConnect
	case banner.KindAccessDenied:
		return ErrAccessDenied
	case banner.KindInvalidSSHID:
		return ErrInvalidSSHID
	default:
		if e.Message != "" {
			log.WithField("banner", e.Message).Warn("received unrecognized SSH banner; treating as connection failure")
		}

		return ErrConnect
	}
}

// bannerCallback handles the banners the gateway sends mid-handshake.
//
// A re-auth banner is not a failure: the gateway is holding this login open
// while it waits for the browser, so the code rides out to the console as a
// frame and the dial stays blocked until the decision lands. Every other banner
// is terminal, and becomes an error the caller maps for the client.
func bannerCallback(conn *Conn) func(string) error {
	return func(message string) error {
		if message == "" {
			return nil
		}

		if kind, code := banner.Classify(message); kind == banner.KindReauthRequired && code != "" {
			if _, err := conn.WriteMessage(&Message{Kind: messageKindReauth, Data: code}); err != nil {
				log.WithError(err).Error("failed to forward the re-auth code to the browser")

				return NewBannerError(message)
			}

			return nil
		}

		return NewBannerError(message)
	}
}

// getAuth gets the authentication methods from credentials.
func getAuth(ctx context.Context, service services.Service, conn *Conn, creds *Credentials) ([]ssh.AuthMethod, error) {
	// Identity mode: the browser presents its own enrolled key. Sign the SSH
	// challenge over the WebSocket with it (the private half never leaves the
	// browser); the gateway resolves the key to the identity via ssh_identities.
	// This bypasses the legacy public-key ACL below entirely.
	if creds.PublicKey != "" {
		pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(creds.PublicKey)) //nolint:dogsled
		if err != nil {
			return nil, ErrDataPublicKey
		}

		signer := &Signer{
			conn:      conn,
			publicKey: &pubKey,
		}

		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	}

	if creds.isPassword() {
		return []ssh.AuthMethod{ssh.Password(creds.Password)}, nil
	}

	device, err := service.GetDevice(ctx, scope.NewUnbounded(reasonWebHandoffDeviceResolve), models.UID(creds.Device))
	if err != nil {
		return nil, ErrFindDevice
	}

	key, err := service.GetPublicKey(ctx, creds.Fingerprint, device.TenantID)
	if err != nil {
		return nil, ErrFindPublicKey
	}

	usernameOK, err := service.EvaluateKeyUsername(ctx, key, creds.Username)
	if err != nil {
		return nil, ErrEvaluatePublicKey
	}

	filterOK, err := service.EvaluateKeyFilter(ctx, key, *device)
	if err != nil {
		return nil, ErrEvaluatePublicKey
	}

	if !usernameOK || !filterOK {
		return nil, ErrForbiddenPublicKey
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data) //nolint:dogsled
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
		return nil, errors.New("invalid signature response")
	}

	signed, ok := msg.Data.(string)
	if !ok {
		return nil, errors.New("data isn't a signed string")
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

func newSession(ctx context.Context, service services.Service, handoff *webhandoff.Store, conn *Conn, creds *Credentials, dim Dimensions, info Info) error {
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
	auth, err := getAuth(ctx, service, conn, creds)
	if err != nil {
		logger.WithError(err).Debug("failed to get the credentials")

		return ErrGetAuth
	}

	// The SSH handshake has nowhere to carry the browser's address or, in identity
	// mode, the logged-in account, so they are parked under the username about to
	// be dialled and claimed by the session on the other side of the loopback.
	handoff.Put(user, webhandoff.Data{
		Device: creds.Device,
		IP:     info.IP,
		UserID: creds.UserID,
	})

	connection, err := ssh.Dial("tcp", "localhost:2222", &ssh.ClientConfig{ //nolint:exhaustruct
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		BannerCallback:  bannerCallback(conn),
	})
	if err != nil {
		var e *BannerError

		// NOTE: if the connection returns an error banner, map it to a standard error for the web client
		// instead of forwarding the raw banner text (which is meant for native SSH clients).
		if errors.As(err, &e) {
			logger.WithError(e).Debug("failed to receive the connection banner")

			return mapBannerError(e)
		}

		// NOTE: Otherwise, any other error from the [ssh.Dial] process, we assume it was an authentication error,
		// keeping the real error internally to avoid exposing some sensitive data.
		logger.WithError(err).Debug("failed to dial to the ssh server")

		return ErrAuthentication
	}

	defer connection.Close() //nolint:errcheck

	// Ask the SSH server for this connection's session UID and relay it to the web
	// client, so a client-side recording can be tied to its server session.
	if ok, reply, err := connection.SendRequest("session-uid@shellhub.io", true, nil); err == nil && ok {
		if _, err := conn.WriteMessage(&Message{Kind: messageKindSession, Data: string(reply)}); err != nil {
			logger.WithError(err).Debug("failed to send the session UID to the web client")
		}
	}

	agent, err := connection.NewSession()
	if err != nil {
		logger.WithError(err).Debug("failed to create a new session")

		return ErrSession
	}

	defer agent.Close() //nolint:errcheck

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

	if err := agent.RequestPty("xterm", int(dim.Rows), int(dim.Cols), ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		logger.WithError(err).Debug("failed to request the pty on session")

		return ErrPty
	}

	if err := agent.Shell(); err != nil {
		logger.WithError(err).Debug("failed to request the shell on session")

		return ErrShell
	}

	go func() {
		defer agent.Close() //nolint:errcheck

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
					logger.Error("input message data is not a string")

					return
				}

				if _, err := stdin.Write([]byte(buffer)); err != nil {
					logger.WithError(err).Error("failed to write the message data on the SSH session")

					return
				}
			case messageKindResize:
				dim, ok := message.Data.(Dimensions)
				if !ok {
					logger.Error("resize message data is not a dimension")

					return
				}

				if err := agent.WindowChange(int(dim.Rows), int(dim.Cols)); err != nil {
					logger.WithError(err).Error("failed to change the size of window for terminal session")

					return
				}
			default:
				// The client sends the other kinds only in the opposite direction.
			}
		}
	}()

	go redirToWs(stdout, conn) //nolint:errcheck
	go io.Copy(conn, stderr)   //nolint:errcheck

	if err := agent.Wait(); err != nil {
		logger.WithError(err).Warning("client remote command returned a error")
	}

	return nil
}

func redirToWs(rd io.Reader, ws *Conn) error {
	// TODO: Evaluate refactoring this function to improve its readability.
	var buf [32 * 1024]byte
	var start, end, buflen int

	for {
		nr, err := rd.Read(buf[start:])
		if err != nil {
			return err
		}

		if nr == 0 {
			// NOTE: "Callers should treat a return of 0 and nil as indicating that nothing happened; in particular it
			// does not indicate EOF", in such a case, the caller should not interpret it as EOF, but instead wait for
			// more data.
			//
			// https://pkg.go.dev/io#Reader
			continue
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

		if end < 0 {
			// NOTE: This workround is to avoid a panic in case the end is negative, which would lead to a negative slice.
			// This situation can happen when the buffer contains only UTF-8 continuation bytes, which are bytes that
			// cannot start a valid UTF-8 rune. In such cases, the loop above will not find a valid rune start and
			// will leave `end` as -1.
			//
			// https://datatracker.ietf.org/doc/html/rfc3629#section-3
			log.WithFields(log.Fields{
				"buf":    buf,
				"buflen": buflen,
				"start":  start,
				"end":    end,
				"nr":     nr,
			}).Warn("end is negative, skipping write to avoid panic")

			end = 0
		}

		if _, err = ws.WriteBinary([]byte(string(bytes.Runes(buf[0:end])))); err != nil {
			return err
		}

		start = buflen - end

		if start > 0 {
			// copy remaning read bytes from the end to the beginning of a buffer
			// so that we will get normal bytes
			for i := range start {
				buf[i] = buf[end+i]
			}
		}
	}
}
