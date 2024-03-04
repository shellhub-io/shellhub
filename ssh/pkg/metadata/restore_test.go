package metadata

import (
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshsrvtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

// **NOTICE**:
// Each test case has a `setup` method responsible for creating and running the server.
// In these cases, we need to expose the session's context outside the handler.
// To achieve this, we pass a reference to an existing context to write on.

func TestRestore(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "succeeds",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(request, "exec")
							*ctx = s.Context()
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
			assert.Equal(t, tc.expected, restore(ctx, request))
		})
	}
}

func TestRestoreRequest(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    string
	}{
		{
			description: "fails when request is not set",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
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
			expected: "",
		},
		{
			description: "succeeds in restoring request type",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(request, "exec")
							*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestoreRequest(ctx))
		})
	}
}

func TestRestoreTarget(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    *target.Target
	}{
		{
			description: "fails when target is not set",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
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
			expected: nil,
		},
		{
			description: "succeeds in restoring target",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(tag, &target.Target{
								Username: "username",
								Data:     "namespace.00-00-00-00-00-00@localhost",
							})
							*ctx = s.Context()
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
			expected: &target.Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00@localhost",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, RestoreTarget(ctx))
		})
	}
}

func TestRestoreAPI(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    internalclient.Client
	}{
		{
			description: "fails when api is not set",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
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
			expected: nil,
		},
		{
			description: "succeeds in restoring api",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(api, *new(internalclient.Client))
							*ctx = s.Context()
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
			expected: *new(internalclient.Client),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, RestoreAPI(ctx))
		})
	}
}

func TestRestoreLookup(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    map[string]string
	}{
		{
			description: "fails when lookup is not set",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
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
			expected: nil,
		},
		{
			description: "succeeds in restoring lookup",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(lookup, map[string]string{"foo": "bar"})
							*ctx = s.Context()
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
			expected: map[string]string{"foo": "bar"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, RestoreLookup(ctx))
		})
	}
}

func TestRestoreDevice(t *testing.T) {
	cases := []struct {
		description string
		setup       func(ctx *gliderssh.Context) *sshsrvtest.Conn
		expected    *models.Device
	}{
		{
			description: "fails when device is not set",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							*ctx = s.Context()
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
			expected: nil,
		},
		{
			description: "succeeds in restoring device",
			setup: func(ctx *gliderssh.Context) *sshsrvtest.Conn {
				return sshsrvtest.New(
					&gliderssh.Server{
						Handler: func(s gliderssh.Session) {
							s.Context().SetValue(
								device,
								&models.Device{
									CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
									Name:             "hostname",
									Identity:         &models.DeviceIdentity{MAC: "mac"},
									Info:             nil,
									PublicKey:        "",
									TenantID:         "00000000-0000-4000-0000-000000000000",
									Online:           true,
									Namespace:        "namespace",
									Status:           "accepted",
									RemoteAddr:       "",
									Position:         nil,
									Tags:             []string{"tag1"},
									PublicURL:        false,
									PublicURLAddress: "",
									Acceptable:       false,
								},
							)
							*ctx = s.Context()
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
			expected: &models.Device{
				CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				Name:             "hostname",
				Identity:         &models.DeviceIdentity{MAC: "mac"},
				Info:             nil,
				PublicKey:        "",
				TenantID:         "00000000-0000-4000-0000-000000000000",
				Online:           true,
				Namespace:        "namespace",
				Status:           "accepted",
				RemoteAddr:       "",
				Position:         nil,
				Tags:             []string{"tag1"},
				PublicURL:        false,
				PublicURLAddress: "",
				Acceptable:       false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			srv := tc.setup(&ctx)

			srv.Start()
			defer srv.Teardown()

			assert.NoError(t, srv.Agent.Run(""))
			assert.Equal(t, tc.expected, RestoreDevice(ctx))
		})
	}
}
