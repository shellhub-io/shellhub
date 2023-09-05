package internalclient

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"

	resty "github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
)

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
	httpClient.SetRetryCount(math.MaxInt32)
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok { // if the error is a network error, retry.
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	c := &client{
		host:   apiHost,
		port:   apiPort,
		scheme: apiScheme,
		http:   httpClient,
	}

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

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
	GetDeviceByPublicURLAddress(address string) (*models.Device, error)
}

type client struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *logrus.Logger
	asynq  *asynq.Client
}

func (c *client) ListDevices() ([]models.Device, error) {
	list := []models.Device{}
	_, err := c.http.R().
		SetResult(list).
		Get(buildURL(c, "/api/devices"))

	return list, err
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device
	resp, err := c.http.R().
		SetResult(&device).
		Get(buildURL(c, fmt.Sprintf("/api/devices/%s", uid)))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 400:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}

func (c *client) GetDeviceByPublicURLAddress(address string) (*models.Device, error) {
	httpClient := resty.New()

	var device *models.Device
	resp, err := httpClient.R().
		SetResult(&device).
		Get(buildURL(c, fmt.Sprintf("/internal/devices/public/%s", address)))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 404:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}

func buildURL(c *client, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d%s", c.scheme, c.host, c.port, uri))

	return u.String()
}
