// Package workers manages the API's workers, setting up a new `asynq.Server` and binding it to the Redis instance.
//
// The core of this package is the `Worker` struct, created with the `New` function. Each
// instance of this struct includes its own server and scheduler. The `Start` method is
// responsible for setting up the necessary handlers and running the server.
//
// The `sessionCleanup` worker is designed to delete recorded sessions older than a specified number
// of days. The retention period is determined by the value of the `SHELLHUB_RECORD_RETENTION` environment
// variable. To disable this worker, set `SHELLHUB_RECORD_RETENTION` to 0 (default behavior). It uses
// a cron expression from `SHELLHUB_RECORD_RETENTION` to schedule its periodic execution.
//
// The patterns of tasks used by the handlers are available as constants with the "Task" prefix.
package workers
