package session

import (
	"errors"
	"net"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	internalclientMocks "github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	httptunnelMocks "github.com/shellhub-io/shellhub/pkg/httptunnel/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	metadataMocks "github.com/shellhub-io/shellhub/ssh/pkg/metadata/mocks"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

type mockConn struct{}

func (mc *mockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (mc *mockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (mc *mockConn) Close() error {
	return nil
}

func (mc *mockConn) LocalAddr() net.Addr {
	return nil
}

func (mc *mockConn) RemoteAddr() net.Addr {
	return nil
}

func (mc *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (mc *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (mc *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestNewSession(t *testing.T) {
	type Expected struct {
		sess *Session
		err  error
	}

	tunnelMock := new(httptunnelMocks.Tunneler)
	connMock := &mockConn{}

	cases := []struct {
		description string
		setup       func(t *testing.T) (gliderssh.Session, gliderssh.Context)
		mocks       func(ctx gliderssh.Context)
		expected    func(client gliderssh.Session) Expected
	}{
		{
			description: "fails when when a firewall connection errors occours",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()

				apiMock.On("FirewallEvaluate", map[string]string{"ip_address": "127.0.0.1", "username": "username"}).
					Return(errors.New("error")).
					Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrFirewallUnknown,
				}
			},
		},
		{
			description: "fails when a firewall block occours",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()

				apiMock.On("FirewallEvaluate", map[string]string{"ip_address": "127.0.0.1", "username": "username"}).
					Return(internalclient.ErrFirewallBlock).
					Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrFirewallBlock,
				}
			},
		},
		{
			description: "fails when an unknown firewall error occours",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()

				apiMock.On("FirewallEvaluate", map[string]string{"ip_address": "127.0.0.1", "username": "username"}).
					Return(internalclient.ErrFirewallConnection).
					Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrFirewallConnection,
				}
			},
		},
		{
			description: "fails when could not found device",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()

				apiMock.On("FirewallEvaluate", map[string]string{"ip_address": "127.0.0.1", "username": "username"}).
					Return(nil).
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_BILLING").
					Return("true").
					Once()

				apiMock.On("GetDevice", "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrFindDevice,
				}
			},
		},
		{
			description: "fails when billing evaluation fails",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()

				apiMock.On("FirewallEvaluate", map[string]string{"ip_address": "127.0.0.1", "username": "username"}).
					Return(nil).
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.On("Get", "SHELLHUB_BILLING").
					Return("true").
					Once()

				apiMock.On("GetDevice", "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c").
					Return(
						&models.Device{
							UID:      "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
							TenantID: "00000000-0000-4000-0000-000000000000",
						},
						nil,
					).
					Once()
				apiMock.On("BillingEvaluate", "00000000-0000-4000-0000-000000000000").
					Return(&models.BillingEvaluation{CanConnect: false}, 200, nil).
					Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrBillingBlock,
				}
			},
		},
		{
			description: "fails when dialing fails",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.On("Get", "SHELLHUB_BILLING").
					Return("false").
					Once()

				tunnelMock.On("Dial", ctx, "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c").Return(nil, errors.New("error")).Once()
			},
			expected: func(ses gliderssh.Session) Expected {
				return Expected{
					sess: nil,
					err:  ErrDial,
				}
			},
		},
		{
			description: "succeeds to create a new session",
			setup: func(t *testing.T) (gliderssh.Session, gliderssh.Context) {
				var ses gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						ses = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return ses, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				envMock := new(envMocks.Backend)
				envs.DefaultBackend = envMock
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock
				apiMock := new(internalclientMocks.Client)

				metadataMock.On("RestoreDevice", ctx).
					Return(
						&models.Device{
							UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						},
					).
					Once()
				metadataMock.On("RestoreTarget", ctx).
					Return(
						&target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						},
					).
					Once()
				metadataMock.On("RestoreAPI", ctx).
					Return(apiMock).
					Once()
				metadataMock.On("RestoreLookup", ctx).
					Return(make(map[string]string)).
					Once()
				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
				metadataMock.On("RestoreUID", ctx).
					Return("448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()

				envMock.On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.On("Get", "SHELLHUB_BILLING").
					Return("false").
					Once()

				tunnelMock.On("Dial", ctx, "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c").Return(connMock, nil).Once()
			},
			expected: func(client gliderssh.Session) Expected {
				return Expected{
					sess: &Session{
						Client:        client,
						Username:      "username",
						Device:        "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						UID:           "448e21ab27144fde4fe112199256db2c206dbba9c23cca1319d573fbfdc8a7b2",
						IPAddress:     "127.0.0.1",
						Type:          "exec",
						Authenticated: false,
						Lookup: map[string]string{
							"ip_address": "127.0.0.1",
							"username":   "username",
						},
						Dialed: connMock,
					},
					err: nil,
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			client, ctx := tc.setup(t)
			tc.mocks(ctx)

			session, err := NewSession(client, httptunnel.Tunneler(tunnelMock))
			assert.Equal(t, tc.expected(client), Expected{session, err})
		})
	}

	tunnelMock.AssertExpectations(t)
}

