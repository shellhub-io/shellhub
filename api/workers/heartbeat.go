package workers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

func StartHeartBeat(_ context.Context, store store.Store) {
	envs, err := getEnvs()
	if err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskHeartbeat}).
			WithError(err).
			Error("Failed to parse the envs.")

		return
	}

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskHeartbeat}).
			WithError(err).
			Errorf("Failed to parse redis URI: %s.", envs.RedisURI)

		return
	}

	aggregate := func(group string, tasks []*asynq.Task) *asynq.Task {
		var b strings.Builder

		for _, task := range tasks {
			b.WriteString(fmt.Sprintf("%s:%d\n", task.Payload(), time.Now().Unix()))
		}

		return asynq.NewTask(TaskHeartbeat, []byte(b.String()))
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{
			GroupAggregator:  asynq.GroupAggregatorFunc(aggregate),
			GroupMaxDelay:    time.Duration(envs.AsynqGroupMaxDelay) * time.Second,
			GroupGracePeriod: time.Duration(envs.AsynqGroupGracePeriod) * time.Second,
			GroupMaxSize:     envs.AsynqGroupMaxSize,
			Queues:           map[string]int{"api": 6},
		},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskHeartbeat, func(ctx context.Context, task *asynq.Task) error {
		log.WithFields(
			log.Fields{
				"component": "worker",
				"task":      TaskHeartbeat,
			}).
			Info("Executing heartbeat worker.")

		scanner := bufio.NewScanner(bytes.NewReader(task.Payload()))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			parts := strings.SplitN(scanner.Text(), ":", 2)
			uid := parts[0]

			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log.WithFields(
					log.Fields{
						"component": "worker",
						"task":      TaskHeartbeat,
						"index":     rune(i),
					}).
					WithError(err).
					Warn("Failed to parse timestamp to integer.")

				continue
			}

			timestamp := time.Unix(i, 0)
			store.DeviceSetOnline(ctx, models.UID(uid), timestamp, true) //nolint:errcheck
		}

		log.WithFields(
			log.Fields{
				"component": "worker",
				"task":      TaskHeartbeat,
			}).
			Info("Finishing heartbeat worker.")

		return nil
	})

	if err := srv.Run(mux); err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskHeartbeat}).
			WithError(err).
			Fatal("Unable to run the server.")
	}
}
