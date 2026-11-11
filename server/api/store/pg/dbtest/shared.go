package dbtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
)

var shared struct {
	mu      sync.Mutex
	running bool
	started bool
	srv     *Server
	err     error
}

var databases atomic.Int64

// Run runs m's tests with one Postgres server that every [Shared] call in the package returns, and
// terminates that server once they finish. Call it from TestMain in place of m.Run and pass what it
// returns to os.Exit. A package whose tests never call [Shared] starts no container. When the server
// does not terminate, it reports the error on stderr and returns a failing code.
func Run(m *testing.M) int {
	shared.mu.Lock()
	shared.running = true
	shared.mu.Unlock()

	code := m.Run()

	shared.mu.Lock()
	defer shared.mu.Unlock()

	if shared.srv != nil {
		if err := shared.srv.Down(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "dbtest: terminating the shared Postgres server:", err)

			return max(code, 1)
		}
	}

	return code
}

// Shared returns the Postgres server the package's tests share, starting it on the first call. It
// fails when the package's TestMain does not go through [Run], which is what terminates the server,
// and when the container does not start. A failed start is not retried: every later call reports the
// same error.
func Shared(ctx context.Context) (*Server, error) {
	shared.mu.Lock()
	defer shared.mu.Unlock()

	if !shared.running {
		return nil, errors.New("dbtest.Shared needs the package's TestMain to call dbtest.Run")
	}

	if !shared.started {
		shared.started = true

		srv := &Server{}
		if err := srv.Up(ctx); err != nil {
			shared.err = err
		} else {
			shared.srv = srv
		}
	}

	return shared.srv, shared.err
}

// DatabaseURL returns the DSN of the database name on the server, without checking that the
// database exists. It fails when the container's host or mapped port 5432 cannot be read.
func (srv *Server) DatabaseURL(ctx context.Context, name string) (string, error) {
	host, err := srv.container.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := srv.container.MappedPort(ctx, "5432")
	if err != nil {
		return "", err
	}

	return "postgres://postgres:postgres@" + host + ":" + port.Port() + "/" + name + "?sslmode=disable", nil
}

// CreateDatabase creates an empty database, or a copy of template when it is not empty, and returns
// its name, unique on the server. Copying fails while anyone is connected to template, so seal a
// template with [Server.SealTemplate] before its first copy. It also fails when template does not
// exist, when [Server.DatabaseURL] fails, and when the server refuses the connection or the
// statement.
func (srv *Server) CreateDatabase(ctx context.Context, template string) (string, error) {
	name := fmt.Sprintf("db%d", databases.Add(1))

	statement := "CREATE DATABASE " + pgx.Identifier{name}.Sanitize()
	if template != "" {
		statement += " TEMPLATE " + pgx.Identifier{template}.Sanitize()
	}

	if err := srv.admin(ctx, statement); err != nil {
		return "", err
	}

	return name, nil
}

// SealTemplate refuses new connections to the database name and ends the ones it has, so
// [Server.CreateDatabase] can copy it. Sessions it ends report a terminated connection to whoever
// still holds them. It fails when name does not exist, when [Server.DatabaseURL] fails, and when
// the server refuses the connection or either statement; a failure after the first statement
// leaves the database refusing connections.
func (srv *Server) SealTemplate(ctx context.Context, name string) error {
	if err := srv.admin(ctx, "ALTER DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH ALLOW_CONNECTIONS false"); err != nil {
		return err
	}

	return srv.admin(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", name)
}

// DropDatabase drops the database name, ending the sessions still connected to it, and succeeds
// when name does not exist. It fails when [Server.DatabaseURL] fails and when the server refuses the
// connection or the statement.
func (srv *Server) DropDatabase(ctx context.Context, name string) error {
	return srv.admin(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
}

func (srv *Server) admin(ctx context.Context, statement string, args ...any) error {
	dsn, err := srv.DatabaseURL(ctx, "postgres")
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, statement, args...)

	return errors.Join(err, db.Close())
}