func TestSetPty(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T) *Session
		expected    bool
	}{
		{
			description: "sets PTY to false when no pty is requested",
			setup: func(t *testing.T) *Session {
				var client gliderssh.Session

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return &Session{
					Client: client,
				}
			},
			expected: false,
		},
		{
			description: "sets PTY to true when an pty is requested in exec mode",
			setup: func(t *testing.T) *Session {
				var client gliderssh.Session
				done := make(chan bool)

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						close(done)
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.RequestPty("xterm", 40, 80, gossh.TerminalModes{}))
				assert.NoError(t, ssh.Session.Run("cmd"))
				<-done

				return &Session{Client: client}
			},
			expected: true,
		},
		{
			description: "sets PTY to true when an pty is requested in shell mode",
			setup: func(t *testing.T) *Session {
				var client gliderssh.Session
				done := make(chan bool)

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						close(done)
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.RequestPty("xterm", 40, 80, gossh.TerminalModes{}))
				assert.NoError(t, ssh.Session.Shell())
				<-done

				return &Session{Client: client}
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sess := tc.setup(t)

			sess.setPty()
			assert.Equal(t, tc.expected, sess.Pty)
		})
	}
}

func TestSetType(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T) (*Session, gliderssh.Context)
		mocks       func(ctx gliderssh.Context)
		expected    string
	}{
		// TODO: test Web
		// TODO: test SFTP
		{
			description: "should return SCP",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("scp cmd"))

				return &Session{Client: client}, ctx
			},
			mocks: func(_ gliderssh.Context) {
			},
			expected: SCP,
		},
		{
			description: "should return HereDoc",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return &Session{Client: client}, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("RestoreRequest", ctx).
					Return("shell").
					Once()
			},
			expected: HereDoc,
		},
		{
			description: "should return Exec with no pty",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run("cmd"))

				return &Session{Client: client}, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
			},
			expected: Exec,
		},
		{
			description: "should return Exec with pty",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.RequestPty("xterm", 40, 80, gossh.TerminalModes{}))
				assert.NoError(t, ssh.Session.Run("cmd"))

				return &Session{Client: client}, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("RestoreRequest", ctx).
					Return("exec").
					Once()
			},
			expected: Exec,
		},
		{
			description: "should return Term",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context
				done := make(chan bool)

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
						close(done)
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.RequestPty("xterm", 40, 80, gossh.TerminalModes{}))
				assert.NoError(t, ssh.Session.Shell())
				<-done

				return &Session{Client: client, Pty: true}, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("RestoreRequest", ctx).
					Return("shell").
					Once()
			},
			expected: Term,
		},
		{
			description: "should return Unknown",
			setup: func(t *testing.T) (*Session, gliderssh.Context) {
				var client gliderssh.Session
				var ctx gliderssh.Context

				ssh := sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						client = s
						ctx = s.Context()
					},
				})
				defer ssh.Teardown()

				assert.NoError(t, ssh.Session.Run(""))

				return &Session{Client: client}, ctx
			},
			mocks: func(ctx gliderssh.Context) {
				metadataMock := new(metadataMocks.Metadata)
				metadata.Backend = metadataMock

				metadataMock.On("RestoreRequest", ctx).
					Return("unk").
					Once()
			},
			expected: Unk,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			client, ctx := tc.setup(t)
			tc.mocks(ctx)

			client.setType()
			assert.Equal(t, tc.expected, client.Type)
		})
	}
}

// func TestNewClientConnWithDeadline(t *testing.T) {
//
// }

// func TestRegister(t *testing.T) {
//
// }

// func TestFinish(t *testing.T) {
//
// }

func TestLoadEnv(t *testing.T) {
	cases := []struct {
		description string
		env         []string
		expected    map[string]string
	}{
		{
			description: "succeeds to create a map of strings",
			env:         []string{"hello=world"},
			expected:    map[string]string{"hello": "world"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			env := loadEnv(tc.env)
			assert.Equal(t, tc.expected, env)
		})
	}
}
