package client

import (
	"context"

	resty "github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/sirupsen/logrus"
)

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
}

type publicAPI interface {
	GetInfo(agentVersion string) (*models.Info, error)
	Endpoints() (*models.Endpoints, error)
	AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthPublicKey(req *models.PublicKeyAuthRequest, token string) (*models.PublicKeyAuthResponse, error)
	NewReverseListener(ctx context.Context, token string) (*revdial.Listener, error)
}

//go:generate mockery --name=Client --filename=client.go
type Client interface {
	commonAPI
	publicAPI
}

type client struct {
	scheme string
	host   string
	port   int
	http   *resty.Client
	logger *logrus.Logger
	// reverser is used to create a reverse listener to Agent from ShellHub's SSH server.
	reverser IReverser
}
