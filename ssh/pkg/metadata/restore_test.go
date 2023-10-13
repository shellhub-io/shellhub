package metadata

import (
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/sshtest"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
)

func TestRestore(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "succeeds",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(request, "exec")
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, restore(ctx, request))
		})
	}
}

func TestRestoreRequest(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "fails when request is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: "",
		},
		{
			description: "succeeds in restoring request type",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(request, "exec")
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestoreRequest(ctx))
		})
	}
}

func TestRestoreAuthenticationMethod(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    AuthMethod
	}{
		{
			description: "fails when authentication method is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: AuthMethodInvalid,
		},
		{
			description: "succeeds in restoring authentication method",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(authentication, AuthMethodPasswd)
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestoreAuthenticationMethod(ctx))
		})
	}
}

func TestRestorePassword(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "fails when password is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: "",
		},
		{
			description: "succeeds in restoring password",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(password, "secret")
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestorePassword(ctx))
		})
	}
}

func TestRestoreFingerprint(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    string
	}{
		{
			description: "fails when fingerprint is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: "",
		},
		{
			description: "succeeds in restoring fingerprint",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(fingerprint, "fingerprint")
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestoreFingerprint(ctx))
		})
	}
}

func TestRestoreTarget(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    *target.Target
	}{
		{
			description: "fails when target is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: nil,
		},
		{
			description: "succeeds in restoring target",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(tag, &target.Target{
							Username: "username",
							Data:     "namespace.00-00-00-00-00-00@localhost",
						})
						*ctx = s.Context()
					},
				})
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

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, RestoreTarget(ctx))
		})
	}
}

// TODO
// func TestRestoreAPI(t *testing.T) {
// }

func TestRestoreLookup(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    map[string]string
	}{
		{
			description: "fails when lookup is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: nil,
		},
		{
			description: "succeeds in restoring lookup",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(lookup, map[string]string{"foo": "bar"})
						*ctx = s.Context()
					},
				})
			},
			expected: map[string]string{"foo": "bar"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var ctx gliderssh.Context

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, RestoreLookup(ctx))
		})
	}
}

func TestRestoreDevice(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    *models.Device
	}{
		{
			description: "fails when device is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: nil,
		},
		{
			description: "succeeds in restoring device",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(device, &models.Device{
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
						})
						*ctx = s.Context()
					},
				})
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

			ssh := tc.setup(t, &ctx)
			defer ssh.Teardown()

			assert.NoError(t, ssh.Session.Run(""))
			assert.Equal(t, tc.expected, RestoreDevice(ctx))
		})
	}
}

// TODO
// func TestRestoreAgent(t *testing.T) {
// }

func TestRestoreEstablished(t *testing.T) {
	cases := []struct {
		description string
		setup       func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest
		expected    bool
	}{
		{
			description: "fails when request is not set",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						*ctx = s.Context()
					},
				})
			},
			expected: false,
		},
		{
			description: "succeeds in restoring request type",
			setup: func(t *testing.T, ctx *gliderssh.Context) *sshtest.SSHTest {
				return sshtest.Start(t, nil, &gliderssh.Server{
					Handler: func(s gliderssh.Session) {
						s.Context().SetValue(established, true)
						*ctx = s.Context()
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
			assert.Equal(t, tc.expected, RestoreEstablished(ctx))
		})
	}
}
