package asynq

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

type client struct {
	asynqClient *asynq.Client
}

func NewClient(redisURI string) (worker.Client, error) {
	opt, err := asynq.ParseRedisURI(redisURI)
	if err != nil {
		return nil, err
	}

	c := &client{asynqClient: asynq.NewClient(opt)}

	if c.asynqClient == nil {
		return nil, worker.ErrClientStartFailed
	}

	return c, nil
}

func (c *client) Close() error {
	return c.asynqClient.Close()
}

func (c *client) Submit(ctx context.Context, pattern worker.TaskPattern, payload []byte) error {
	if !pattern.Validate() {
		return worker.ErrTaskPatternInvalid
	}

	task := asynq.NewTask(pattern.String(), payload)
	if _, err := c.asynqClient.EnqueueContext(ctx, task, asynq.Queue(pattern.Queue())); err != nil {
		return worker.ErrSubmitFailed
	}

	return nil
}

func (c *client) SubmitToBatch(ctx context.Context, pattern worker.TaskPattern, payload []byte) error {
	if !pattern.Validate() {
		return worker.ErrTaskPatternInvalid
	}

	task := asynq.NewTask(pattern.String(), payload)
	if _, err := c.asynqClient.EnqueueContext(ctx, task, asynq.Queue(pattern.Queue()), asynq.Group(pattern.String())); err != nil {
		return worker.ErrSubmitFailed
	}

	return nil
}
