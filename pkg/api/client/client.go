package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/parnurzeal/gorequest"
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

func NewClient(opts ...Opt) Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{}
	retryClient.RetryMax = math.MaxInt32
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if _, ok := err.(net.Error); ok {
			return true, nil
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	gorequest.DisableTransportSwap = true

	httpClient := gorequest.New()
	httpClient.Client = retryClient.StandardClient()

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
		retryClient.Logger = &LeveledLogger{c.logger}
	}

	return c
}

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
	GetNamespace(tenant string) (*models.Namespace, error)
}

type client struct {
	scheme string
	host   string
	port   int
	http   *gorequest.SuperAgent
	logger *logrus.Logger
}

func (c *client) ListDevices() ([]models.Device, error) {
	list := []models.Device{}
	_, _, errs := c.http.Get(buildURL(c, "/api/devices")).EndStruct(&list)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return list, nil
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device
	resp, _, errs := c.http.Get(buildURL(c, fmt.Sprintf("/api/devices/%s", uid))).EndStruct(&device)
	if len(errs) > 0 {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode {
	case 400:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}

func (c *client) GetNamespace(tenant string) (*models.Namespace, error) {
	var namespace *models.Namespace
	resp, _, errs := c.http.Get(buildURL(c, fmt.Sprintf("/api/namespaces/%s", tenant))).EndStruct(&namespace)
	if len(errs) > 0 {
		return nil, ErrConnectionFailed
	}

	if resp.StatusCode == 400 {
		return nil, ErrNotFound
	} else if resp.StatusCode == 200 {
		return namespace, nil
	}

	return nil, ErrUnknown
}

func buildURL(c *client, uri string) string {
	u, _ := url.Parse(fmt.Sprintf("%s://%s:%d", c.scheme, c.host, c.port))
	u.Path = path.Join(u.Path, uri)

	return u.String()
}
