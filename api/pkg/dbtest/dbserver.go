package dbtest

import (
	"context"
	"fmt"

	"github.com/shellhub-io/mongotest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

// WARN: copy of https://github.com/testcontainers/testcontainers-go/pull/2469. should be removed
// if merged.
func withReplicaSet() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Cmd = append(req.Cmd, "--replSet", "rs")
		req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
			PostReadies: []testcontainers.ContainerHook{
				func(ctx context.Context, c testcontainers.Container) error {
					cIP, err := c.ContainerIP(ctx)
					if err != nil {
						return err
					}

					cmd := []string{
						"/bin/mongo",
						"--eval",
						fmt.Sprintf("rs.initiate({ _id: 'rs', members: [ { _id: 0, host: '%s:27017' } ] })", cIP),
					}

					if exitCode, _, err := c.Exec(ctx, cmd); err != nil || exitCode != 0 {
						return fmt.Errorf("fails to initiate replica set with status %d: %s", exitCode, err)
					}

					return nil
				},
			},
		})
	}
}

// DB represents a MongoDB test server instance.
type DB struct {
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

func (db *DB) configure(ctx context.Context) error {
	ports, err := db.tContainer.Ports(ctx)
	if err != nil {
		return err
	}
	// Index 0 is the IPV4 addr
	db.Container.ExposedPort = ports["27017/tcp"][0].HostPort

	cIP, err := db.tContainer.ContainerIP(ctx)
	if err != nil {
		return err
	}
	db.Container.ConnectionString = "mongodb://" + cIP + ":27017"

	if db.Container.Database == "" {
		db.Container.Database = "test"
	}

	return nil
}

// Up starts a new MongoDB container, configures the database to receive fixtures,
// and returns a DBServer instance.
func (db *DB) Up(ctx context.Context) error {
	var err error

	db.tContainer, err = mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:4.4.8"), withReplicaSet())
	if err != nil {
		return err
	}

	if err := db.configure(ctx); err != nil {
		return err
	}

	mongotest.Configure(mongotest.Config{
		URL:            db.Container.ConnectionString,
		Database:       db.Container.Database,
		FixtureRootDir: db.Fixtures.Root,
		PreInsertFuncs: db.Fixtures.PreInsertFuncs,
		FixtureFormat:  mongotest.FixtureFormatJSON,
	})

	return nil
}

// Down gracefully terminates the MongoDB container.
func (db *DB) Down(ctx context.Context) error {
	return db.tContainer.Terminate(ctx)
}

// Apply applies specified fixtures to the database.
func (*DB) Apply(fixtures ...string) error {
	return mongotest.UseFixture(fixtures...)
}

// Reset resets the entire database, removing all data.
func (*DB) Reset() error {
	return mongotest.DropDatabase()
}
