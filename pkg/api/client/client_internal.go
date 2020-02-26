// +build internal_api

package client

const (
	apiHost = "api"
	apiPort = 8080
)

type Client interface {
	commonAPI
	internalAPI
}

type internalAPI interface {
	LookupDevice()
}

func (c *client) LookupDevice() {
}
