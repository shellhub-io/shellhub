package main

import (
	"context"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// workerDeleteSessionRecord deletes session's records registers older than days defined by SHELLHUB_RECORD_RETENTION.
// When SHELLHUB_RECORD_RETENTION is equals to zero, records will never be deleted.
// When SHELLHUB_RECORD_RETENTION is less than zero, nothing happen.
//
// If something inside the function does not work properly, it should panic.
func workerDeleteSessionRecord() {
	logrus.Info("Running worker to delete session's records...")

	type config struct {
		MongoURI               string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
		SessionRecordRetention int    `envconfig:"record_retention" default:"0"`
	}

	// Loading env variables.
	var envs config
	if err := envconfig.Process("api", &envs); err != nil {
		logrus.WithError(err).Fatal("Failed to load environment variables")
	}

	// Session record retention time was not defined.
	if envs.SessionRecordRetention == 0 {
		logrus.Info("A time to clean the session's record was not defined. Skipping...")

		return
	}

	if envs.SessionRecordRetention < 0 {
		logrus.Fatal("Invalid time interval")
	}

	// Registers older than that date will be deleted.
	dateLimit := time.Now().UTC().AddDate(0, 0, envs.SessionRecordRetention*-1)

	logrus.Debug("Connecting to MongoDB...")

	// Applying MongoDB URI to client options.
	clientOptions := options.Client().ApplyURI(envs.MongoURI)
	// Connecting to MongoDB.
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	logrus.Debug("Connected! Pinging...")

	// Testing if MongoDB is connected.
	if err = client.Ping(context.TODO(), nil); err != nil {
		logrus.WithError(err).Fatal("Failed to ping MongoDB")
	}

	logrus.Debug("Pinged! Deleting session's record data...")

	/*
		This worker will delete all data inside recorded_session's collection older than a date limit and set the "recorded"
		status from session's collection to false.
	*/
	// Deleting registers from recorded_session's collection.
	deleted, err := client.Database("main").Collection("recorded_sessions").DeleteMany(context.Background(),
		bson.M{"time": bson.D{{"$lte", dateLimit}}},
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to delete the session's records from MongoDB")
	}

	logrus.Debug("Deleted! Updating the record status from sessions...")
	// Setting session records that were deleted to correct status: no recorded.
	updated, err := client.Database("main").Collection("sessions").UpdateMany(context.Background(),
		bson.M{"started_at": bson.D{{"$lte", dateLimit}}, "recorded": bson.M{"$eq": true}},
		bson.M{"$set": bson.M{"recorded": false}})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to set the sessions from MongoDB to no recorded")
	}

	logrus.Info(deleted.DeletedCount, " session's records deleted")
	logrus.Info(updated.ModifiedCount, " sessions set to no recorded")

	logrus.Info("Closing worker to delete session's records...")
}

var workerCmd = &cobra.Command{
	Use: "worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Info("Initializing workers...")
		// Running worker to clean the session recorded from a defined time interval.
		workerDeleteSessionRecord()
		logrus.Info("Successfully ran all workers!")

		return nil
	},
}
