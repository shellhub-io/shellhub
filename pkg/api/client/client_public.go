// +build !internal_api

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

	_, _, errs := c.http.Get(buildURL(c, "/info?agent_version=" + agentVersion)).EndStruct(&info)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return info, nil
}

func (c *client) AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	var res *models.DeviceAuthResponse
	_, _, errs := c.http.Post(buildURL(c, "/api/devices/auth")).Send(req).EndStruct(&res)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return res, nil
}

func (c *client) Endpoints() (*models.Endpoints, error) {
	var endpoints *models.Endpoints
	_, _, errs := c.http.Get(buildURL(c, "/endpoints")).EndStruct(&endpoints)
	if len(errs) > 0 {
		return nil, errs[0]
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
	_, _, errs := c.http.Post(buildURL(c, "/api/auth/ssh")).Set("Authorization", fmt.Sprintf("Bearer %s", token)).Send(req).EndStruct(&res)

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return res, nil
}

func tunnelDial(ctx context.Context, protocol, address string, port int, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("%s://%s:%d", protocol, address, port), path}, ""), nil)
}
