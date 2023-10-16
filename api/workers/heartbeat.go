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

func StartHeartBeat(ctx context.Context, store store.Store) error {
	envs, err := getEnvs()
	if err != nil {
		return fmt.Errorf("failed to get the envs: %w", err)
	}

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		return fmt.Errorf("failed to parse redis uri: %w", err)
	}

	aggregate := func(group string, tasks []*asynq.Task) *asynq.Task {
		var b strings.Builder

		for _, task := range tasks {
			b.WriteString(fmt.Sprintf("%s:%d\n", task.Payload(), time.Now().Unix()))
		}

		return asynq.NewTask("api:heartbeat", []byte(b.String()))
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{
			BaseContext: func() context.Context {
				return ctx
			},
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
			parts := strings.SplitN(scanner.Text(), ":", 2)
			uid := parts[0]

			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log.Error(err)

				continue
			}

			timestamp := time.Unix(i, 0)

			store.DeviceSetOnline(ctx, models.UID(uid), timestamp, true) //nolint:errcheck
		}

		return nil
	})

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}

	return nil
}
