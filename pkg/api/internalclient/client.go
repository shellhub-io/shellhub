package internalclient

import (
	"errors"
	"net"
	"net/http"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/sirupsen/logrus"
)

type Client interface {
	deviceAPI
	namespaceAPI
	billingAPI
	sessionAPI
	sshkeyAPI
	firewallAPI
}

// Config holds configuration options for the client.
type Config struct {
	// RetryCount defines how many times the client should retry a request in case of failure.
	RetryCount int
}

type client struct {
	http   *resty.Client
	logger *logrus.Logger
	worker worker.Client

	Config *Config
}

const (
	DeviceUIDHeader = "X-Device-UID"
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
	ErrUnknown          = errors.New("unknown error")
)

func NewClient(opts ...clientOption) (Client, error) {
	httpClient := resty.New()

	c := &client{
		http: httpClient,
		Config: &Config{
			RetryCount: 3,
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.logger != nil {
		httpClient.SetLogger(&LeveledLogger{c.logger})
	}

	httpClient.SetBaseURL("http://api:8080")
	httpClient.SetRetryCount(c.Config.RetryCount)
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok { // if the error is a network error, retry.
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	return c, nil
}

// mustWorker panics if [client.worker] is nil.
func (c *client) mustWorker() {
	if c.worker == nil {
		panic("Client does not have any worker")
	}
}
