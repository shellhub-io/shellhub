// +build internal_api

package client

const (
	apiHost   = "api"
	apiPort   = 8080
	apiScheme = "http"
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
