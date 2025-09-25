package internalclient

import (
	"errors"
	"net"
	"net/http"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/worker"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name=Client --filename=internalclient.go
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
	// RetryWaitTime defines the wait time between retries.
	RetryWaitTime time.Duration
	// RetryMaxWaitTime defines the maximum wait time between retries.
	RetryMaxWaitTime time.Duration

	// BaseURL defines the base URL for the API.
	BaseURL string
}

type client struct {
	http   *resty.Client
	logger *log.Logger
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
			// NOTE: Default values can be overridden using the WithConfig option.
			RetryCount:       3,
			RetryWaitTime:    5 * time.Second,
			RetryMaxWaitTime: 20 * time.Second,
			BaseURL:          "http://api:8080",
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

	httpClient.SetBaseURL(c.Config.BaseURL)
	httpClient.SetRetryCount(c.Config.RetryCount)
	httpClient.SetRetryWaitTime(c.Config.RetryWaitTime)
	httpClient.SetRetryMaxWaitTime(c.Config.RetryMaxWaitTime)
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok { // if the error is a network error, retry.
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	httpClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		// NOTE: Add a unique request ID to each request for better traceability.
		r.Header.Set("X-Request-Id", randomString(32))

		log.WithFields(log.Fields{
			"id":      r.Header.Get("X-Request-Id"),
			"attempt": r.Attempt,
			"method":  r.Method,
			"url":     r.URL,
		}).Info("internal client request send")

		return nil
	})

	httpClient.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		log.WithFields(log.Fields{
			"id":      r.Header().Get("X-Request-Id"),
			"attempt": r.Request.Attempt,
			"method":  r.Request.Method,
			"url":     r.Request.URL,
			"status":  r.StatusCode(),
		}).Info("internal client response received")

		return nil
	})

	return c, nil
}

// mustWorker panics if [client.worker] is nil.
func (c *client) mustWorker() {
	if c.worker == nil {
		panic("Client does not have any worker")
	}
}
