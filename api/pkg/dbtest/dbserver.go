package dbtest

import (
	"context"

	"github.com/shellhub-io/mongotest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

// Server represents a MongoDB test server instance.
type Server struct {
	tContainer *mongodb.MongoDBContainer // Container is the MongoDB container instance.

	Container struct {
		ConnectionString string
		ExposedPort      string
		Database         string
	}

	Fixtures struct {
		Root           string                    // Root is the absolute path to seek fixture files.
		PreInsertFuncs []mongotest.PreInsertFunc // PreInsertFuncs is a list of functions to run before inserting data.
	}
}

func (srv *Server) configure(ctx context.Context) error {
	ports, err := srv.tContainer.Ports(ctx)
	if err != nil {
		return err
	}
	// Index 0 is the IPV4 addr
	srv.Container.ExposedPort = ports["27017/tcp"][0].HostPort

	cIP, err := srv.tContainer.ContainerIP(ctx)
	if err != nil {
		return err
	}
	srv.Container.ConnectionString = "mongodb://" + cIP + ":27017"

	if srv.Container.Database == "" {
		srv.Container.Database = "test"
	}

	return nil
}

// Up starts a new MongoDB container, configures the database to receive fixtures,
// and returns a DBServer instance.
func (srv *Server) Up(ctx context.Context) error {
	var err error

	srv.tContainer, err = mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:4.4.8"), mongodb.WithReplicaSet())
	if err != nil {
		return err
	}

	if err := srv.configure(ctx); err != nil {
		return err
	}

	mongotest.Configure(mongotest.Config{
		URL:            srv.Container.ConnectionString,
		Database:       srv.Container.Database,
		FixtureRootDir: srv.Fixtures.Root,
		PreInsertFuncs: srv.Fixtures.PreInsertFuncs,
		FixtureFormat:  mongotest.FixtureFormatJSON,
	})

	return nil
}

// Down gracefully terminates the MongoDB container.
func (srv *Server) Down(ctx context.Context) error {
	return srv.tContainer.Terminate(ctx)
}

// Apply applies specified fixtures to the database.
func (*Server) Apply(fixtures ...string) error {
	return mongotest.UseFixture(fixtures...)
}

// Reset resets the entire database, removing all data.
func (*Server) Reset() error {
	return mongotest.DropDatabase()
}
