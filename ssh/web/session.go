package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/ssh/web/share"
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
func getAuth(ctx context.Context, conn *Conn, creds *Credentials) ([]ssh.AuthMethod, error) {
	if creds.isPassword() {
		return []ssh.AuthMethod{ssh.Password(creds.Password)}, nil
	}

	cli, err := internalclient.NewClient(nil)
	if err != nil {
		return nil, err
	}

	// Trys to get a device from the API.
	device, err := cli.GetDevice(ctx, creds.Device)
	if err != nil {
		return nil, ErrFindDevice
	}

	// Trys to get a public key from the API.
	key, err := cli.GetPublicKey(ctx, creds.Fingerprint, device.TenantID)
	if err != nil {
		return nil, ErrFindPublicKey
	}

	// Trys to evaluate the public key from the API.
	ok, err := cli.EvaluateKey(ctx, creds.Fingerprint, device, creds.Username)
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

// startShare exposes the current console session as a public shareable terminal. It resolves the
// device's tenant, registers an in-process share whose producer is this session, and returns the
// hub (to feed output / drain guest input), a close function and the share token.
func startShare(ctx context.Context, shares *share.Registry, deviceUID string, dim Dimensions, req ShareRequest) (*share.Hub, func(), string, error) {
	cli, err := internalclient.NewClient(nil)
	if err != nil {
		return nil, nil, "", err
	}

	device, err := cli.GetDevice(ctx, deviceUID)
	if err != nil {
		return nil, nil, "", err
	}

	token, hub, closeFn := shares.CreateLocal(deviceUID, device.TenantID, models.ShareCreateRequest{
		Name:       req.Name,
		Command:    "console session",
		Writable:   req.Writable,
		TTLSeconds: req.TTL,
	})

	hub.Resize(share.Dimensions{Cols: int(dim.Cols), Rows: int(dim.Rows)})

	return hub, closeFn, token, nil
}

func newSession(ctx context.Context, cache cache.Cache, conn *Conn, creds *Credentials, dim Dimensions, info Info, shares *share.Registry) error {
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
	auth, err := getAuth(ctx, conn, creds)
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

		// NOTE: if the connection returns an error banner, map it to a standard error for the web client
		// instead of forwarding the raw banner text (which is meant for native SSH clients).
		if errors.As(err, &e) {
			logger.WithError(e).Debug("failed to receive the connection banner")

			switch {
			case strings.Contains(e.Message, "Connection Failed"):
				return ErrConnect
			case strings.Contains(e.Message, "Access Denied"):
				return ErrFindDevice
			default:
				return ErrConnect
			}
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

	// shareHub holds the share's hub once the user shares this session (nil until then). The output
	// goroutine reads it to tee output; the input goroutine writes it.
	var shareHub atomic.Pointer[share.Hub]

	go func() {
		defer agent.Close()

		// currentDim tracks the live terminal size so a share started mid-session matches geometry.
		currentDim := dim

		// shareClose tears the share down when this session ends. Only this goroutine touches it.
		var shareClose func()
		defer func() {
			if shareClose != nil {
				shareClose()
			}
		}()

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
				buffer := message.Data.(string)

				if _, err := stdin.Write([]byte(buffer)); err != nil {
					logger.WithError(err).Error("failed to write the message data on the SSH session")

					return
				}
			case messageKindResize:
				dim := message.Data.(Dimensions)
				currentDim = dim

				if err := agent.WindowChange(int(dim.Rows), int(dim.Cols)); err != nil {
					logger.WithError(err).Error("failed to change the size of window for terminal session")

					return
				}

				if h := shareHub.Load(); h != nil {
					h.Resize(share.Dimensions{Cols: int(dim.Cols), Rows: int(dim.Rows)})
				}
			case messageKindShare:
				if shareHub.Load() != nil {
					continue // already shared
				}

				req := message.Data.(ShareRequest)

				hub, closeFn, token, err := startShare(ctx, shares, creds.Device, currentDim, req)
				if err != nil {
					logger.WithError(err).Error("failed to start the share")

					continue
				}

				shareClose = closeFn
				shareHub.Store(hub)

				// In collaborative mode, guest keystrokes flow into the same PTY stdin as the local user.
				if req.Writable {
					go func() {
						for {
							select {
							case <-hub.Done():
								return
							case data := <-hub.Input():
								if _, err := stdin.Write(data); err != nil {
									return
								}
							}
						}
					}()
				}

				if _, err := conn.WriteMessage(&Message{Kind: messageKindShare, Data: token}); err != nil {
					logger.WithError(err).Error("failed to send the share token to the client")
				}
			}
		}
	}()

	go redirToWs(stdout, conn, &shareHub) // nolint:errcheck
	go io.Copy(conn, stderr)              //nolint:errcheck

	if err := agent.Wait(); err != nil {
		logger.WithError(err).Warning("client remote command returned a error")
	}

	return nil
}

func redirToWs(rd io.Reader, ws *Conn, shareHub *atomic.Pointer[share.Hub]) error {
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

		chunk := []byte(string(bytes.Runes(buf[0:end])))

		if _, err = ws.WriteBinary(chunk); err != nil {
			return err
		}

		// Tee the same output to the share hub when this session is being shared.
		if h := shareHub.Load(); h != nil {
			h.Output(chunk)
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
