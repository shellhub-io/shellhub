package metadata

import (
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshsrvtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

// **NOTICE**:
// Each test case has a `setup` method responsible for creating and running the server.
// In these cases, we need to expose the session's context (and sometimes the error) outside the handler.
// To achieve this, we pass a reference to an existing context to write on.

func TestStore(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "succeeds in storing key/value",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							store(*ctx, "key", "val")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "val",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, ctx.Value("key").(string))
		})
	}
}

func TestStoreRequest(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "succeeds in storing request type",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							StoreRequest(*ctx, "exec")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "exec",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, ctx.Value(request).(string))
		})
	}
}

func TestMaybeStore(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "succeeds in storing key/value",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							maybeStore(*ctx, "key", "val")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "val",
		},
		{
			description: "succeeds in retrieving fingerprint when it is already defined",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue("key", "val")
							*ctx = s.Context()
							maybeStore(*ctx, "key", "other value")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "val",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, ctx.Value("key").(string))
		})
	}
}

func TestMaybeStoreSSHID(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "succeeds in storing sshid",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							MaybeStoreSSHID(*ctx, "namespace.00-00-00-00-00-00@localhost")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "namespace.00-00-00-00-00-00@localhost",
		},
		{
			description: "succeeds in retrieving sshid when it is already defined",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(sshid, "namespace.00-00-00-00-00-00@localhost")
							*ctx = s.Context()
							MaybeStoreSSHID(*ctx, "other sshid")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: "namespace.00-00-00-00-00-00@localhost",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
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
		setup       func(ctx *gliderssh.Context, err *error) *sshsrvtest.Conn
		expected    Expected
	}{
		{
			description: "fails when target is invalid",
			setup: func(ctx *gliderssh.Context, err *error) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							_, *err = MaybeStoreTarget(*ctx, "username")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: Expected{
				target: nil,
				err:    target.ErrSplitTarget,
			},
		},
		{
			description: "succeeds when target is valid",
			setup: func(ctx *gliderssh.Context, err *error) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
							_, *err = MaybeStoreTarget(*ctx, "username@namespace.00-00-00-00-00-00@localhost")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
			},
			expected: Expected{
				target: &target.Target{
					Username: "username",
					Data:     "namespace.00-00-00-00-00-00@localhost",
				},
				err: nil,
			},
		},
		{
			description: "succeeds in retrieving target when it is already defined",
			setup: func(ctx *gliderssh.Context, err *error) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(tag, &target.Target{Username: "username", Data: "namespace.00-00-00-00-00-00@localhost"})
							*ctx = s.Context()
							_, *err = MaybeStoreTarget(*ctx, "other value")
						},
					},
					&gossh.ClientConfig{
						User: "user",
						Auth: []gossh.AuthMethod{
							gossh.Password("123"),
						},
						HostKeyCallback: gossh.InsecureIgnoreHostKey(),
					},
				)
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

			srv := tc.setup(&ctx, &err)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			if err != nil {
				assert.Equal(t, tc.expected, Expected{nil, err})
			} else {
				assert.Equal(t, tc.expected, Expected{ctx.Value(tag).(*target.Target), err})
			}
		})
	}
}
