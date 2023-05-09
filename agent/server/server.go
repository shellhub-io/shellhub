package server

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/server/command"
	"github.com/shellhub-io/shellhub/agent/server/utmp"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// List of SSH subsystems names supported by the agent.
const (
	// SFTPSubsystemName is the name of the SFTP subsystem.
	SFTPSubsystemName = "sftp"
)

type sshConn struct {
	net.Conn
	closeCallback func(string)
	ctx           gliderssh.Context
}

func (c *sshConn) Close() error {
	if id, ok := c.ctx.Value(gliderssh.ContextKeySessionID).(string); ok {
		c.closeCallback(id)
	}

	return c.Conn.Close()
}

type Server struct {
	sshd               *gliderssh.Server
	api                client.Client
	authData           *models.DeviceAuthResponse
	cmds               map[string]*exec.Cmd
	Sessions           map[string]net.Conn
	deviceName         string
	mu                 sync.Mutex
	keepAliveInterval  int
	singleUserPassword string
}

// NewServer creates a new server SSH agent server.
func NewServer(api client.Client, authData *models.DeviceAuthResponse, privateKey string, keepAliveInterval int, singleUserPassword string) *Server {
	server := &Server{
		api:               api,
		authData:          authData,
		cmds:              make(map[string]*exec.Cmd),
		Sessions:          make(map[string]net.Conn),
		keepAliveInterval: keepAliveInterval,
	}

	server.sshd = &gliderssh.Server{
		PasswordHandler:        server.passwordHandler,
		PublicKeyHandler:       server.publicKeyHandler,
		Handler:                server.sessionHandler,
		SessionRequestCallback: server.sessionRequestCallback,
		RequestHandlers:        gliderssh.DefaultRequestHandlers,
		SubsystemHandlers: map[string]gliderssh.SubsystemHandler{
			SFTPSubsystemName: server.sftpSubsystemHandler,
		},
		ConnCallback: func(ctx gliderssh.Context, conn net.Conn) net.Conn {
			closeCallback := func(id string) {
				server.mu.Lock()
				defer server.mu.Unlock()

				if v, ok := server.cmds[id]; ok {
					v.Process.Kill() // nolint:errcheck
					delete(server.cmds, id)
				}
			}

			return &sshConn{conn, closeCallback, ctx}
		},
		LocalPortForwardingCallback: func(ctx gliderssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx gliderssh.Context, destinationHost string, destinationPort uint32) bool {
			return false
		},
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			"session":       gliderssh.DefaultSessionHandler,
			"direct-tcpip":  gliderssh.DirectTCPIPHandler,
			"dynamic-tcpip": gliderssh.DirectTCPIPHandler,
		},
	}

	err := server.sshd.SetOption(gliderssh.HostKeyFile(privateKey))
	if err != nil {
		log.Warn(err)
	}

	return server
}

// startKeepAlive sends a keep alive message to the server every in keepAliveInterval seconds.
func (s *Server) startKeepAliveLoop(session gliderssh.Session) {
	interval := time.Duration(s.keepAliveInterval) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.WithFields(log.Fields{
		"interval": interval,
	}).Debug("Starting keep alive loop")

loop:
	for {
		select {
		case <-ticker.C:
			if conn, ok := session.Context().Value(gliderssh.ContextKeyConn).(gossh.Conn); ok {
				if _, _, err := conn.SendRequest("keepalive", false, nil); err != nil {
					log.Error(err)
				}
			}
		case <-session.Context().Done():
			log.Debug("Stopping keep alive loop after session closed")
			ticker.Stop()

			break loop
		}
	}
}

