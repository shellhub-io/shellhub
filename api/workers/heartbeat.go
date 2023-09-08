package workers

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type HeartBeat struct {
	store store.Store
}

var _ Worker = (*HeartBeat)(nil)

func NewHeartbeater(store store.Store) *HeartBeat {
	return &HeartBeat{
		store: store,
	}
}

func (h *HeartBeat) Start(_ context.Context, msgs chan WorkerMessage) {
	envs, err := getEnvs()
	if err != nil {
		msgs <- NewWorkerMessage("failed to get the envs", err)

		return
	}

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		msgs <- NewWorkerMessage("failed to parse redis uri", err)

		return
	}

	aggregate := func(group string, tasks []*asynq.Task) *asynq.Task {
		var b strings.Builder

		for _, task := range tasks {
			b.Write(task.Payload())
			b.WriteString("\n")
		}

		return asynq.NewTask("api:heartbeat", []byte(b.String()))
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

	mux.HandleFunc("api:heartbeat", func(ctx context.Context, task *asynq.Task) error {
		scanner := bufio.NewScanner(bytes.NewReader(task.Payload()))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			h.store.DeviceSetOnline(ctx, models.UID(scanner.Text()), true) //nolint:errcheck
		}

		return nil
	})

	msgs <- WorkerMessageStarted

	if err := srv.Run(mux); err != nil {
		msgs <- WorkerMessageStopped
	}
}
