package client

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
)

const (
	apiHost   = "ssh.shellhub.io"
	apiPort   = 80
	apiScheme = "https"
)

type Client interface {
	commonAPI
	publicAPI
}

type publicAPI interface {
	GetInfo(agentVersion string) (*models.Info, error)
	Endpoints() (*models.Endpoints, error)
	AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	NewReverseListener(token string) (*revdial.Listener, error)
	AuthPublicKey(req *models.PublicKeyAuthRequest, token string) (*models.PublicKeyAuthResponse, error)
}

func (c *client) GetInfo(agentVersion string) (*models.Info, error) {
	var info *models.Info

	_, err := c.http.R().
		SetResult(&info).
		Get(buildURL(c, "/info?agent_version="+agentVersion))
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (c *client) AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	var res *models.DeviceAuthResponse
	_, err := c.http.R().
		SetBody(req).
		SetResult(&res).
		Post(buildURL(c, "/api/devices/auth"))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *client) Endpoints() (*models.Endpoints, error) {
	var endpoints *models.Endpoints
	_, err := c.http.R().
		SetResult(&endpoints).
		Get(buildURL(c, "/endpoints"))
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (c *client) NewReverseListener(token string) (*revdial.Listener, error) {
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	url := regexp.MustCompile(`^http`).ReplaceAllString(buildURL(c, "/ssh/connection"), "ws")
	conn, _, err := websocket.DefaultDialer.Dial(url, req.Header)
	if err != nil {
		return nil, err
	}

	listener := revdial.NewListener(wsconnadapter.New(conn),
		func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
			return tunnelDial(ctx, strings.Replace(c.scheme, "http", "ws", 1), c.host, c.port, path)
		},
	)

	return listener, nil
}

func (c *client) AuthPublicKey(req *models.PublicKeyAuthRequest, token string) (*models.PublicKeyAuthResponse, error) {
	var res *models.PublicKeyAuthResponse
	_, err := c.http.R().
		SetBody(req).
		SetResult(&res).
		SetAuthToken(token).
		Post(buildURL(c, "/api/auth/ssh"))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func tunnelDial(ctx context.Context, protocol, address string, port int, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("%s://%s:%d", protocol, address, port), path}, ""), nil)
}