func (s *Server) sessionHandler(session gliderssh.Session) {
	sspty, winCh, isPty := session.Pty()

	log.Info("New session request")

	go s.startKeepAliveLoop(session)
	requestType := session.Context().Value("request_type").(string) //nolint:forcetypeassert

	switch {
	case isPty:
		scmd := newShellCmd(s, session.User(), sspty.Term)

		pts, err := startPty(scmd, session, winCh)
		if err != nil {
			log.Warn(err)
		}

		u := osauth.LookupUser(session.User())

		err = os.Chown(pts.Name(), int(u.UID), -1)
		if err != nil {
			log.Warn(err)
		}

		remoteAddr := session.RemoteAddr()

		log.WithFields(log.Fields{
			"user":       session.User(),
			"pty":        pts.Name(),
			"ispty":      isPty,
			"remoteaddr": remoteAddr,
			"localaddr":  session.LocalAddr(),
		}).Info("Session started")

		ut := utmp.UtmpStartSession(
			pts.Name(),
			session.User(),
			remoteAddr.String(),
		)

		s.mu.Lock()
		s.cmds[session.Context().Value(gliderssh.ContextKeySessionID).(string)] = scmd
		s.mu.Unlock()

		if err := scmd.Wait(); err != nil {
			log.Warn(err)
		}

		log.WithFields(log.Fields{
			"user":       session.User(),
			"pty":        pts.Name(),
			"remoteaddr": remoteAddr,
			"localaddr":  session.LocalAddr(),
		}).Info("Session ended")

		utmp.UtmpEndSession(ut)
	case !isPty && requestType == "shell":
		cmd := newShellCmd(s, session.User(), "")

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		serverConn, ok := session.Context().Value(gliderssh.ContextKeyConn).(*gossh.ServerConn)
		if !ok {
			return
		}

		go func() {
			serverConn.Wait()  // nolint:errcheck
			cmd.Process.Kill() // nolint:errcheck
		}()

		log.WithFields(log.Fields{
			"user":        session.User(),
			"ispty":       isPty,
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command started")

		err := cmd.Start()
		if err != nil {
			log.Warn(err)
		}

		go func() {
			if _, err := io.Copy(stdin, session); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}

			stdin.Close()
		}()

		go func() {
			combinedOutput := io.MultiReader(stdout, stderr)
			if _, err := io.Copy(session, combinedOutput); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}
		}()

		err = cmd.Wait()
		if err != nil {
			log.Warn(err)
		}

		session.Exit(cmd.ProcessState.ExitCode()) //nolint:errcheck

		log.WithFields(log.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command ended")
	default:
		u := osauth.LookupUser(session.User())
		if len(session.Command()) == 0 {
			log.WithFields(log.Fields{
				"user":      session.User(),
				"localaddr": session.LocalAddr(),
			}).Error("None command was received")

			log.Info("Session ended")
			_ = session.Exit(1)

			return
		}

		cmd := command.NewCmd(u, "", "", s.deviceName, session.Command()...)

		stdout, _ := cmd.StdoutPipe()
		stdin, _ := cmd.StdinPipe()
		stderr, _ := cmd.StderrPipe()

		serverConn, ok := session.Context().Value(gliderssh.ContextKeyConn).(*gossh.ServerConn)
		if !ok {
			return
		}

		log.WithFields(log.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command started")

		err := cmd.Start()
		if err != nil {
			log.Warn(err)
		}

		go func() {
			serverConn.Wait()  // nolint:errcheck
			cmd.Process.Kill() // nolint:errcheck
		}()

		go func() {
			if _, err := io.Copy(stdin, session); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}

			stdin.Close()
		}()

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			combinedOutput := io.MultiReader(stdout, stderr)
			if _, err := io.Copy(session, combinedOutput); err != nil {
				fmt.Println(err) //nolint:forbidigo
			}

			wg.Done()
		}()

		wg.Wait()

		err = cmd.Wait()
		if err != nil {
			log.Warn(err)
		}

		session.Exit(cmd.ProcessState.ExitCode()) //nolint:errcheck

		log.WithFields(log.Fields{
			"user":        session.User(),
			"remoteaddr":  session.RemoteAddr(),
			"localaddr":   session.LocalAddr(),
			"Raw command": session.RawCommand(),
		}).Info("Command ended")
	}
}

