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

type client struct {
	http   *resty.Client
	logger *logrus.Logger
	asynq  *asynq.Client
}

type Client interface {
	deviceAPI
	namespaceAPI
	billingAPI
	sessionAPI
	sshkeyAPI
	firewallAPI
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

// Options wraps injectable values to a new API internal client.
// NOTE(r): This is a workaround to inject the Asynq client to the API internal client, because the [client] structure
// and its properties are privated.
type Options struct {
	Asynq *asynq.Client
}

type Opt func(*Options) error

func NewClient(opts ...Opt) Client {
	httpClient := resty.New()
	httpClient.SetBaseURL("http://api:8080")
	httpClient.SetRetryCount(math.MaxInt32)
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok { // if the error is a network error, retry.
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	c := &client{http: httpClient}

	o := new(Options)
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil
		}
	}

	if o.Asynq != nil {
		c.asynq = o.Asynq
	}

	if c.logger != nil {
		httpClient.SetLogger(&LeveledLogger{c.logger})
	}

	return c
}

// NewClientWithAsynq creates a new API internal client with the Asynq client injected to turn the API internal client
// able to enqueue ping tasks to the Asynq server and late process by API server.
//
// It uses the [NewClient] function to create a new API internal client and injects the Asynq client to it through the
// [Options] structure.
func NewClientWithAsynq(uri string) Client {
	return NewClient(func(o *Options) error {
		// The internal client used by the SSH server needs to be able to enqueue tasks to the Asynq server, due that,
		// we must set the Asynq client to the internal client as a configuration function.
		options, err := asynq.ParseRedisURI(uri)
		if err != nil {
			return err
		}

		client := asynq.NewClient(options)
		if client == nil {
			return errors.New("failed to create Asynq client")
		}

		o.Asynq = client

		return nil
	})
}
