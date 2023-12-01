package workers

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store"
	log "github.com/sirupsen/logrus"
)

type Workers struct {
	store store.Store

	addr      asynq.RedisConnOpt
	srv       *asynq.Server
	mux       *asynq.ServeMux
	env       *Envs
	scheduler *asynq.Scheduler
}

// New creates a new Workers instance with the provided store. It initializes
// the worker's components, such as server, scheduler, and environment settings.
func New(store store.Store) (*Workers, error) {
	env, err := getEnvs()
	if err != nil {
		log.WithFields(log.Fields{"component": "worker"}).
			WithError(err).
			Error("Failed to parse the envs.")

		return nil, err
	}

	addr, err := asynq.ParseRedisURI(env.RedisURI)
	if err != nil {
		log.WithFields(log.Fields{"component": "worker"}).
			WithError(err).
			Errorf("Failed to parse redis URI: %s.", env.RedisURI)

		return nil, err
	}

	mux := asynq.NewServeMux()
	srv := asynq.NewServer(
		addr,
		asynq.Config{ //nolint:exhaustruct
			// NOTICE:
			// To include any new task binding to a new queue (e.g., "queue:group" where 'queue' is the new queue),
			// ensure that the created queue is added here. Failure to do so will result in the server not executing the task handler.
			Queues: map[string]int{
				"api":            1,
				"session_record": 1,
			},
			GroupAggregator: asynq.GroupAggregatorFunc(
				func(group string, tasks []*asynq.Task) *asynq.Task {
					var b strings.Builder

					for _, task := range tasks {
						b.WriteString(fmt.Sprintf("%s:%d\n", task.Payload(), time.Now().Unix()))
					}

					return asynq.NewTask(TaskHeartbeat, []byte(b.String()))
				},
			),
			GroupMaxDelay:    time.Duration(env.AsynqGroupMaxDelay) * time.Second,
			GroupGracePeriod: time.Duration(env.AsynqGroupGracePeriod) * time.Second,
			GroupMaxSize:     env.AsynqGroupMaxSize,
			Concurrency:      runtime.NumCPU(),
		},
	)
	scheduler := asynq.NewScheduler(addr, nil)

	w := &Workers{
		addr:      addr,
		env:       env,
		srv:       srv,
		mux:       mux,
		scheduler: scheduler,
		store:     store,
	}

	return w, nil
}

// Start initiates the server. It creates two new goroutines: one for the server itself
// and another for the scheduler. This method is also responsible for setting up all
// the server handlers.
func (w *Workers) Start() {
	log.WithFields(log.Fields{"component": "worker"}).Info("Starting workers")

	w.setupHandlers()

	go func() {
		if err := w.srv.Run(w.mux); err != nil {
			log.WithFields(log.Fields{"component": "worker"}).
				WithError(err).
				Error("Unable to run the server.")
		}
	}()

	go func() {
		if err := w.scheduler.Run(); err != nil {
			log.WithFields(log.Fields{"component": "worker"}).
				WithError(err).
				Error("Unable to run the scheduler.")
		}
	}()
}

// setupHandlers is responsible for registering all the handlers of the server. It needs
// to be called before any initialization.
func (w *Workers) setupHandlers() {
	w.registerSessionCleanup()
	w.registerHeartbeat()
}
