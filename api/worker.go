package main

import (
	"context"
	"net/url"
	"runtime"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type envs struct {
	MongoURI               string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	SessionRecordRetention int    `envconfig:"record_retention" default:"0"`
}

// Connect connects to MongoDB.
func Connect(ctx context.Context, uri string) (*mongo.Database, error) {
	connStr, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, errors.Wrap(err, "invalid Mongo URI format")
	}

	// Applying MongoDB URI to client options.
	clientOptions := options.Client().ApplyURI(uri)

	// Connecting to MongoDB.
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MongoDB")
	}

	// Testing if MongoDB is connected.
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping MongoDB")
	}

	return client.Database(connStr.Database), nil
}

/*
	This worker will delete all data inside recorded_session's collection older than a date limit and set the "recorded"
	status from session's collection to false.
*/

// Delete deletes registers from recorded_session's collection.
func Delete(ctx context.Context, db *mongo.Database, limit time.Time) (int64, error) {
	deleted, err := db.Collection("recorded_sessions").DeleteMany(ctx,
		bson.M{"time": bson.D{{"$lte", limit}}},
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete the session's records from MongoDB")
	}

	return deleted.DeletedCount, nil
}

// Update updates session's records that were deleted to status no recorded.
func Update(ctx context.Context, db *mongo.Database, limit time.Time) (int64, error) {
	updated, err := db.Collection("sessions").UpdateMany(ctx,
		bson.M{"started_at": bson.D{{"$lte", limit}}, "recorded": bson.M{"$eq": true}},
		bson.M{"$set": bson.M{"recorded": false}})
	if err != nil {
		return 0, errors.Wrap(err, "failed to set the sessions from MongoDB to no recorded")
	}

	return updated.ModifiedCount, nil
}

// sessionRecordCleanup deletes session's records registers older than days defined by SHELLHUB_RECORD_RETENTION.
//
// If something inside the function does not work properly, it should panic.
// When SHELLHUB_RECORD_RETENTION is equals to zero, records will never be deleted.
// When SHELLHUB_RECORD_RETENTION is less than zero, nothing happen.
func sessionRecordCleanup() error {
	logrus.Info("running worker to delete session's records")

	ctx := context.Background()

	var env envs
	if err := envconfig.Process("api", &env); err != nil {
		return errors.Wrap(err, "failed to load environment variables")
	}

	// Session record retention time was not defined.
	if env.SessionRecordRetention == 0 {
		logrus.Warn("a time to clean the session's record was not defined. Skipping")

		return nil
	}

	if env.SessionRecordRetention < 0 {
		return errors.New("invalid time interval")
	}

	// Session's record older than that date will be deleted.
	date := time.Now().UTC().AddDate(0, 0, env.SessionRecordRetention*-1)

	logrus.Trace("connecting to MongoDB")

	database, err := Connect(ctx, env.MongoURI)
	if err != nil {
		logrus.WithError(err).Error("failed to connect to MongoDB")

		return err
	}

	deleted, err := Delete(ctx, database, date)
	if err != nil {
		logrus.WithError(err).Error("failed to delete the records on database")

		return err
	}

	logrus.Trace("session's record deleted")

	updated, err := Update(ctx, database, date)
	if err != nil {
		return err
	}

	logrus.Trace("session's record updated to no recorded")

	logrus.Info(deleted, " session's records deleted")
	logrus.Info(updated, " sessions set to no recorded")

	logrus.Info("closing worker to delete session's records")

	return nil
}

func startWorker(cfg *config) error {
	addr, err := url.Parse(cfg.RedisURI)
	if err != nil {
		return err
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: addr.Host},
		asynq.Config{
			Concurrency: runtime.NumCPU(),
		},
	)

	mux := asynq.NewServeMux()

	// Handle session_record:cleanup task
	mux.HandleFunc("session_record:cleanup", func(ctx context.Context, task *asynq.Task) error {
		if err := sessionRecordCleanup(); err != nil {
			logrus.Error(err)
		}

		return nil
	})

	go func() {
		if err := srv.Run(mux); err != nil {
			logrus.Fatal(err)
		}
	}()

	scheduler := asynq.NewScheduler(asynq.RedisClientOpt{Addr: addr.Host}, nil)

	// Schedule session_record:cleanup to run once a day
	if _, err := scheduler.Register(cfg.SessionRecordCleanupSchedule,
		asynq.NewTask("session_record:cleanup", nil, asynq.TaskID("session_record:cleanup"))); err != nil {
		logrus.Error(err)
	}

	return scheduler.Run()
}
