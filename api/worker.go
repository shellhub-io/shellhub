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

// sessionRecordCleanup deletes session's records registers older than days defined by SHELLHUB_RECORD_RETENTION.
// When SHELLHUB_RECORD_RETENTION is equals to zero, records will never be deleted.
// When SHELLHUB_RECORD_RETENTION is less than zero, nothing happen.
//
// If something inside the function does not work properly, it should panic.
func sessionRecordCleanup() error {
	logrus.Info("Running worker to delete session's records...")

	type config struct {
		MongoURI               string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
		SessionRecordRetention int    `envconfig:"record_retention" default:"0"`
	}

	// Loading env variables.
	var envs config
	if err := envconfig.Process("api", &envs); err != nil {
		return errors.Wrap(err, "Failed to load environment variables")
	}

	// Session record retention time was not defined.
	if envs.SessionRecordRetention == 0 {
		logrus.Warn("A time to clean the session's record was not defined. Skipping...")

		return nil
	}

	if envs.SessionRecordRetention < 0 {
		return errors.New("Invalid time interval")
	}

	// Registers older than that date will be deleted.
	dateLimit := time.Now().UTC().AddDate(0, 0, envs.SessionRecordRetention*-1)

	logrus.Debug("Connecting to MongoDB...")

	connStr, err := connstring.ParseAndValidate(envs.MongoURI)
	if err != nil {
		return errors.Wrap(err, "Invalid Mongo URI format")
	}

	// Applying MongoDB URI to client options.
	clientOptions := options.Client().ApplyURI(envs.MongoURI)
	// Connecting to MongoDB.
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return errors.Wrap(err, "Failed to connect to MongoDB")
	}

	logrus.Debug("Connected! Pinging...")

	// Testing if MongoDB is connected.
	if err = client.Ping(context.TODO(), nil); err != nil {
		return errors.Wrap(err, "Failed to ping MongoDB")
	}

	logrus.Debug("Pinged! Deleting session's record data...")

	db := client.Database(connStr.Database)

	/*
		This worker will delete all data inside recorded_session's collection older than a date limit and set the "recorded"
		status from session's collection to false.
	*/
	// Deleting registers from recorded_session's collection.
	deleted, err := db.Collection("recorded_sessions").DeleteMany(context.Background(),
		bson.M{"time": bson.D{{"$lte", dateLimit}}},
	)
	if err != nil {
		return errors.Wrap(err, "Failed to delete the session's records from MongoDB")
	}

	logrus.Debug("Deleted! Updating the record status from sessions...")
	// Setting session records that were deleted to correct status: no recorded.
	updated, err := db.Collection("sessions").UpdateMany(context.Background(),
		bson.M{"started_at": bson.D{{"$lte", dateLimit}}, "recorded": bson.M{"$eq": true}},
		bson.M{"$set": bson.M{"recorded": false}})
	if err != nil {
		return errors.Wrap(err, "Failed to set the sessions from MongoDB to no recorded")
	}

	logrus.Info(deleted.DeletedCount, " session's records deleted")
	logrus.Info(updated.ModifiedCount, " sessions set to no recorded")

	logrus.Info("Closing worker to delete session's records...")

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
