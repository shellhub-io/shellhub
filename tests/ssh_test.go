package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func NewAgentContainer(ctx context.Context, port string, opts ...NewAgentContainerOption) (testcontainers.Container, error) {
	envs := map[string]string{
		"SHELLHUB_SERVER_ADDRESS":     fmt.Sprintf("http://localhost:%s", port),
		"SHELLHUB_TENANT_ID":          "00000000-0000-4000-0000-000000000000",
		"SHELLHUB_PRIVATE_KEY":        "/tmp/shellhub.key",
		"SHELLHUB_LOG_FORMAT":         "json",
		"SHELLHUB_KEEPALIVE_INTERVAL": "1",
		"SHELLHUB_LOG_LEVEL":          "trace",
	}

	for _, opt := range opts {
		opt(envs)
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:         envs,
			NetworkMode: "host",
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
		Logger: log.New(io.Discard, "", log.LstdFlags),
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
					assert.Equal(tt, http.StatusOK, resp.StatusCode())
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
					assert.Equal(tt, http.StatusOK, resp.StatusCode())
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				_, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
				require.Equal(t, http.StatusOK, resp.StatusCode())
				require.NoError(t, err)

				signer, err := ssh.NewSignerFromKey(privateKey)
				require.NoError(t, err)

				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				_, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.Error(t, err)
			},
		},
		/*{
			name: "connection keepalive when session is requested",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var globalConn ssh.Conn

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					dialed, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config.Timeout)
					assert.NoError(tt, err)

					conn, _, _, err := ssh.NewClientConn(dialed, fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
					assert.NoError(tt, err)

					globalConn = conn
				}, 30*time.Second, 1*time.Second)

				ch, reqs, err := globalConn.OpenChannel("session", nil)
				assert.NoError(t, err)

				ok, err := ch.SendRequest("shell", true, nil)
				assert.True(t, ok)
				assert.NoError(t, err)

				req := <-reqs
				assert.True(t, strings.HasPrefix(req.Type, "keepalive"))

				ch.Close()
				globalConn.Close()
			},
		}*/
		{
			name: "connection SHELL with Pty",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
			name: "connection EXEC and a SHELL on same connection",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password("password"),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				var conn *ssh.Client

				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					var err error

					conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
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
		{
			name:    "direct tcpip port redirect",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, env *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", env.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)

				type Data struct {
					DestAddr   string
					DestPort   uint32
					OriginAddr string
					OriginPort uint32
				}

				port := environment.GetFreePort(t)

				listener, err := net.Listen("tcp", ":"+port)
				require.NoError(t, err)

				wg := new(sync.WaitGroup)

				wg.Add(1)
				go func() {
					defer wg.Done()

					conn, err := listener.Accept()
					require.NoError(t, err)

					buffer := make([]byte, 1024)

					read, err := conn.Read(buffer)
					require.NoError(t, err)

					require.Equal(t, read, 4)
					require.Equal(t, "test", string(buffer[:4]))

					conn.Close()
				}()

				dest, err := strconv.Atoi(port)
				require.NoError(t, err)

				orig, err := strconv.Atoi(environment.GetFreePort(t))
				require.NoError(t, err)

				data := Data{
					DestAddr:   "0.0.0.0",
					DestPort:   uint32(dest), //nolint:gosec
					OriginAddr: "127.0.0.1",
					OriginPort: uint32(orig), //nolint:gosec
				}

				ch, _, err := conn.OpenChannel("direct-tcpip", ssh.Marshal(data))
				require.NoError(t, err)

				wrote, err := ch.Write([]byte("test"))
				require.NoError(t, err)

				require.Equal(t, wrote, 4)

				wg.Wait()

				ch.Close()
				conn.Close()
			},
		},
		{
			name: "session timeout behavior",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)

				sess, err := conn.NewSession()
				require.NoError(t, err)

				output, err := sess.CombinedOutput("sleep 3 && echo -n 'still alive'")
				require.NoError(t, err)

				assert.Equal(t, "still alive", string(output))

				sess.Close()
				conn.Close()
			},
		},
		{
			name:    "connection SFTP to upload large file",
			options: []NewAgentContainerOption{},
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				client, err := sftp.NewClient(conn)
				require.NoError(t, err)
				defer client.Close()

				// Create a large file (10MB)
				fileSize := 10 * 1024 * 1024 // 10MB
				randomData := make([]byte, fileSize)
				_, err = rand.Read(randomData)
				require.NoError(t, err)

				tempFile, err := os.CreateTemp("", "large-file-test-*.bin")
				require.NoError(t, err)
				defer os.Remove(tempFile.Name())

				_, err = tempFile.Write(randomData)
				require.NoError(t, err)
				tempFile.Close()

				localFile, err := os.Open(tempFile.Name())
				require.NoError(t, err)
				defer localFile.Close()

				remoteFile, err := client.Create("/tmp/large-file-test.bin")
				require.NoError(t, err)
				defer remoteFile.Close()

				written, err := io.Copy(remoteFile, localFile)
				require.NoError(t, err)
				assert.Equal(t, int64(fileSize), written)

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				output, err := sess.Output("stat -c %s /tmp/large-file-test.bin")
				require.NoError(t, err)

				size, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
				require.NoError(t, err)
				assert.Equal(t, int64(fileSize), size)
			},
		},
		{
			name: "connection EXEC with large output",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				// Generate large output (around 1MB).
				output, err := sess.Output("yes X | tr -d '\n' | head -c 1048576")
				require.NoError(t, err)

				assert.Equal(t, 1024*1024, len(output))
				for _, b := range output {
					assert.Equal(t, byte('X'), b)
				}
			},
		},
		{
			name: "connection EXEC with environment variables",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				err = sess.Setenv("TEST_VAR1", "test_value1")
				require.NoError(t, err)
				err = sess.Setenv("TEST_VAR2", "test_value2")
				require.NoError(t, err)

				output, err := sess.Output("echo -n $TEST_VAR1-$TEST_VAR2")
				require.NoError(t, err)

				assert.Equal(t, "test_value1-test_value2", string(output))
			},
		},
		{
			name: "terminal window size change",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				initialWidth, initialHeight := 80, 24
				err = sess.RequestPty("xterm", initialHeight, initialWidth, ssh.TerminalModes{
					ssh.ECHO: 1,
				})
				require.NoError(t, err)

				stdin, _ := sess.StdinPipe()
				stdout, _ := sess.StdoutPipe()

				err = sess.Shell()
				require.NoError(t, err)

				_, err = fmt.Fprintf(stdin, "stty size; echo DONE\n")
				require.NoError(t, err)

				scanner := bufio.NewScanner(stdout)
				var initialSizeOutput string
				for scanner.Scan() {
					line := scanner.Text()
					if line == "DONE" {
						break
					}
					initialSizeOutput = line
				}

				newWidth, newHeight := 120, 40
				err = sess.WindowChange(newHeight, newWidth)
				require.NoError(t, err)

				_, err = fmt.Fprintf(stdin, "stty size; echo DONE\n")
				require.NoError(t, err)

				var newSizeOutput string
				for scanner.Scan() {
					line := scanner.Text()
					if line == "DONE" {
						break
					}
					newSizeOutput = line
				}

				assert.Equal(t, fmt.Sprintf("%d %d", initialHeight, initialWidth), strings.TrimSpace(initialSizeOutput))
				assert.Equal(t, fmt.Sprintf("%d %d", newHeight, newWidth), strings.TrimSpace(newSizeOutput))
			},
		},
		{
			name: "connection EXEC with invalid command",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				_, err = sess.Output("this-command-does-not-exist")
				require.Error(t, err)

				var exitErr *ssh.ExitError
				require.ErrorAs(t, err, &exitErr)
				assert.NotEqual(t, 0, exitErr.ExitStatus())
			},
		},
		{
			name: "handling special characters in commands and output",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				specialChars := "!@#$%^&*()_+{}[]|;:,.<>/?`~"
				output, err := sess.Output(fmt.Sprintf("echo -n '%s'", specialChars))
				require.NoError(t, err)

				assert.Equal(t, specialChars, string(output))

				// Test Unicode characters
				unicodeChars := "こんにちは世界 ñáéíóú 你好世界"
				sess2, err := conn.NewSession()
				require.NoError(t, err)
				defer sess2.Close()

				output, err = sess2.Output(fmt.Sprintf("echo -n '%s'", unicodeChars))
				require.NoError(t, err)

				assert.Equal(t, unicodeChars, string(output))
			},
		},
		{
			name: "connection with cipher and MAC preferences",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
					Config: ssh.Config{
						Ciphers: []string{
							"aes256-ctr", "aes192-ctr", "aes128-ctr",
							"aes256-gcm@openssh.com", "aes128-gcm@openssh.com",
						},
						MACs: []string{
							"hmac-sha2-256-etm@openssh.com",
							"hmac-sha2-512-etm@openssh.com",
							"hmac-sha2-256",
							"hmac-sha2-512",
						},
					},
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				output, err := sess.Output("echo -n 'cipher test'")
				require.NoError(t, err)
				assert.Equal(t, "cipher test", string(output))
			},
		},
		{
			name: "multiple concurrent SSH sessions",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				const numConnections = 5
				var wg sync.WaitGroup
				errors := make(chan error, numConnections)

				for i := range numConnections {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
						if err != nil {
							errors <- fmt.Errorf("connection %d failed: %w", id, err)

							return
						}
						defer conn.Close()

						sess, err := conn.NewSession()
						if err != nil {
							errors <- fmt.Errorf("session %d failed: %w", id, err)

							return
						}
						defer sess.Close()

						expected := fmt.Sprintf("session-%d", id)
						output, err := sess.Output(fmt.Sprintf("echo -n '%s'", expected))
						if err != nil {
							errors <- fmt.Errorf("command %d failed: %w", id, err)

							return
						}

						if string(output) != expected {
							errors <- fmt.Errorf("unexpected output from session %d: got %q, want %q", id, string(output), expected)
						}
					}(i)
				}

				wg.Wait()
				close(errors)

				for err := range errors {
					require.NoError(t, err)
				}
			},
		},
		{
			name: "connection with strict host key checking simulation",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				var learnedKey ssh.PublicKey
				config1 := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
						learnedKey = key

						return nil
					},
				}

				conn1, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config1)
				require.NoError(t, err)
				conn1.Close()

				config2 := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
						if !bytes.Equal(key.Marshal(), learnedKey.Marshal()) {
							return fmt.Errorf("host key mismatch")
						}

						return nil
					},
				}

				conn2, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config2)
				require.NoError(t, err)
				defer conn2.Close()
			},
		},
		{
			name: "connection with keep-alive and heartbeat",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
					Timeout:         10 * time.Second,
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				go func() {
					ticker := time.NewTicker(2 * time.Second)
					defer ticker.Stop()
					for range 3 {
						<-ticker.C
						_, _, err := conn.SendRequest("keepalive@shellhub.io", true, nil)
						if err != nil {
							t.Logf("Keep-alive failed: %v", err)

							return
						}
					}
				}()

				time.Sleep(8 * time.Second)

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				output, err := sess.Output("echo -n 'alive after keepalive'")
				require.NoError(t, err)
				assert.Equal(t, "alive after keepalive", string(output))
			},
		},
		{
			name: "connection with subsystem request (sftp)",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				err = sess.RequestSubsystem("sftp")
				require.NoError(t, err)

				stdin, err := sess.StdinPipe()
				require.NoError(t, err)

				stdout, err := sess.StdoutPipe()
				require.NoError(t, err)

				initPacket := []byte{0, 0, 0, 5, 1, 0, 0, 0, 3} // SSH_FXP_INIT with version 3
				_, err = stdin.Write(initPacket)
				require.NoError(t, err)

				response := make([]byte, 9)
				n, err := stdout.Read(response)
				require.NoError(t, err)
				assert.Equal(t, 9, n)
				assert.Equal(t, byte(2), response[4]) // SSH_FXP_VERSION
			},
		},
		{
			name: "connection with pseudo-terminal modes",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				modes := ssh.TerminalModes{
					ssh.ECHO:          0,     // Disable echo
					ssh.TTY_OP_ISPEED: 14400, // Input speed
					ssh.TTY_OP_OSPEED: 14400, // Output speed
					ssh.ICRNL:         1,     // Map CR to NL on input
					ssh.OPOST:         1,     // Enable output processing
				}

				err = sess.RequestPty("xterm-256color", 24, 80, modes)
				require.NoError(t, err)

				stdin, err := sess.StdinPipe()
				require.NoError(t, err)

				stdout, err := sess.StdoutPipe()
				require.NoError(t, err)

				err = sess.Shell()
				require.NoError(t, err)

				_, err = stdin.Write([]byte("stty -echo && echo 'no echo test' && exit\n"))
				require.NoError(t, err)

				buffer := make([]byte, 1024)
				n, err := stdout.Read(buffer)
				require.NoError(t, err)
				assert.Greater(t, n, 0)
			},
		},
		{
			name: "connection with signal handling",
			run: func(t *testing.T, environment *Environment, device *models.Device) {
				config := &ssh.ClientConfig{
					User: fmt.Sprintf("%s@%s.%s", ShellHubAgentUsername, ShellHubNamespaceName, device.Name),
					Auth: []ssh.AuthMethod{
						ssh.Password(ShellHubAgentPassword),
					},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
				}

				conn, err := ssh.Dial("tcp", fmt.Sprintf("localhost:%s", environment.services.Env("SHELLHUB_SSH_PORT")), config)
				require.NoError(t, err)
				defer conn.Close()

				sess, err := conn.NewSession()
				require.NoError(t, err)
				defer sess.Close()

				err = sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{})
				require.NoError(t, err)

				stdin, err := sess.StdinPipe()
				require.NoError(t, err)

				err = sess.Shell()
				require.NoError(t, err)

				_, err = stdin.Write([]byte("sleep 30 &\n"))
				require.NoError(t, err)

				time.Sleep(100 * time.Millisecond)

				err = sess.Signal(ssh.SIGINT)
				if err != nil {
					t.Logf("Signal sending not supported: %v", err)
				}

				err = sess.Signal(ssh.SIGTERM)
				if err != nil {
					t.Logf("Signal sending not supported: %v", err)
				}

				_, err = stdin.Write([]byte("echo 'signal test done'\n"))
				require.NoError(t, err)
			},
		},
	}

	ctx := context.Background()

	compose := environment.New(t).Up(ctx)
	t.Cleanup(func() {
		compose.Down()
	})

	compose.NewUser(ctx, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(ctx, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace)

	auth := models.UserAuthResponse{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).
			SetBody(map[string]string{
				"username": ShellHubUsername,
				"password": ShellHubPassword,
			}).
			SetResult(&auth).
			Post("/api/login")
		assert.Equal(tt, http.StatusOK, resp.StatusCode())
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)

	// compose.R(ctx).SetAuthScheme("Bearer")
	// compose.R(ctx).SetAuthToken(auth.Token)

	compose.JWT(auth.Token)

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			agent, err := NewAgentContainer(
				ctx,
				compose.Env("SHELLHUB_HTTP_PORT"),
				test.options...,
			)
			require.NoError(t, err)

			err = agent.Start(ctx)
			require.NoError(t, err)

			t.Cleanup(func() {
				assert.NoError(t, agent.Terminate(ctx))
			})

			devices := []models.Device{}

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).SetResult(&devices).
					Get("/api/devices?status=pending")
				assert.Equal(tt, http.StatusOK, resp.StatusCode())
				assert.NoError(tt, err)

				assert.Len(tt, devices, 1)
			}, 30*time.Second, 1*time.Second)

			resp, err := compose.R(ctx).
				Patch(fmt.Sprintf("/api/devices/%s/accept", devices[0].UID))
			require.Equal(t, http.StatusOK, resp.StatusCode())
			require.NoError(t, err)

			device := models.Device{}

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).
					SetResult(&device).
					Get(fmt.Sprintf("/api/devices/%s", devices[0].UID))
				assert.Equal(tt, http.StatusOK, resp.StatusCode())
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
