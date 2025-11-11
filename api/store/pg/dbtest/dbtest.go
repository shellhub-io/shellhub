package dbtest

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Server represents a Postgres test server instance.
type Server struct {
	container *postgres.PostgresContainer
}

// Up starts a new Postgres container. Use [Server.ConnectionString] to access the connection string.
func (srv *Server) Up(ctx context.Context) error {
	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second)),
	}

	container, err := postgres.Run(ctx, "postgres:18.0", opts...)
	if err != nil {
		return err
	}

	srv.container = container

	return nil
}

// Down gracefully terminates the Postgres container.
func (srv *Server) Down(ctx context.Context) error {
	return srv.container.Terminate(ctx)
}

func (srv *Server) ConnectionString(ctx context.Context) (string, error) {
	host, err := srv.container.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := srv.container.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}

	return "postgres://postgres:postgres@" + host + ":" + port.Port() + "/test?sslmode=disable", nil
}
