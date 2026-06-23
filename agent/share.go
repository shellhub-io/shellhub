package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	creackpty "github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/server/modes/host/command"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

// ShareSession hosts a command inside a PTY and exposes its output as a public shareable terminal
// (tmate/upterm style). The local user keeps using the terminal normally while remote guests watch
// live through a web link; in collaborative mode they can also type.
//
// ShareOptions holds the tunables for a share session.
type ShareOptions struct {
	// Command to run; empty means the user's login shell.
	Command []string
	// Name is an optional label shown in the namespace's list.
	Name string
	// Writable enables collaborative input (guests can type).
	Writable bool
	// TTLSeconds controls the share lifetime: 0 = server default, <0 = no expiry, >0 = custom.
	TTLSeconds int
	// User is the host account the command runs as (resolved via the OS, like an SSH login).
	User string
}

type ShareSession struct {
	config *Config
	opts   ShareOptions
}

// NewShareSession creates a share session with the given options.
func NewShareSession(config *Config, opts ShareOptions) *ShareSession {
	return &ShareSession{config: config, opts: opts}
}

// wsWriter serializes writes to the upstream websocket so the output and resize goroutines never
// write to the connection concurrently (gorilla forbids concurrent writers).
type wsWriter struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *wsWriter) output(p []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.WriteMessage(websocket.BinaryMessage, p)
}

func (w *wsWriter) resize(cols, rows int) error {
	data, err := json.Marshal(map[string]any{"kind": "resize", "cols": cols, "rows": rows})
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	return w.conn.WriteMessage(websocket.TextMessage, data)
}

// buildCmd resolves the host user and builds the command to run inside the PTY. It uses the same
// host-execution path as SSH sessions (command.NewCmd), so in Docker mode the command runs on the
// host via nsenter/setpriv as the resolved user, rather than inside the agent container.
func (s *ShareSession) buildCmd(host string) (*exec.Cmd, error) {
	username := s.opts.User
	if username == "" {
		username = "root"
	}

	user, err := osauth.LookupUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to look up user %q: %w", username, err)
	}

	shell := user.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}

	argv := s.opts.Command
	if len(argv) == 0 {
		argv = []string{shell, "--login"}
	}

	return command.NewCmd(user, shell, term, host, nil, argv...), nil
}

// createShare registers the share on the server and returns the public link.
func (s *ShareSession) createShare(ctx context.Context, token string, cols, rows int) (*models.ShareCreateResponse, error) {
	label := strings.Join(s.opts.Command, " ")
	if label == "" {
		label = "login shell"
	}

	payload, err := json.Marshal(models.ShareCreateRequest{
		Name:       s.opts.Name,
		Command:    label,
		Writable:   s.opts.Writable,
		TTLSeconds: s.opts.TTLSeconds,
		Term:       os.Getenv("TERM"),
		Cols:       cols,
		Rows:       rows,
	})
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(s.config.ServerAddress, "/") + "/ssh/shares"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the server: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server rejected the share request with status %d", res.StatusCode)
	}

	var share models.ShareCreateResponse
	if err := json.NewDecoder(res.Body).Decode(&share); err != nil {
		return nil, fmt.Errorf("failed to decode the share response: %w", err)
	}

	return &share, nil
}

// dialStream opens the producer websocket used to push PTY output to the server.
func (s *ShareSession) dialStream(ctx context.Context, token, authToken string) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s/ssh/shares/%s/stream", strings.TrimRight(s.config.ServerAddress, "/"), token)

	conn, _, err := client.DialContext(ctx, url, http.Header{
		"Authorization": []string{"Bearer " + authToken},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open the share stream: %w", err)
	}

	return conn, nil
}

// Run authenticates the device, spawns the command in a PTY and streams it to the server until the
// command exits.
func (s *ShareSession) Run(ctx context.Context) error {
	ag, err := NewAgentWithConfig(s.config, new(HostMode))
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	if err := ag.Initialize(); err != nil {
		return fmt.Errorf("failed to authenticate device: %w", err)
	}

	authToken := ag.authData.Token

	cols, rows := 80, 24
	if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		cols, rows = w, h
	}

	// Start the PTY before registering the share, so a bad command never leaves an orphaned share
	// dangling in the namespace's list.
	cmd, err := s.buildCmd(ag.authData.Name)
	if err != nil {
		return err
	}

	ptmx, err := creackpty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	_ = creackpty.Setsize(ptmx, &creackpty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}) //nolint:gosec

	share, err := s.createShare(ctx, authToken, cols, rows)
	if err != nil {
		return err
	}

	conn, err := s.dialStream(ctx, share.Token, authToken)
	if err != nil {
		return err
	}
	defer conn.Close()

	upstream := &wsWriter{conn: conn}
	_ = upstream.resize(cols, rows)

	access := "public, read-only"
	if s.opts.Writable {
		access = "public, collaborative — guests can type"
	}

	fmt.Printf("\r\nSharing this terminal (%s). Anyone with this link can watch live:\r\n\r\n    %s\r\n\r\nThe share ends when this command exits (press Ctrl-D).\r\n\r\n", access, share.URL)

	// In collaborative mode, guest keystrokes arrive as binary frames on the producer connection;
	// write them straight into the PTY so they reach the running command.
	if s.opts.Writable {
		go func() {
			for {
				typ, data, err := conn.ReadMessage()
				if err != nil {
					return
				}

				if typ == websocket.BinaryMessage {
					_, _ = ptmx.Write(data)
				}
			}
		}()
	}

	// Put the local terminal in raw mode so the spawned command behaves interactively.
	if oldState, err := term.MakeRaw(int(os.Stdin.Fd())); err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState) //nolint:errcheck
	}

	// Forward local window resizes to the PTY and to the guests.
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)

	go func() {
		for range winch {
			w, h, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				continue
			}

			_ = creackpty.Setsize(ptmx, &creackpty.Winsize{Rows: uint16(h), Cols: uint16(w)}) //nolint:gosec
			_ = upstream.resize(w, h)
		}
	}()

	// Local user input always goes to the PTY. Remote guest input is additionally forwarded above
	// when the share is collaborative.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()

	// PTY output goes to the local terminal AND, best-effort, to the upstream stream. An upstream
	// failure must never tear down the local session, so it is handled independently of stdout.
	buf := make([]byte, 32*1024)
	upstreamAlive := true

	for {
		n, readErr := ptmx.Read(buf)
		if n > 0 {
			_, _ = os.Stdout.Write(buf[:n])

			if upstreamAlive {
				if err := upstream.output(buf[:n]); err != nil {
					upstreamAlive = false
					log.WithError(err).Debug("share stream closed; continuing local session")
				}
			}
		}

		if readErr != nil {
			break
		}
	}

	return nil
}
