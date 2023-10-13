package metadata

import (
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
)

func TestStoreRequest(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "succeeds in storing request type",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						StoreRequest(*ctx, "exec")
					},
				})
			},
			expected: "exec",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(request).(string))
		})
	}
}

func TestStoreAuthenticationMethod(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    AuthMethod
	}{
		{
			description: "succeeds in storing authentication method",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						StoreAuthenticationMethod(*ctx, AuthMethodPasswd)
					},
				})
			},
			expected: AuthMethodPasswd,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(authentication).(AuthMethod))
		})
	}
}

func TestPassword(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "succeeds in storing password",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						StorePassword(*ctx, "secret")
					},
				})
			},
			expected: "secret",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(password).(string))
		})
	}
}

func TestMaybeStoreSSHID(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "succeeds in storing sshid",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						MaybeStoreSSHID(*ctx, "namespace.00-00-00-00-00-00@localhost")
					},
				})
			},
			expected: "namespace.00-00-00-00-00-00@localhost",
		},
		{
			description: "succeeds in storing sshid when is already defined",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(sshid, "namespace.00-00-00-00-00-00@localhost")

						*ctx = s.Context()
						MaybeStoreSSHID(*ctx, "namespace.00-00-00-00-00-00@localhost")
					},
				})
			},
			expected: "namespace.00-00-00-00-00-00@localhost",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(sshid).(string))
		})
	}
}

func TestMaybeStoreFingerprint(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "succeeds in storing fingerprint",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						MaybeStoreSSHID(*ctx, "fingerprint")
					},
				})
			},
			expected: "fingerprint",
		},
		{
			description: "succeeds in storing fingerprint when is already defined",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(sshid, "fingerprint")

						*ctx = s.Context()
						MaybeStoreSSHID(*ctx, "fingerprint")
					},
				})
			},
			expected: "fingerprint",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(sshid).(string))
		})
	}
}

func TestMaybeStoreTarget(t *testing.T) {
	type Expected struct {
		target *target.Target
		err    error
	}

	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context, err *error) *sshtest.SSHTest
		expected    Expected
	}{
		{
			description: "fails when sshid is invalid",
			setup: func(t *testing.T, ctx *gliderssh.Context, err *error) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						_, *err = MaybeStoreTarget(*ctx, "username")
					},
				})
			},
			expected: Expected{
				target: nil,
				err:    target.ErrSplitTarget,
			},
		},
		{
			description: "succeeds in storing target",
			setup: func(t *testing.T, ctx *gliderssh.Context, err *error) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						_, *err = MaybeStoreTarget(*ctx, "username@namespace.00-00-00-00-00-00@localhost")
					},
				})
			},
			expected: Expected{
				target: &target.Target{
					Username: "username",
					Data:     "namespace.00-00-00-00-00-00@localhost",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context
			var err error

			ssh := tc.setup(t, &ctx, &err)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			if err != nil {
				assert.Equal(t, tc.expected, Expected{nil, err})
			} else {
				assert.Equal(t, tc.expected, Expected{ctx.Value(tag).(*target.Target), err})
			}
		})
	}
}

func TestMaybeStoreEstablished(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    bool
	}{
		{
			description: "succeeds in storing authentication method",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
						MaybeStoreEstablished(*ctx, true)
					},
				})
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(established).(bool))
		})
	}
}
