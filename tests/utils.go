package main

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	testcontainers "github.com/testcontainers/testcontainers-go/modules/compose"
)

// getFreePort returns a randomly available TCP port.
func getFreePort() string {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

// loadEnv loads environment variables from the .env file and sets values for SHELLHUB_HTTP_PORT and SHELLHUB_SSH_PORT
func loadEnv(t *testing.T) map[string]string {
	env, err := godotenv.Read("../.env")
	assert.NoError(t, err)

	// Automatically assigns a random free port for HTTP and SSH services
	env["SHELLHUB_HTTP_PORT"] = getFreePort()
	env["SHELLHUB_SSH_PORT"] = getFreePort()

	return env
}

// withTestEnvironment sets up a test environment using Docker Compose.
// It takes a testing instance, environment variables, and a callback function for the core test logic as parameters.
func withTestEnvironment(t *testing.T, env map[string]string, cb func()) {
	compose, err := testcontainers.NewDockerCompose("../docker-compose.yml", "../docker-compose.dev.yml")
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := compose.Down(context.Background(), testcontainers.RemoveOrphans(true), testcontainers.RemoveImagesLocal)
		assert.NoError(t, err)
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err = compose.WithEnv(env).Up(ctx, testcontainers.Wait(true))
	assert.NoError(t, err)

	cb()
}