func (s *Server) passwordHandler(ctx gliderssh.Context, pass string) bool {
	log := log.WithFields(log.Fields{
		"user": ctx.User(),
	})
	var ok bool

	if s.singleUserPassword == "" {
		ok = osauth.AuthUser(ctx.User(), pass)
	} else {
		ok = osauth.VerifyPasswordHash(s.singleUserPassword, pass)
	}

	if ok {
		log.Info("Accepted password")
	} else {
		log.Info("Failed password")
	}

	return ok
}

func (s *Server) publicKeyHandler(ctx gliderssh.Context, key gliderssh.PublicKey) bool {
	if osauth.LookupUser(ctx.User()) == nil {
		return false
	}

	type Signature struct {
		Username  string
		Namespace string
	}

	sig := &Signature{
		Username:  ctx.User(),
		Namespace: s.deviceName,
	}

	sigBytes, err := json.Marshal(sig)
	if err != nil {
		return false
	}

	sigHash := sha256.Sum256(sigBytes)

	res, err := s.api.AuthPublicKey(&models.PublicKeyAuthRequest{
		Fingerprint: gossh.FingerprintLegacyMD5(key),
		Data:        string(sigBytes),
	}, s.authData.Token)
	if err != nil {
		return false
	}

	digest, err := base64.StdEncoding.DecodeString(res.Signature)
	if err != nil {
		return false
	}

	cryptoKey, ok := key.(gossh.CryptoPublicKey)
	if !ok {
		return false
	}

	pubCrypto := cryptoKey.CryptoPublicKey()

	pubKey, ok := pubCrypto.(*rsa.PublicKey)
	if !ok {
		return false
	}

	if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sigHash[:], digest); err != nil {
		return false
	}

	return true
}

func (s *Server) sessionRequestCallback(session gliderssh.Session, requestType string) bool {
	session.Context().SetValue("request_type", requestType)

	return true
}

// sftpSubsystemHandler handles the SFTP subsystem session.
func (s *Server) sftpSubsystemHandler(session gliderssh.Session) {
	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session started")
	defer session.Close()

	cmd := exec.Command("/proc/self/exe", []string{"sftp"}...)

	looked, err := user.Lookup(session.User())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to lookup user")

		return
	}

	home := fmt.Sprintf("HOME=%s", looked.HomeDir)
	gid := fmt.Sprintf("GID=%s", looked.Gid)
	uid := fmt.Sprintf("UID=%s", looked.Uid)

	cmd.Env = append(cmd.Env, home)
	cmd.Env = append(cmd.Env, gid)
	cmd.Env = append(cmd.Env, uid)

	input, err := cmd.StdinPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdin pipe")

		return
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stdout pipe")

		return
	}

	erro, err := cmd.StderrPipe()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to get stderr pipe")

		return
	}

	if err := cmd.Start(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to start command")

		return
	}

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying input to session")

		if _, err := io.Copy(input, session); err != nil && err != io.EOF {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdin to command")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing input to session ends")

		input.Close()
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying output to session")

		if _, err := io.Copy(session, output); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stdout to session")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing output to session ends")
	}()

	go func() {
		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("copying error to session")

		if _, err := io.Copy(session, erro); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": session.Context().User(),
			}).Error("Failed to copy stderr to session")

			return
		}

		log.WithFields(log.Fields{
			"user": session.Context().User(),
		}).Trace("closing error to session ends")
	}()

	go s.startKeepAliveLoop(session)

	if err = cmd.Wait(); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": session.Context().User(),
		}).Error("Failed to wait command")

		return
	}

	log.WithFields(log.Fields{
		"user": session.Context().User(),
	}).Info("SFTP session closed")
}

func (s *Server) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

func (s *Server) SetDeviceName(name string) {
	s.deviceName = name
}

func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func (s *Server) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}

func newShellCmd(s *Server, username, term string) *exec.Cmd {
	shell := os.Getenv("SHELL")

	user := osauth.LookupUser(username)

	if shell == "" {
		shell = user.Shell
	}

	if term == "" {
		term = "xterm"
	}

	cmd := command.NewCmd(user, shell, term, s.deviceName, shell, "--login")

	return cmd
}
