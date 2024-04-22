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

func New() Client {
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

	if c.logger != nil {
		httpClient.SetLogger(&LeveledLogger{c.logger})
	}

	return c
}
