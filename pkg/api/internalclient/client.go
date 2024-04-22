package internalclient

import (
	"errors"
	"math"
	"net"
	"net/http"

	resty "github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type Client interface {
	Close()

	deviceAPI
	namespaceAPI
	billingAPI
	sessionAPI
	sshkeyAPI
	firewallAPI
}

type Option func(c *client) error

type client struct {
	http   *resty.Client
	logger *logrus.Logger
	asynq  *asynq.Client
}

// Ensures the client implements Client.
var _ Client = (*client)(nil)

const (
	DeviceUIDHeader = "X-Device-UID"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrNotFound         = errors.New("not found")
	ErrUnknown          = errors.New("unknown error")
)

func WithAsynq(redisURI string) Option {
	return func(c *client) error {
		uri, err := asynq.ParseRedisURI(redisURI)
		if err != nil {
			return err
		}

		if c.asynq = asynq.NewClient(uri); c.asynq == nil {
			return errors.New("failed to create Asynq client")
		}

		return nil
	}
}

func New(opts ...Option) Client {
	client := &client{
		http: resty.New(),
	}

	client.http.SetBaseURL("http://api:8080")
	client.http.SetRetryCount(math.MaxInt32)
	client.http.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok { // if the error is a network error, retry.
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil //   TODO: return err
		}
	}

	if client.logger != nil {
		client.http.SetLogger(&LeveledLogger{client.logger})
	}

	return client
}

func (c *client) Close() {
	if c.asynq != nil {
		if err := c.asynq.Close(); err != nil {
			logrus.WithError(err).Error("failed to close internalclient asynq")
		}
	}
}
