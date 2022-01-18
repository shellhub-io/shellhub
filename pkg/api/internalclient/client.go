package internalclient

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"path"

	resty "github.com/go-resty/resty/v2"
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

type Opt func(*client) error

func NewClient(opts ...Opt) Client {
	httpClient := resty.New()
	httpClient.SetRetryCount(math.MaxInt32)
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	c := &client{
		host:   apiHost,
		port:   apiPort,
		scheme: apiScheme,
		http:   httpClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil
		}
	}

	if c.logger != nil {
		httpClient.SetLogger(&LeveledLogger{c.logger})
	}

	return c
}

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
}

type client struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *logrus.Logger
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

func buildURL(c *client, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", c.scheme, c.host, c.port))
	u.Path = path.Join(u.Path, uri)

	return u.String()
}
