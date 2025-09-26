package internalclient

import "github.com/shellhub-io/shellhub/pkg/worker/asynq"

type clientOption func(c *client) error

func WithAsynqWorker(redisURI string) clientOption { //nolint:revive
	return func(c *client) error {
		asynqClient, err := asynq.NewClient(redisURI)
		if err != nil {
			return err
		}

		c.worker = asynqClient

		return nil
	}
}
