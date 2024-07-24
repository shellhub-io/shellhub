package worker

// Server represents a process that handles tasks and cronjobs. A [Client] can submit
// tasks.
type Server interface {
	// HandleTask registers a task with the provided pattern. The task will be executed
	// every time a client enqueues a payload matching the pattern. Batch tasks will be executed
	// when the client enqueues a series of payloads or when the specified time delay is reached.
	//
	// It panics if the pattern is invalid. Batch tasks are specified in options.
	HandleTask(pattern TaskPattern, cb TaskHandler, opts ...TaskOption)
	// HandleCron executes the cronFunc every time the cron specification is met.
	//
	// It panics if the cron specification is invalid.
	HandleCron(spec CronSpec, cronFunc CronHandler)
	// Start initializes and starts the worker in a non-blocking manner. The server is
	// turned off whedn the context was done.
	//
	// It returns an error if any issues occur during the startup process.
	Start() error
	// Shutdown gracefully shuts down the server.
	Shutdown()
}
