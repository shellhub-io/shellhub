package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

var (
	ShellHubAgentUsername = "root"
	ShellHubAgentPassword = "password"
)

const (
	ShellHubUsername      = "test"
	ShellHubPassword      = "password"
	ShellHubNamespaceName = "testspace"
	ShellHubNamespace     = "00000000-0000-4000-0000-000000000000"
	ShellHubEmail         = "test@ossystems.com.br"
)

type NewAgentContainerOption func(envs map[string]string)

func NewAgentContainerWithIdentity(identity string) NewAgentContainerOption {
	return func(envs map[string]string) {
		envs["SHELLHUB_PREFERRED_IDENTITY"] = identity
	}
}

func NewAgentContainer(ctx context.Context, network string, opts ...NewAgentContainerOption) (testcontainers.Container, error) {
	envs := map[string]string{
		"SHELLHUB_SERVER_ADDRESS": "http://gateway:80",
		"SHELLHUB_TENANT_ID":      "00000000-0000-4000-0000-000000000000",
		"SHELLHUB_PRIVATE_KEY":    "/tmp/shellhub.key",
		"SHELLHUB_LOG_FORMAT":     "json",
	}

	for _, opt := range opts {
		opt(envs)
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:      envs,
			Networks: []string{network},
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       "..",
				Dockerfile:    "agent/Dockerfile.test",
				PrintBuildLog: false,
				KeepImage:     true,
				BuildArgs: map[string]*string{
					"USERNAME": &ShellHubAgentUsername,
					"PASSWORD": &ShellHubAgentPassword,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func TestSSH(t *testing.T) {
	type Environment struct {
		services *environment.DockerCompose
		agent    testcontainers.Container
	}

	tests := []struct {
		name    string
		options []NewAgentContainerOption
		run     func(*testing.T, *Environment, *models.Device)
	}{
		{
			name: "reconnect to server",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				ctx := context.Background()

				err := environment.agent.Stop(ctx, nil)
				require.NoError(t, err)

				err = environment.agent.Start(ctx)
				require.NoError(t, err)

				model := models.Device{}

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					resp, err := environment.services.R(ctx).
						SetResult(&model).
						Get(fmt.Sprintf("/api/devices/%s", device.UID))
					assert.Equal(tt, 200, resp.StatusCode())
					assert.NoError(tt, err)

					assert.True(tt, model.Online)
				}, 30*time.Second, 1*time.Second)
			},
		},
		{
			name: "reconnect to server with custom identity",
			options: []NewAgentContainerOption{
				NewAgentContainerWithIdentity("test"),
			},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				ctx := context.Background()

				err := environment.agent.Stop(ctx, nil)
				require.NoError(t, err)

				err = environment.agent.Start(ctx)
				require.NoError(t, err)

				model := models.Device{}

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					resp, err := environment.services.R(ctx).
						SetResult(&model).
						Get(fmt.Sprintf("/api/devices/%s", device.UID))
					assert.Equal(tt, 200, resp.StatusCode())
					assert.NoError(tt, err)

					assert.True(tt, model.Online)
				}, 30*time.Second, 1*time.Second)
			},
		},
		{
			name: "authenticate with password",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				conn.Close()
			},
		},
		{
			name: "fail to authenticate with password",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password("wrongpassword"),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				_, err := ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.Error(t, err)
			},
		},
		{
			name: "authenticate with password with custom identity",
			options: []NewAgentContainerOption{
				NewAgentContainerWithIdentity("test"),
			},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				conn.Close()
			},
		},
		{
			name: "authenticate with public key",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				ctx := context.Background()

				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
				require.NoError(t, err)

				model := requests.PublicKeyCreate{
					Name:     ShellHubAgentUsername,
					Username: ".*",
					Data:     ssh.MarshalAuthorizedKey(publicKey),
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				resp, err := environment.services.R(ctx).
					SetBody(&model).
					Post("/api/sshkeys/public-keys")
				require.Equal(t, 200, resp.StatusCode())
				require.NoError(t, err)

				signer, err := ssh.NewSignerFromKey(privateKey)
				require.NoError(t, err)

				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)

				conn.Close()
			},
		},
		{
			name: "fail to authenticate with public key",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)

				signer, err := ssh.NewSignerFromKey(privateKey)
				require.NoError(t, err)

				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				_, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.Error(t, err)
			},
		},
		{
			name: "connection SHELL with Pty",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				err = sess.RequestPty("xterm", 100, 100, ssh.TerminalModes{
					ssh.ECHO:          1,
					ssh.TTY_OP_ISPEED: 14400,
					ssh.TTY_OP_OSPEED: 14400,
				})
				require.NoError(t, err)

				err = sess.Shell()
				require.NoError(t, err)

				sess.Close()
				conn.Close()
			},
		},
		{
			name: "connection SHELL without pty",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				err = sess.Shell()
				require.NoError(t, err)

				sess.Close()
				conn.Close()
			},
		},
		{
			name: "connection SHELL without pty with output",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				stdin, _ := sess.StdinPipe()
				stdout, _ := sess.StdoutPipe()

				err = sess.Shell()
				require.NoError(t, err)

				var data string

				_, err = fmt.Fprintf(stdin, `echo -n "test"`)
				require.NoError(t, err)

				stdin.Close()

				_, err = fmt.Fscanf(stdout, "%s", &data)
				require.NoError(t, err)

				assert.Equal(t, data, "test")

				sess.Close()
				conn.Close()
			},
		},
		{
			name: "connection EXEC and a SHELL on same connection",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password("password"),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				{
					sess, err := conn.NewSession()
					require.NoError(t, err)

					output, err := sess.Output(`echo -n "test"`)
					require.NoError(t, err)

					assert.Equal(t, "test", string(output))

					sess.Close()
				}
				{
					sess, err := conn.NewSession()
					require.NoError(t, err)

					err = sess.RequestPty("xterm", 100, 100, ssh.TerminalModes{
						ssh.ECHO:          1,
						ssh.TTY_OP_ISPEED: 14400,
						ssh.TTY_OP_OSPEED: 14400,
					})
					require.NoError(t, err)

					err = sess.Shell()
					require.NoError(t, err)

					sess.Close()
				}

				conn.Close()
			},
		},
		{
			name: "connection EXEC",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password("password"),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				output, err := sess.Output(`echo -n "test"`)
				require.NoError(t, err)

				assert.Equal(t, "test", string(output))

				sess.Close()
				conn.Close()
			},
		},
		{
			name: "connection EXEC with non zero status code",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				var status *ssh.ExitError

				// NOTICE: write to stderr to simulate a error from connection.
				output, err := sess.CombinedOutput(`echo -n "test" 1>&2; exit 142`)
				require.ErrorAs(t, err, &status)

				assert.Equal(t, 142, status.ExitStatus())
				assert.Equal(t, "test", string(output))

				sess.Close()
				conn.Close()
			},
		},
		{
			name: "connection EXEC with custom identity",
			options: []NewAgentContainerOption{
				NewAgentContainerWithIdentity("test"),
			},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				output, err := sess.Output(`echo -n "test"`)
				require.NoError(t, err)

				assert.Equal(t, "test", string(output))

				sess.Close()
				conn.Close()
			},
		},
		{
			name:    "connection SFTP to upload file",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := sftp.NewClient(conn)
				require.NoError(t, err)

				sent, err := sess.OpenFile("/tmp/sent", (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
				require.NoError(t, err)

				wrote, err := fmt.Fprintf(sent, "sent file content")
				require.NoError(t, err)

				assert.Equal(t, 17, wrote)

				sess.Close()
				conn.Close()
			},
		},
		{
			name:    "connection SFTP to download file",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := sftp.NewClient(conn)
				require.NoError(t, err)

				received, err := sess.OpenFile("/etc/os-release", (os.O_RDONLY))
				require.NoError(t, err)

				var data string

				_, err = fmt.Fscanf(received, "%s", &data)
				require.NoError(t, err)

				// NOTICE: This assertion brake if the Docker image used to build the Agent wasn't the Alpine.
				assert.Contains(t, data, "Alpine")

				sess.Close()
				conn.Close()
			},
		},
		{
			name:    "connection SCP to upload file",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := scp.NewClientBySSH(conn)
				require.NoError(t, err)

				ctx := context.Background()

				file := bytes.NewBuffer(make([]byte, 1024))

				err = sess.CopyFilePassThru(ctx, file, "/tmp/sent", "0644", io.LimitReader)
				require.NoError(t, err)

				sess.Close()
				conn.Close()
			},
		},
		{
			name:    "connection SCP to download file",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("0.0.0.0:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)
				}, 30*time.Second, 1*time.Second)

				sess, err := scp.NewClientBySSH(conn)
				require.NoError(t, err)

				ctx := context.Background()

				file := bytes.NewBuffer(make([]byte, 1024))

				err = sess.CopyFromRemotePassThru(ctx, file, "/etc/os-release", nil)
				require.NoError(t, err)

				sess.Close()
				conn.Close()
			},
		},
	}

	mu := new(sync.Mutex)

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			mu.Lock()
			compose := environment.New(t).Up(ctx)
			mu.Unlock()

			t.Cleanup(func() {
				compose.Down()
			})

			agent, err := NewAgentContainer(
				ctx,
				compose.Env("SHELLHUB_NETWORK"),
				test.options...,
			)
			require.NoError(t, err)

			compose.NewUser(ctx, ShellHubUsername, ShellHubEmail, ShellHubPassword)
			compose.NewNamespace(ctx, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace)

			err = agent.Start(ctx)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, agent.Terminate(ctx))
			})

			auth := models.UserAuthResponse{}

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).
					SetBody(map[string]string{
						"username": ShellHubUsername,
						"password": ShellHubPassword,
					}).
					SetResult(&auth).
					Post("/api/login")
				assert.Equal(tt, 200, resp.StatusCode())
				assert.NoError(tt, err)
			}, 30*time.Second, 1*time.Second)

			// compose.R(ctx).SetAuthScheme("Bearer")
			// compose.R(ctx).SetAuthToken(auth.Token)

			compose.JWT(auth.Token)

			devices := []models.Device{}

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).SetResult(&devices).
					Get("/api/devices?status=pending")
				assert.Equal(tt, 200, resp.StatusCode())
				assert.NoError(tt, err)

				assert.Len(tt, devices, 1)
			}, 30*time.Second, 1*time.Second)

			resp, err := compose.R(ctx).
				Patch(fmt.Sprintf("/api/devices/%s/accept", devices[0].UID))
			require.Equal(t, 200, resp.StatusCode())
			require.NoError(t, err)

			device := models.Device{}

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).
					SetResult(&device).
					Get(fmt.Sprintf("/api/devices/%s", devices[0].UID))
				assert.Equal(tt, 200, resp.StatusCode())
				assert.NoError(tt, err)

				assert.True(tt, device.Online)
			}, 30*time.Second, 1*time.Second)

			// --

			test.run(t, &Environment{
				services: compose,
				agent:    agent,
			}, &device)
		})
	}
}
