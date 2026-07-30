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

type Client interface {
	deviceAPI
	namespaceAPI
	licenseAPI
	sessionAPI
	sshkeyAPI
	firewallAPI
	sshApprovalAPI
	accessPolicyAPI
	sshIdentityAPI
}

type client struct {
	http   *resty.Client
	logger *log.Logger
	worker worker.Client
}

const (
	DeviceUIDHeader = "X-Device-UID"
)

const (
	// apiBaseURL is the API's own listener. Every caller of this package runs
	// inside the server process, so this is always a call to ourselves.
	apiBaseURL = "http://127.0.0.1:8080"

	retryCount       = 3
	retryWaitTime    = 5 * time.Second
	retryMaxWaitTime = 20 * time.Second
)

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("forbidden")
	ErrUnknown          = errors.New("unknown error")
)

func NewClient(opts ...clientOption) (Client, error) {
	httpClient := resty.New()

	c := &client{ //nolint:exhaustruct
		http: httpClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.logger != nil {
		httpClient.SetLogger(&LeveledLogger{c.logger})
	}

	// NOTE: Avoid setting a global base URL on the Resty client; each call sets its own URL.
	httpClient.SetRetryCount(retryCount)
	httpClient.SetRetryWaitTime(retryWaitTime)
	httpClient.SetRetryMaxWaitTime(retryMaxWaitTime)
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

func (c *client) mustWorker() {
	if c.worker == nil {
		panic("Client does not have any worker")
	}
}
